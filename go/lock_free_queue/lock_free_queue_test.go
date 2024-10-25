package lockfreequeue

import "testing"

func BenchmarkLockFreeQueueEnqueue(b *testing.B) {
	q := NewLockFreeQueue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
	}
}
