package lockfreequeue

import (
	"sync/atomic"
	"unsafe"
)

// Node represents a single element in the queue.
type Node[T any] struct {
	value T
	next  unsafe.Pointer // points to the next node (for atomic operations)
}

// LockFreeQueue is the lock-free queue structure.
type LockFreeQueue[T any] struct {
	head unsafe.Pointer // points to the first node (head)
	tail unsafe.Pointer // points to the last node (tail)
	size int64          // atomic counter for queue size
}

// NewLockFreeQueue initializes a new lock-free queue with a sentinel node.
func NewLockFreeQueue[T any]() *LockFreeQueue[T] {
	sentinel := unsafe.Pointer(&Node[T]{})
	return &LockFreeQueue[T]{
		head: sentinel,
		tail: sentinel,
		size: 0,
	}
}

// Enqueue inserts an element at the end of the queue.
func (q *LockFreeQueue[T]) Enqueue(value T) {
	newNode := unsafe.Pointer(&Node[T]{value: value})

	for {
		tail := atomic.LoadPointer(&q.tail)                  // Read the current tail
		next := atomic.LoadPointer(&((*Node[T])(tail).next)) // Read the next pointer of the tail node

		// If tail has not changed, check if the next node is nil (i.e., tail is really the last node).
		if tail == atomic.LoadPointer(&q.tail) {
			if next == nil {
				// Try to link the new node at the end of the queue.
				if atomic.CompareAndSwapPointer(&((*Node[T])(tail).next), nil, newNode) {
					// If successful, try to move the tail pointer forward.
					atomic.CompareAndSwapPointer(&q.tail, tail, newNode)
					atomic.AddInt64(&q.size, 1)
					return
				}
			} else {
				// Tail is not the last node, move the tail pointer forward.
				atomic.CompareAndSwapPointer(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the element at the front of the queue.
// It returns the zero value of T and false if the queue is empty.
func (q *LockFreeQueue[T]) Dequeue() (T, bool) {
	var zeroValue T

	for {
		head := atomic.LoadPointer(&q.head)                  // Read the current head
		tail := atomic.LoadPointer(&q.tail)                  // Read the current tail
		next := atomic.LoadPointer(&((*Node[T])(head).next)) // Read the next pointer of the head node

		// If head has not changed, check if the queue is empty (head == tail and no next node).
		if head == atomic.LoadPointer(&q.head) {
			if head == tail {
				if next == nil {
					// Queue is empty.
					return zeroValue, false
				}
				// Tail is lagging, move it forward.
				atomic.CompareAndSwapPointer(&q.tail, tail, next)
			} else {
				// Queue is not empty, try to move the head pointer forward.
				value := (*Node[T])(next).value
				if atomic.CompareAndSwapPointer(&q.head, head, next) {
					// Decrement the size counter after successful dequeue
					// Use max(0, size-1) to avoid negative size
					for {
						size := atomic.LoadInt64(&q.size)
						if size <= 0 {
							atomic.StoreInt64(&q.size, 0)
							break
						}
						if atomic.CompareAndSwapInt64(&q.size, size, size-1) {
							break
						}
					}
					return value, true
				}
			}
		}
	}
}

// ResetSize recalculates the size by walking through the queue
// This is an expensive operation but can be used to correct the size counter if it gets out of sync
func (q *LockFreeQueue[T]) ResetSize() {
	// Walk the queue and count nodes
	var count int64
	head := atomic.LoadPointer(&q.head)
	current := atomic.LoadPointer(&((*Node[T])(head).next))

	for current != nil {
		count++
		current = atomic.LoadPointer(&((*Node[T])(unsafe.Pointer(current)).next))
	}

	// Atomically update the size
	atomic.StoreInt64(&q.size, count)
}

// Size returns the current size of the queue.
func (q *LockFreeQueue[T]) Size() int {
	// Check if queue is structurally empty
	head := atomic.LoadPointer(&q.head)
	next := atomic.LoadPointer(&((*Node[T])(head).next))

	if next == nil {
		// Queue is definitely empty - reset size counter to 0
		atomic.StoreInt64(&q.size, 0)
		return 0
	}

	// Get current size
	size := atomic.LoadInt64(&q.size)
	if size < 0 {
		// Negative size is impossible - reset to avoid returning negative
		atomic.StoreInt64(&q.size, 0)
		return 0
	}

	return int(size)
}

// IsEmpty returns true if the queue is empty.
func (q *LockFreeQueue[T]) IsEmpty() bool {
	// Always check the structure first
	head := atomic.LoadPointer(&q.head)
	next := atomic.LoadPointer(&((*Node[T])(head).next))

	if next == nil {
		// Queue is structurally empty - ensure size is 0
		atomic.StoreInt64(&q.size, 0)
		return true
	}

	return false
}

// Clear resets the queue to its initial empty state
func (q *LockFreeQueue[T]) Clear() {
	// Create a new sentinel node
	sentinel := unsafe.Pointer(&Node[T]{})

	// Atomically update head and tail to point to the new sentinel
	atomic.StorePointer(&q.head, sentinel)
	atomic.StorePointer(&q.tail, sentinel)

	// Reset the size counter
	atomic.StoreInt64(&q.size, 0)
}

// Peek returns the value at the front of the queue without removing it.
// It returns the zero value of T and false if the queue is empty.
func (q *LockFreeQueue[T]) Peek() (T, bool) {
	var zeroValue T

	for {
		head := atomic.LoadPointer(&q.head)
		tail := atomic.LoadPointer(&q.tail)
		next := atomic.LoadPointer(&((*Node[T])(head).next))

		if head == atomic.LoadPointer(&q.head) {
			if head == tail {
				if next == nil {
					return zeroValue, false
				}
			} else {
				return (*Node[T])(next).value, true
			}
		}
	}
}

// func main() {
// 	queue := NewLockFreeQueue()
//
// 	// Example of concurrent enqueue and dequeue
// 	queue.Enqueue(1)
// 	queue.Enqueue(2)
// 	queue.Enqueue(3)
//
// 	fmt.Println(queue.Dequeue()) // Output: 1
// 	fmt.Println(queue.Dequeue()) // Output: 2
// 	fmt.Println(queue.Dequeue()) // Output: 3
// 	fmt.Println(queue.Dequeue()) // Output: nil (empty queue)
// }
