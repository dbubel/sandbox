package bst

import "fmt"

type BinarySearchtree struct {
	root *Node
}
type Node struct {
	data  int
	left  *Node
	right *Node
}

func New() *BinarySearchtree {
	return &BinarySearchtree{}
}

func (bst *BinarySearchtree) Insert(v int) {
	// case for empty tree
	if bst.root == nil {
		bst.root = &Node{data: v}
		return
	}
	current := bst.root
	for {
		if v > current.data {
			if current.right == nil {
				current.right = &Node{data: v}
				return
			}
			current = current.right
			continue
		}
	}
}

// InOrder prints the tree in in-order traversal (sorted order).
func (bst *BinarySearchtree) InOrder(node *Node) {
	if node != nil {
		bst.InOrder(node.left)
		fmt.Printf("%d ", node.data)
		bst.InOrder(node.right)
	}
}
