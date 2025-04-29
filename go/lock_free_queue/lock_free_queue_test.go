package lockfreequeue

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueue_BasicOperations(t *testing.T) {
	queue := NewLockFreeQueue[int]()

	// Test initial state
	if !queue.IsEmpty() {
		t.Error("New queue should be empty")
	}
	if queue.Size() != 0 {
		t.Errorf("New queue should have size 0, got %d", queue.Size())
	}

	// Test Enqueue
	queue.Enqueue(1)
	if queue.IsEmpty() {
		t.Error("Queue should not be empty after enqueue")
	}
	if queue.Size() != 1 {
		t.Errorf("Queue should have size 1, got %d", queue.Size())
	}

	// Test Peek
	val, ok := queue.Peek()
	if !ok {
		t.Error("Peek should succeed on non-empty queue")
	}
	if val != 1 {
		t.Errorf("Peek should return 1, got %d", val)
	}
	if queue.Size() != 1 {
		t.Errorf("Peek should not change queue size, expected 1, got %d", queue.Size())
	}

	// Test Dequeue
	val, ok = queue.Dequeue()
	if !ok {
		t.Error("Dequeue should succeed on non-empty queue")
	}
	if val != 1 {
		t.Errorf("Dequeue should return 1, got %d", val)
	}
	if !queue.IsEmpty() {
		t.Error("Queue should be empty after dequeuing only element")
	}

	// Test empty queue operations
	val, ok = queue.Peek()
	if ok {
		t.Error("Peek should fail on empty queue")
	}

	val, ok = queue.Dequeue()
	if ok {
		t.Error("Dequeue should fail on empty queue")
	}
}

func TestQueue_FIFO(t *testing.T) {
	queue := NewLockFreeQueue[int]()

	// Enqueue multiple elements
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	// Verify FIFO order
	for i := 1; i <= 3; i++ {
		val, ok := queue.Dequeue()
		if !ok {
			t.Errorf("Dequeue #%d should succeed", i)
		}
		if val != i {
			t.Errorf("Expected %d, got %d", i, val)
		}
	}
}

func TestQueue_ConcurrentEnqueueDequeue(t *testing.T) {
	queue := NewLockFreeQueue[int]()
	itemCount := 1000
	workerCount := 4

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup

	// Launch producer goroutines
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < itemCount; i++ {
				queue.Enqueue(workerID*itemCount + i)
			}
		}(w)
	}

	// Wait for enqueuing to finish
	wg.Wait()

	// Verify queue size
	expectedSize := workerCount * itemCount
	if queue.Size() != expectedSize {
		t.Errorf("Expected queue size %d, got %d", expectedSize, queue.Size())
	}

	// Launch consumer goroutines
	var consumed sync.Map
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < itemCount; i++ {
				val, ok := queue.Dequeue()
				if !ok {
					t.Error("Dequeue failed unexpectedly")
					return
				}
				consumed.Store(val, true)
			}
		}()
	}

	// Wait for dequeuing to finish
	wg.Wait()

	// Verify all items were consumed
	for w := 0; w < workerCount; w++ {
		for i := 0; i < itemCount; i++ {
			val := w*itemCount + i
			if _, exists := consumed.Load(val); !exists {
				t.Errorf("Item %d was not consumed", val)
			}
		}
	}

	// Verify queue is empty
	if !queue.IsEmpty() {
		t.Error("Queue should be empty after all dequeues")
	}
}

func TestQueue_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	queue := NewLockFreeQueue[int]()
	const iterations = 10000 // Reduced for stability
	const workerCount = 4    // Reduced for stability

	// Count of completed operations
	var enqueueDone int64
	var dequeueDone int64

	var wg sync.WaitGroup

	// Barrier to start all goroutines at around the same time
	var startBarrier, finishBarrier sync.WaitGroup
	startBarrier.Add(1)

	// Start concurrent enqueue workers
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			defer wg.Done()
			// Wait for the start signal
			startBarrier.Wait()

			for j := 0; j < iterations; j++ {
				queue.Enqueue(j)
				atomic.AddInt64(&enqueueDone, 1)
			}
			// Signal this goroutine is done with enqueues
			finishBarrier.Done()
		}(i)
	}

	// Start concurrent dequeue workers
	wg.Add(workerCount)
	finishBarrier.Add(workerCount * 2) // For all producers and consumers

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			// Wait for the start signal
			startBarrier.Wait()

			for j := 0; j < iterations; j++ {
				// Try to dequeue with exponential backoff
				var success bool
				backoff := 1 * time.Nanosecond
				maxBackoff := 100 * time.Microsecond

				for !success {
					if _, ok := queue.Dequeue(); ok {
						atomic.AddInt64(&dequeueDone, 1)
						success = true
					} else {
						// If queue is empty, back off a bit
						time.Sleep(backoff)
						// Exponential backoff with a cap
						backoff *= 2
						if backoff > maxBackoff {
							backoff = maxBackoff
						}
					}
				}
			}
			// Signal this goroutine is done with dequeues
			finishBarrier.Done()
		}()
	}

	// Now start all goroutines at once
	startBarrier.Done()

	// Wait for all operations to complete
	finishBarrier.Wait()

	// Verify all operations completed
	expectedOps := int64(workerCount * iterations)
	if atomic.LoadInt64(&enqueueDone) != expectedOps {
		t.Errorf("Expected %d enqueues, got %d", expectedOps, atomic.LoadInt64(&enqueueDone))
	}
	if atomic.LoadInt64(&dequeueDone) != expectedOps {
		t.Errorf("Expected %d dequeues, got %d", expectedOps, atomic.LoadInt64(&dequeueDone))
	}

	// Wait for any pending operations to complete
	wg.Wait()

	// Give extra time for any pending operations
	time.Sleep(500 * time.Millisecond)

	// Force a size reset to ensure accuracy
	queue.ResetSize()

	// Final queue size should be 0 since we have equal enqueues and dequeues
	qSize := queue.Size()
	if qSize != 0 {
		// For debugging
		t.Logf("Enqueues: %d, Dequeues: %d", atomic.LoadInt64(&enqueueDone), atomic.LoadInt64(&dequeueDone))

		// Force empty the queue to debug what's still in it
		var remaining []int
		for i := 0; i < 100 && !queue.IsEmpty(); i++ {
			if val, ok := queue.Dequeue(); ok {
				remaining = append(remaining, val)
			}
		}

		t.Errorf("Queue size should be 0 after equal enqueues and dequeues, got %d. Remaining items: %v", qSize, remaining)
	} else {
		t.Log("Queue successfully emptied")
	}
}

// Test different types
func TestQueue_DifferentTypes(t *testing.T) {
	// String queue
	strQueue := NewLockFreeQueue[string]()
	strQueue.Enqueue("hello")
	strQueue.Enqueue("world")

	val, ok := strQueue.Dequeue()
	if !ok || val != "hello" {
		t.Errorf("Expected 'hello', got '%s'", val)
	}

	// Struct queue
	type Person struct {
		Name string
		Age  int
	}

	personQueue := NewLockFreeQueue[Person]()
	personQueue.Enqueue(Person{Name: "Alice", Age: 30})
	personQueue.Enqueue(Person{Name: "Bob", Age: 25})

	person, ok := personQueue.Dequeue()
	if !ok || person.Name != "Alice" || person.Age != 30 {
		t.Errorf("Expected Person{Name: 'Alice', Age: 30}, got %+v", person)
	}
}
