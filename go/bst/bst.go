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

func (bst *BinarySearchtree) InsertRec(v int) {
	bst.root = insert(bst.root, v)
}

func insert(n *Node, v int) *Node {
	if n == nil {
		return &Node{data: v}
	}
	if v > n.data {
		n.right = insert(n.right, v)
	} else if v < n.data {
		n.left = insert(n.left, v)
	}
	return n
}

func (bst *BinarySearchtree) InsertIter(v int) {
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

		if v < current.data {
			if current.left == nil {
				current.left = &Node{data: v}
				return
			}
			current = current.left
			continue
		}
		// case that v already exists in the tree
		return
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

func (bst *BinarySearchtree) del() {
	 
}
