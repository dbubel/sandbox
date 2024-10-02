package bst

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestBinarySearchTree(t *testing.T) {
	bst := New()
	bst.InsertIter(1337)
	if bst.root.data != 1337 {
		t.Fail()
		fmt.Println("root data not correct")
		return
	}
	bst.InsertIter(2000)
	bst.InsertIter(3000)
	bst.InsertIter(2500)

	bst.InOrder(bst.root)
}

func TestBinarySearchTreeRec(t *testing.T) {
	bst := New()
	bst.InsertRec(1337)
	if bst.root.data != 1337 {
		t.Fail()
		fmt.Println("root data not correct")
		return
	}
	bst.InsertRec(2000)
	bst.InsertRec(2500)
	bst.InsertRec(3000)

	bst.InOrder(bst.root)
}

func BenchmarkInsertRec(b *testing.B) {
	bst := New()

	for i := 0; i < b.N; i++ {
		bst.InsertRec(rand.Intn(math.MaxInt)) // Using random numbers for insertion
	}
}

func BenchmarkInsertIter(b *testing.B) {
	bst := New()

	for i := 0; i < b.N; i++ {
		bst.InsertIter(rand.Intn(math.MaxInt)) // Using random numbers for insertion
	}
}
