package bst

import (
	"fmt"
	"testing"
)

func TestBinarySearchTree(t *testing.T) {
	bst := New()
	bst.Insert(1337)
	if bst.root.data != 1337 {
		t.Fail()
		fmt.Println("root data not correct")
		return
	}
	bst.Insert(2000)
	bst.Insert(3000)
	
	bst.InOrder(bst.root)
}
