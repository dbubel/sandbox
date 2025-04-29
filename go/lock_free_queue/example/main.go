package main

import (
	"fmt"
	"sync"
	"time"

	"lockfreequeue"
)

func main() {
	fmt.Println("Lock-Free Queue Example")

	// Basic usage example
	fmt.Println("\n=== Basic Usage ===")
	basicExample()

	// Concurrent usage example
	fmt.Println("\n=== Concurrent Usage ===")
	concurrentExample()

	// Custom type example
	fmt.Println("\n=== Custom Type Example ===")
	customTypeExample()
}

func basicExample() {
	// Create a new queue of integers
	queue := lockfreequeue.NewLockFreeQueue[int]()

	// Add elements
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	// Get queue size
	fmt.Println("Queue size:", queue.Size())

	// Check if queue is empty
	fmt.Println("Is queue empty:", queue.IsEmpty())

	// Look at the front element without removing it
	val, ok := queue.Peek()
	if ok {
		fmt.Println("Front element:", val)
	}

	// Remove and get elements in FIFO order
	for !queue.IsEmpty() {
		val, ok := queue.Dequeue()
		if ok {
			fmt.Println("Dequeued:", val) 
		}
	}

	// Now the queue is empty
	fmt.Println("Queue is now empty:", queue.IsEmpty())
}

func concurrentExample() {
	queue := lockfreequeue.NewLockFreeQueue[int]()
	var wg sync.WaitGroup

	// Number of items to enqueue/dequeue
	const itemCount = 100
	// Number of concurrent producers and consumers
	const workerCount = 5

	// Create a channel to collect dequeued items for verification
	results := make(chan int, itemCount*workerCount)

	// Start producer goroutines
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < itemCount; i++ {
				// Create unique values based on worker ID and item index
				value := workerID*itemCount + i
				queue.Enqueue(value)
				// Small sleep to interleave operations
				time.Sleep(time.Microsecond)
			}
		}(w)
	}

	// Start consumer goroutines
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < itemCount; i++ {
				for {
					// Try to dequeue an item
					if val, ok := queue.Dequeue(); ok {
						results <- val
						break
					}
					// If queue is empty, wait a bit and try again
					time.Sleep(time.Microsecond)
				}
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)

	// Count the results
	count := 0
	for range results {
		count++
	}

	fmt.Println("Total items processed:", count)
	fmt.Println("Expected items:", itemCount*workerCount)
	fmt.Println("Queue is empty:", queue.IsEmpty())
	fmt.Println("Final queue size:", queue.Size())
}

// Example with a custom type
type Task struct {
	ID       int
	Name     string
	Priority int
}

func (t Task) String() string {
	return fmt.Sprintf("Task{ID: %d, Name: '%s', Priority: %d}", t.ID, t.Name, t.Priority)
}

func customTypeExample() {
	// Create a queue of tasks
	taskQueue := lockfreequeue.NewLockFreeQueue[Task]()

	// Add tasks to the queue
	taskQueue.Enqueue(Task{ID: 1, Name: "Complete project", Priority: 3})
	taskQueue.Enqueue(Task{ID: 2, Name: "Write tests", Priority: 2})
	taskQueue.Enqueue(Task{ID: 3, Name: "Fix bugs", Priority: 1})

	fmt.Println("Task queue size:", taskQueue.Size())

	// Process tasks in order
	fmt.Println("Processing tasks:")
	for !taskQueue.IsEmpty() {
		task, ok := taskQueue.Dequeue()
		if ok {
			fmt.Println("-", task)
		}
	}
}
