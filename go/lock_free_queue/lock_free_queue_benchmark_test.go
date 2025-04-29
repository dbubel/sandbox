package lockfreequeue

import (
	"sync"
	"testing"
)

// Benchmark sequential operations
func BenchmarkQueue_Sequential(b *testing.B) {
	queue := NewLockFreeQueue[int]()

	b.Run("Enqueue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			queue.Enqueue(i)
		}
	})

	// Reset the queue
	queue = NewLockFreeQueue[int]()
	// Fill queue for dequeue benchmarks
	for i := 0; i < 1000000; i++ {
		queue.Enqueue(i)
	}

	b.Run("Dequeue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%1000000 == 0 {
				// Refill the queue if it might be empty
				for j := 0; j < 1000000; j++ {
					queue.Enqueue(j)
				}
			}
			queue.Dequeue()
		}
	})

	b.Run("Peek", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			queue.Peek()
		}
	})
}

// Benchmark concurrent operations
func BenchmarkQueue_Concurrent(b *testing.B) {
	// For a fair comparison, scale down b.N to account for multiple goroutines
	operationsPerGoroutine := b.N / 8
	if operationsPerGoroutine < 1 {
		operationsPerGoroutine = 1
	}

	b.Run("EnqueueOnly", func(b *testing.B) {
		b.ResetTimer()
		queue := NewLockFreeQueue[int]()
		var wg sync.WaitGroup

		// Launch 8 goroutines to enqueue concurrently
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					queue.Enqueue(j)
				}
			}(i)
		}

		wg.Wait()
	})

	b.Run("DequeueOnly", func(b *testing.B) {
		queue := NewLockFreeQueue[int]()
		// Fill queue for dequeue benchmarks
		for i := 0; i < operationsPerGoroutine*10; i++ {
			queue.Enqueue(i)
		}

		b.ResetTimer()
		var wg sync.WaitGroup

		// Launch 8 goroutines to dequeue concurrently
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					queue.Dequeue()
				}
			}()
		}

		wg.Wait()
	})

	b.Run("MixedOperations", func(b *testing.B) {
		queue := NewLockFreeQueue[int]()
		var wg sync.WaitGroup

		// Prefill the queue with some items
		for i := 0; i < 1000; i++ {
			queue.Enqueue(i)
		}

		b.ResetTimer()

		// Launch 4 goroutines to enqueue concurrently
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					queue.Enqueue(j)
				}
			}(i)
		}

		// Launch 4 goroutines to dequeue concurrently
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					queue.Dequeue()
				}
			}()
		}

		wg.Wait()
	})
}

// Benchmark against a mutex-based queue implementation for comparison
func BenchmarkQueue_ComparisonWithMutexQueue(b *testing.B) {
	// Simple mutex-based queue implementation for comparison
	type MutexQueue struct {
		data []int
		mu   sync.Mutex
	}

	newMutexQueue := func() *MutexQueue {
		return &MutexQueue{data: make([]int, 0, 1000)}
	}

	enqueue := func(q *MutexQueue, val int) {
		q.mu.Lock()
		defer q.mu.Unlock()
		q.data = append(q.data, val)
	}

	dequeue := func(q *MutexQueue) (int, bool) {
		q.mu.Lock()
		defer q.mu.Unlock()
		if len(q.data) == 0 {
			return 0, false
		}
		val := q.data[0]
		q.data = q.data[1:]
		return val, true
	}

	// Benchmark lock-free queue
	b.Run("LockFree-ConcurrentMixed", func(b *testing.B) {
		queue := NewLockFreeQueue[int]()
		operationsPerGoroutine := b.N / 8
		if operationsPerGoroutine < 1 {
			operationsPerGoroutine = 1
		}

		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					queue.Enqueue(j)
				}
			}()
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					queue.Dequeue()
				}
			}()
		}
		wg.Wait()
	})

	// Benchmark mutex-based queue
	b.Run("Mutex-ConcurrentMixed", func(b *testing.B) {
		queue := newMutexQueue()
		operationsPerGoroutine := b.N / 8
		if operationsPerGoroutine < 1 {
			operationsPerGoroutine = 1
		}

		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					enqueue(queue, j)
				}
			}()
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					dequeue(queue)
				}
			}()
		}
		wg.Wait()
	})
}
