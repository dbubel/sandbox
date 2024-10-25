package lockfreequeue

import (
	"sync/atomic"
	"unsafe"
)

// Node represents a single element in the queue.
type Node struct {
	value interface{}
	next  unsafe.Pointer // points to the next node (for atomic operations)
}

// LockFreeQueue is the lock-free queue structure.
type LockFreeQueue struct {
	head unsafe.Pointer // points to the first node (head)
	tail unsafe.Pointer // points to the last node (tail)
}

// NewLockFreeQueue initializes a new lock-free queue with a sentinel node.
func NewLockFreeQueue() *LockFreeQueue {
	sentinel := unsafe.Pointer(&Node{})
	return &LockFreeQueue{
		head: sentinel,
		tail: sentinel,
	}
}

// Enqueue inserts an element at the end of the queue.
func (q *LockFreeQueue) Enqueue(value interface{}) {
	newNode := unsafe.Pointer(&Node{value: value})

	for {
		tail := atomic.LoadPointer(&q.tail)               // Read the current tail
		next := atomic.LoadPointer(&((*Node)(tail).next)) // Read the next pointer of the tail node

		// If tail has not changed, check if the next node is nil (i.e., tail is really the last node).
		if tail == atomic.LoadPointer(&q.tail) {
			if next == nil {
				// Try to link the new node at the end of the queue.
				if atomic.CompareAndSwapPointer(&((*Node)(tail).next), nil, newNode) {
					// If successful, try to move the tail pointer forward.
					atomic.CompareAndSwapPointer(&q.tail, tail, newNode)
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
// It returns nil if the queue is empty.
func (q *LockFreeQueue) Dequeue() interface{} {
	for {
		head := atomic.LoadPointer(&q.head)               // Read the current head
		tail := atomic.LoadPointer(&q.tail)               // Read the current tail
		next := atomic.LoadPointer(&((*Node)(head).next)) // Read the next pointer of the head node

		// If head has not changed, check if the queue is empty (head == tail and no next node).
		if head == atomic.LoadPointer(&q.head) {
			if head == tail {
				if next == nil {
					// Queue is empty.
					return nil
				}
				// Tail is lagging, move it forward.
				atomic.CompareAndSwapPointer(&q.tail, tail, next)
			} else {
				// Queue is not empty, try to move the head pointer forward.
				if atomic.CompareAndSwapPointer(&q.head, head, next) {
					// Return the value of the old head node.
					return (*Node)(next).value
				}
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
