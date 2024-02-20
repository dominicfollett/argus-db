// Package naive provides a basic implementation of an AVL (Adelson-Velsky and Landis) tree.
// It includes functionalities to insert new nodes and retrieve keys in sorted order.
package naive

import "sync"

// AVL represents an AVL tree with a pointer to the root node.
type AVL struct {
	Root *Node // Root points to the root node of the AVL tree.
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewAVL() *AVL {
	return &AVL{}
}

// Insert adds a new node with the given key and data to the AVL tree.
// It ensures that the tree remains balanced after the insertion.
func (tree *AVL) Insert(key string, data *Data) {
	tree.Root = tree.Root.insertAVL(key, data)
}

// GetKeys retrieves all keys from the AVL tree in sorted descending order.
func (tree *AVL) GetKeys() []string {
	keys := tree.Root.inorderDesc([]string{})
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

	B := A.Left
	tmp := B.Right

	B.Right = A
	A.Left = tmp

	// Fix the Node heights from child to parent
	A.Height = 1 + max(A.Left.getHeight(), A.Right.getHeight())
	B.Height = 1 + max(B.Left.getHeight(), B.Right.getHeight())

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

	C := A.Right
	tmp := C.Left

	C.Left = A
	A.Right = tmp

	// Fix the Node heights from child to parent
	A.Height = 1 + max(A.Left.getHeight(), A.Right.getHeight())
	C.Height = 1 + max(C.Left.getHeight(), C.Right.getHeight())

	return C
}

// insertAVL adds a new node with the given key and data to the tree rooted at the current node.
// It ensures the AVL tree properties are maintained by performing necessary rotations.
func (root *Node) insertAVL(key string, data *Data) *Node  {
	
	if root == nil {
		return &Node{
			Key: key,
			Lock: sync.Mutex{},
			Data: data,
			Height: 0, // Leaves have a height of 0
		}
	}

	if root.Key == key {
		// Nothing further to do so we can safely return
		return root
	}

	if root.Key > key{
		root.Right = root.Right.insertAVL(key, data)
	}

	if root.Key < key {
		root.Left = root.Left.insertAVL(key, data)
	}

	// Update height
	root.Height = 1 + max(root.Left.getHeight(), root.Right.getHeight())

	balanceFactor := root.Left.getHeight() - root.Right.getHeight() 

	// Conditions under which balanceFactor itself would not lead to a balancing operation:
	//  -1 =< bf <= 1
	if -1 <= balanceFactor && balanceFactor <= 1 {
		return root
	}

	// TODO: Optimize this
	if balanceFactor == 2 {
		// Left-Left ==> Right Rotation
		if root.Left.getBalanceFactor() == 1 || root.Left.getBalanceFactor() == 0 {
			root = root.rotateRight()
		}

		// Left-right ==> Left Rotation followed by Right Rotation
		if root.Left.getBalanceFactor() == -1 {
			root.Left = root.Left.rotateLeft()
			root = root.rotateRight()
		}
	}

	if balanceFactor == -2 {
		// Right-Right ==> Left Rotation
		if root.Right.getBalanceFactor() == -1 || root.Right.getBalanceFactor() == 0 {
			root = root.rotateLeft()
		}

		// Right-Left ==> Right Rotation followed by Left Rotation
		if root.Right.getBalanceFactor() == 1 {
			root.Right = root.Right.rotateRight()
			root = root.rotateLeft()
		}
	}

	return root
}