package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/yourusername/distributed/pkg/utils"
	pb "github.com/yourusername/distributed/proto"
)

const (
	defaultID         = 1
	defaultPort       = 5001
	defaultMasterAddr = "localhost:50051"
)

// ActiveTask represents a task currently being processed
type ActiveTask struct {
	TaskID     string
	StartTime  time.Time
	Deadline   time.Time
	TaskType   string
	Processing bool
}

// Slave represents the slave server
type Slave struct {
	pb.UnimplementedDistributedSystemServer
	id            int32
	address       string
	port          int32
	masterAddress string
	masterClient  pb.DistributedSystemClient
	status        string
	load          float64
	activeTasks   map[string]*ActiveTask
	tasksMutex    sync.RWMutex
	maxLoad       float64
}

// Heartbeat handles heartbeat requests from master
func (s *Slave) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	return &pb.HeartbeatResponse{
		SlaveId: s.id,
		Status:  s.status,
		Load:    s.load,
	}, nil
}

// AssignTask handles task assignment from master
func (s *Slave) AssignTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	taskID := req.TaskId
	taskType := req.TaskType
	deadline := time.Unix(req.Deadline, 0)

	log.Printf("Received task assignment: %s (type: %s, deadline: %v)", taskID, taskType, deadline)

	// Check if we can accept more tasks
	s.tasksMutex.RLock()
	currentLoad := s.load
	s.tasksMutex.RUnlock()

	if currentLoad >= s.maxLoad {
		log.Printf("Rejecting task %s: load too high (%.2f/%.2f)", taskID, currentLoad, s.maxLoad)
		return &pb.TaskResponse{
			TaskId:   taskID,
			Accepted: false,
			Message:  fmt.Sprintf("Load too high: %.2f/%.2f", currentLoad, s.maxLoad),
		}, nil
	}

	// Accept and process the task
	task := &ActiveTask{
		TaskID:     taskID,
		StartTime:  time.Now(),
		Deadline:   deadline,
		TaskType:   taskType,
		Processing: true,
	}

	s.tasksMutex.Lock()
	s.activeTasks[taskID] = task
	s.load += 0.1 // Increase the load
	s.tasksMutex.Unlock()

	// Process the task in a goroutine
	go s.processTask(task, req.Payload)

	return &pb.TaskResponse{
		TaskId:   taskID,
		Accepted: true,
		Message:  "Task accepted for processing",
	}, nil
}

// RegisterSlave handles slave registration (not used by slave)
func (s *Slave) RegisterSlave(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// This would typically be called by slaves to master
	// For this example, we'll just return a dummy response
	return &pb.RegisterResponse{
		Success: false,
		Message: "Slave cannot accept registration requests",
	}, nil
}

// CompleteTask handles task completion (not used by slave)
func (s *Slave) CompleteTask(ctx context.Context, req *pb.TaskResult) (*pb.TaskAck, error) {
	// This would typically be called by slaves to master
	// For this example, we'll just return a dummy response
	return &pb.TaskAck{
		TaskId:   req.TaskId,
		Received: false,
	}, nil
}

// processTask handles the actual processing of a task
func (s *Slave) processTask(task *ActiveTask, payload []byte) {
	log.Printf("Processing task %s of type %s", task.TaskID, task.TaskType)

	// Simulate work
	result, err := utils.SimulateWork(task.TaskType)
	success := err == nil
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	// Report task completion to master
	s.reportTaskCompletion(task.TaskID, success, result, errorMessage)

	// Update local state
	s.tasksMutex.Lock()
	delete(s.activeTasks, task.TaskID)
	s.load -= 0.1 // Decrease the load
	if s.load < 0 {
		s.load = 0
	}
	s.tasksMutex.Unlock()
}

// reportTaskCompletion sends task results back to master
func (s *Slave) reportTaskCompletion(taskID string, success bool, result []byte, errorMessage string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.masterClient.CompleteTask(ctx, &pb.TaskResult{
		TaskId:         taskID,
		Success:        success,
		Result:         result,
		ErrorMessage:   errorMessage,
		CompletionTime: time.Now().Unix(),
	})

	if err != nil {
		log.Printf("Failed to report task completion to master: %v", err)
	} else {
		log.Printf("Successfully reported completion of task %s", taskID)
	}
}

// registerWithMaster attempts to register this slave with the master
func (s *Slave) registerWithMaster() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.masterClient.RegisterSlave(ctx, &pb.RegisterRequest{
		SlaveId: s.id,
		Address: s.address,
		Port:    s.port,
	})

	if err != nil {
		return fmt.Errorf("failed to register with master: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("master rejected registration: %s", resp.Message)
	}

	log.Printf("Successfully registered with master: %s", resp.Message)
	return nil
}

// startServer starts the gRPC server for this slave
func startServer(id int32, port int32, masterAddress string) {
	// Prepare the slave object
	slave := &Slave{
		id:            id,
		address:       "localhost", // In a real system, this would be determined dynamically
		port:          port,
		masterAddress: masterAddress,
		status:        "starting",
		load:          0.0,
		activeTasks:   make(map[string]*ActiveTask),
		maxLoad:       1.0, // Maximum load this slave can handle
	}

	// Connect to the master
	conn, err := grpc.Dial(masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to master: %v", err)
	}
	defer conn.Close()

	slave.masterClient = pb.NewDistributedSystemClient(conn)
	slave.status = "active"

	// Start the gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDistributedSystemServer(grpcServer, slave)

	// Register with master before serving
	if err := slave.registerWithMaster(); err != nil {
		log.Fatalf("Failed to register with master: %v", err)
	}

	log.Printf("Slave %d started on port %d and registered with master at %s", id, port, masterAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	id := flag.Int("id", defaultID, "The ID of this slave")
	port := flag.Int("port", defaultPort, "The server port for this slave")
	masterAddr := flag.String("master", defaultMasterAddr, "The master server address")
	flag.Parse()

	startServer(int32(*id), int32(*port), *masterAddr)
}
