// Package naive provides a basic implementation of an AVL (Adelson-Velsky and Landis) tree.
// It includes functionalities to insert new nodes and retrieve keys in sorted order.
package naive

import (
	"sync"
	"sync/atomic"
)

// AVL represents an AVL tree with a pointer to the root node.
type AVL struct {
	root *Node // Root points to the root node of the AVL tree.
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewAVL() *AVL {
	return &AVL{}
}

// Insert adds a new node with the given key and data to the AVL tree.
// It ensures that the tree remains balanced after the insertion.
func (tree *AVL) Insert(key string, data *Data) {
	tree.root = tree.root.insertAVL(key, data)
}

// GetKeys retrieves all keys from the AVL tree in sorted descending order.
func (tree *AVL) GetKeys() []string {
	keys := tree.root.inorderDesc([]string{})
	return keys
}

// rotateRight performs a right rotation on the node.
// This is used to rebalance the tree in case of a left-heavy imbalance.
func (A * Node) rotateRight() *Node {
	/*
			 A
			/ \
		   B   C
		  / \
		 D   E
		/
 	  F
	*/

	B := A.left
	tmp := B.right

	B.right = A
	A.left = tmp

	// Fix the Node heights from child to parent
	A.height.Store(
		1 + max(
			A.left.getHeight(),
			A.right.getHeight(),
		),
	) 
	B.height.Store(
		1 + max(
			B.left.getHeight(),
			B.right.getHeight(),
		),
	) 

	return B
}

// rotateLeft performs a left rotation on the node.
// This is used to rebalance the tree in case of a right-heavy imbalance.
func (A *Node) rotateLeft() *Node {
	/*
		 A
		/ \
	   B   C
		  / \
		 F   D
			  \
			   E
	*/

	C := A.right
	tmp := C.left

	C.left = A
	A.right = tmp

	// Fix the Node heights from child to parent
	A.height.Store(
		1 + max(
			A.left.getHeight(),
			A.right.getHeight(),
		),
	) 
	C.height.Store(
		1 + max(
			C.left.getHeight(),
			C.right.getHeight(),
		),
	) 

	return C
}

// insertAVL adds a new node with the given key and data to the tree rooted at the current node.
// It ensures the AVL tree properties are maintained by performing necessary rotations.
func (root *Node) insertAVL(key string, data *Data) *Node  {
	
	if root == nil {
		return &Node{
			key: key,
			lock: sync.Mutex{},
			data: data,
			height: atomic.Int32{}, // Leaves have a height of 0
		}
	}

	if root.key == key {
		// Nothing further to do so we can safely return
		return root
	}

	if root.key > key{
		root.right = root.right.insertAVL(key, data)
	}

	if root.key < key {
		root.left = root.left.insertAVL(key, data)
	}

	// TODO Optimize
	// Update height
	root.height.Store(
		1 + max(
			root.left.getHeight(),
			root.right.getHeight(),
		),
	) 
	balanceFactor := root.getBalanceFactor()

	// Conditions under which balanceFactor itself would not lead to a balancing operation:
	//  -1 =< bf <= 1
	if -1 <= balanceFactor && balanceFactor <= 1 {
		return root
	}

	// TODO: Optimize this
	if balanceFactor == 2 {
		// Left-Left ==> Right Rotation
		if root.left.getBalanceFactor() == 1 || root.left.getBalanceFactor() == 0 {
			root = root.rotateRight()
		}

		// Left-right ==> Left Rotation followed by Right Rotation
		if root.left.getBalanceFactor() == -1 {
			root.left = root.left.rotateLeft()
			root = root.rotateRight()
		}
	}

	if balanceFactor == -2 {
		// Right-Right ==> Left Rotation
		if root.right.getBalanceFactor() == -1 || root.right.getBalanceFactor() == 0 {
			root = root.rotateLeft()
		}

		// Right-Left ==> Right Rotation followed by Left Rotation
		if root.right.getBalanceFactor() == 1 {
			root.right = root.right.rotateRight()
			root = root.rotateLeft()
		}
	}

	return root
}