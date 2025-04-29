# Lock-Free Queue

This package provides a high-performance, thread-safe queue implementation using lock-free algorithms in Go.

## Features

- **Fully concurrent**: Safe for multiple goroutines to access simultaneously
- **Type-safe**: Uses Go generics for compile-time type safety
- **High-performance**: No locks means better performance under high contention
- **Memory efficient**: Only allocates memory when needed
- **Comprehensive API**: Full functionality for queue operations

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	
	"github.com/yourusername/lockfreequeue"
)

func main() {
	// Create a new queue of integers
	queue := lockfreequeue.NewLockFreeQueue[int]()
	
	// Add elements
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)
	
	// Get queue size
	fmt.Println("Queue size:", queue.Size()) // Output: 3
	
	// Check if queue is empty
	fmt.Println("Is empty:", queue.IsEmpty()) // Output: false
	
	// Look at the front element without removing it
	val, ok := queue.Peek()
	if ok {
		fmt.Println("Front element:", val) // Output: 1
	}
	
	// Remove and get elements in FIFO order
	for !queue.IsEmpty() {
		val, ok := queue.Dequeue()
		if ok {
			fmt.Println("Dequeued:", val)
		}
	}
	
	// Output:
	// Dequeued: 1
	// Dequeued: 2
	// Dequeued: 3
}
```

### With Custom Types

```go
type Person struct {
	Name string
	Age  int
}

func main() {
	// Create a queue of Person structs
	queue := lockfreequeue.NewLockFreeQueue[Person]()
	
	// Add people to the queue
	queue.Enqueue(Person{Name: "Alice", Age: 30})
	queue.Enqueue(Person{Name: "Bob", Age: 25})
	
	// Process people in the queue
	for !queue.IsEmpty() {
		person, ok := queue.Dequeue()
		if ok {
			fmt.Printf("Processing: %s, %d years old\n", person.Name, person.Age)
		}
	}
}
```

## API Reference

### Types

- `LockFreeQueue[T any]` - The main queue type, generic over type T

### Functions

- `NewLockFreeQueue[T any]() *LockFreeQueue[T]` - Creates a new empty queue

### Methods

- `Enqueue(value T)` - Adds an element to the end of the queue
- `Dequeue() (T, bool)` - Removes and returns the element at the front of the queue, with success flag
- `Peek() (T, bool)` - Returns the element at the front of the queue without removing it, with success flag
- `Size() int` - Returns the number of elements in the queue
- `IsEmpty() bool` - Returns true if the queue contains no elements

## Implementation Details

This queue is implemented using the Michael-Scott queue algorithm, a lock-free concurrent queue algorithm. It uses atomic operations to ensure thread safety without locks.

Key implementation features:
- Uses a sentinel node to simplify the algorithm
- Employs Compare-And-Swap (CAS) operations for atomic updates
- Maintains separate head and tail pointers for efficient operations
- Handles the "ABA problem" through careful pointer management

## Performance

The lock-free queue significantly outperforms traditional mutex-based queues under high contention. Run the benchmarks to see performance comparisons:

```
go test -bench=. -benchmem
```

## License

[MIT License](LICENSE) 