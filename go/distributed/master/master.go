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
	defaultPort = 50051
)

// Slave represents a connected slave server
type Slave struct {
	ID        int32
	Address   string
	Port      int32
	Status    string
	Load      float64
	LastSeen  time.Time
	Client    pb.DistributedSystemClient
	Available bool
}

// Master represents the master server
type Master struct {
	pb.UnimplementedDistributedSystemServer
	slaves       map[int32]*Slave
	slavesMutex  sync.RWMutex
	tasks        map[string]*utils.Task
	tasksMutex   sync.RWMutex
	results      map[string]*utils.TaskResult
	resultsMutex sync.RWMutex
}

// RegisterSlave handles slave registration
func (m *Master) RegisterSlave(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	m.slavesMutex.Lock()
	defer m.slavesMutex.Unlock()

	slaveID := req.SlaveId
	address := req.Address
	port := req.Port

	log.Printf("Slave %d at %s:%d is registering", slaveID, address, port)

	// Check if slave already exists
	if _, exists := m.slaves[slaveID]; exists {
		return &pb.RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("Slave with ID %d already registered", slaveID),
		}, nil
	}

	// Create client connection to the slave
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", address, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect to slave: %v", err),
		}, nil
	}

	client := pb.NewDistributedSystemClient(conn)

	// Add the new slave
	m.slaves[slaveID] = &Slave{
		ID:        slaveID,
		Address:   address,
		Port:      port,
		Status:    "active",
		Load:      0.0,
		LastSeen:  time.Now(),
		Client:    client,
		Available: true,
	}

	log.Printf("Slave %d successfully registered", slaveID)
	return &pb.RegisterResponse{
		Success: true,
		Message: "Successfully registered",
	}, nil
}

// Heartbeat handles heartbeat requests
func (m *Master) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	// In a real implementation, this would be called by slaves
	// For this example, we'll just return a dummy response
	return &pb.HeartbeatResponse{
		SlaveId: -1, // Master's ID
		Status:  "active",
		Load:    0.0,
	}, nil
}

// AssignTask handles task assignments
func (m *Master) AssignTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	// This would typically be called by the master to slaves
	// For this example, we'll just return a dummy response
	return &pb.TaskResponse{
		TaskId:   req.TaskId,
		Accepted: true,
		Message:  "Task accepted",
	}, nil
}

// CompleteTask handles task completion reports
func (m *Master) CompleteTask(ctx context.Context, req *pb.TaskResult) (*pb.TaskAck, error) {
	taskID := req.TaskId

	log.Printf("Received task completion for task %s. Success: %v", taskID, req.Success)

	m.tasksMutex.Lock()
	delete(m.tasks, taskID)
	m.tasksMutex.Unlock()

	m.resultsMutex.Lock()
	m.results[taskID] = &utils.TaskResult{
		TaskID:         taskID,
		Success:        req.Success,
		Result:         req.Result,
		ErrorMessage:   req.ErrorMessage,
		CompletionTime: time.Now(),
	}
	m.resultsMutex.Unlock()

	// Mark the slave as available again
	if slaveID := findSlaveForTask(taskID); slaveID > 0 {
		m.slavesMutex.Lock()
		if slave, exists := m.slaves[slaveID]; exists {
			slave.Available = true
			slave.Load -= 0.1 // Decrease the load
			if slave.Load < 0 {
				slave.Load = 0
			}
		}
		m.slavesMutex.Unlock()
	}

	return &pb.TaskAck{
		TaskId:   taskID,
		Received: true,
	}, nil
}

// findSlaveForTask is a helper to find which slave was assigned a task
// In a real system, you'd have proper bookkeeping
func findSlaveForTask(taskID string) int32 {
	// This is a placeholder - in a real system, you'd track which slave has which task
	return -1
}

// startHeartbeatCheck periodically checks slave heartbeats
func (m *Master) startHeartbeatCheck() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			m.checkSlaveHeartbeats()
		}
	}()
}

// checkSlaveHeartbeats checks all slaves are still alive
func (m *Master) checkSlaveHeartbeats() {
	m.slavesMutex.RLock()
	slaves := make([]*Slave, 0, len(m.slaves))
	for _, slave := range m.slaves {
		slaves = append(slaves, slave)
	}
	m.slavesMutex.RUnlock()

	for _, slave := range slaves {
		go func(s *Slave) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			resp, err := s.Client.Heartbeat(ctx, &pb.HeartbeatRequest{
				Timestamp: time.Now().Unix(),
			})

			m.slavesMutex.Lock()
			defer m.slavesMutex.Unlock()

			if err != nil {
				log.Printf("Failed to get heartbeat from slave %d: %v", s.ID, err)
				s.Status = "unreachable"
			} else {
				s.Status = resp.Status
				s.Load = resp.Load
				s.LastSeen = time.Now()
			}
		}(slave)
	}
}

// dispatchTasks sends tasks to available slaves
func (m *Master) dispatchTasks() {
	for {
		time.Sleep(1 * time.Second)

		m.slavesMutex.RLock()
		if len(m.slaves) == 0 {
			m.slavesMutex.RUnlock()
			continue
		}

		// Find available slaves
		availableSlaves := make([]*Slave, 0)
		for _, slave := range m.slaves {
			if slave.Available && slave.Status == "active" {
				availableSlaves = append(availableSlaves, slave)
			}
		}
		m.slavesMutex.RUnlock()

		if len(availableSlaves) == 0 {
			continue
		}

		// Create a new task
		taskID := utils.GenerateRandomID("task")
		taskTypes := []string{"fast", "medium", "slow"}
		taskType := taskTypes[rand.Intn(len(taskTypes))]

		task := &utils.Task{
			ID:       taskID,
			Type:     taskType,
			Payload:  []byte(fmt.Sprintf("Task data for %s", taskID)),
			Deadline: time.Now().Add(30 * time.Second),
		}

		m.tasksMutex.Lock()
		m.tasks[taskID] = task
		m.tasksMutex.Unlock()

		// Pick a random available slave
		slave := availableSlaves[rand.Intn(len(availableSlaves))]

		m.slavesMutex.Lock()
		slave.Available = false
		slave.Load += 0.1 // Increase the load
		m.slavesMutex.Unlock()

		// Assign the task
		go func(s *Slave, t *utils.Task) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			deadline := t.Deadline.Unix()
			resp, err := s.Client.AssignTask(ctx, &pb.TaskRequest{
				TaskId:   t.ID,
				TaskType: t.Type,
				Payload:  t.Payload,
				Deadline: deadline,
			})

			if err != nil {
				log.Printf("Failed to assign task %s to slave %d: %v", t.ID, s.ID, err)

				m.slavesMutex.Lock()
				s.Available = true
				m.slavesMutex.Unlock()

				m.tasksMutex.Lock()
				delete(m.tasks, t.ID)
				m.tasksMutex.Unlock()
				return
			}

			if !resp.Accepted {
				log.Printf("Slave %d rejected task %s: %s", s.ID, t.ID, resp.Message)

				m.slavesMutex.Lock()
				s.Available = true
				m.slavesMutex.Unlock()

				m.tasksMutex.Lock()
				delete(m.tasks, t.ID)
				m.tasksMutex.Unlock()
				return
			}

			log.Printf("Task %s assigned to slave %d", t.ID, s.ID)
		}(slave, task)
	}
}

// startServer starts the gRPC server
func startServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	master := &Master{
		slaves:  make(map[int32]*Slave),
		tasks:   make(map[string]*utils.Task),
		results: make(map[string]*utils.TaskResult),
	}

	// Start the heartbeat checker
	master.startHeartbeatCheck()

	// Start task dispatcher
	go master.dispatchTasks()

	grpcServer := grpc.NewServer()
	pb.RegisterDistributedSystemServer(grpcServer, master)

	log.Printf("Master server started on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	port := flag.Int("port", defaultPort, "The server port")
	flag.Parse()

	startServer(*port)
}
