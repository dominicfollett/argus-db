// Package naive provides a basic implementation of an AVL (Adelson-Velsky and Landis) tree.
// It includes functionalities to insert new nodes and retrieve keys in sorted order.
package naive

import (
	"sync/atomic"
)

// AVL represents an AVL tree with a pointer to the root node.
type AVL struct {
	root *Node
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewAVL() *AVL {
	return &AVL{}
}

// TODO can we use *[]string instead
func (tree *AVL) Survey(evict func(data any) bool) []string {
	return tree.root.surveyAVL(evict, []string{})
}

// surveyAVL traverses the tree in an in-order manner (descending order) and collects the keys
// that return true when applying the user-defined evict function on the node's data.
func (node *Node) surveyAVL(evict func(data any) bool, keys []string) []string {
	if node == nil {
		return keys
	}

	keys = node.left.surveyAVL(evict, keys)

	if evict(node.data) {
		keys = append(keys, node.key)
	}

	keys = node.right.surveyAVL(evict, keys)

	return keys
}

// GetKeys retrieves all keys from the AVL tree in sorted descending order.
func (tree *AVL) GetKeys() []string {
	keys := tree.root.inorderDesc([]string{})
	return keys
}

// Delete removes a node from the AVL tree with the given key.
func (tree *AVL) Delete(key string) {
	tree.root = tree.root.deleteAVL(key)
}

// deleteAVL searches for the the node to delete, removes it from the tree while maintaining height
// invariance through rebalancing operations.
func (node *Node) deleteAVL(key string) *Node {
	if node == nil {
		return nil
	}

	if node.key == key {
		if node.left == nil && node.right == nil {
			return nil
		}

		switch {
		case node.left == nil:
			return node.right
		case node.right == nil:
			return node.left
		}

		// Find node successor
		successorNode := findAndRemoveMinimum(node.right)

		// Update this node to mirror the successor
		node.key = successorNode.key
		node.data = successorNode.data

		// Update the node's height because left or right subtree heights may have changed
		node.height.Store(
			1 + max(node.left.getHeight(), node.right.getHeight()),
		)

		// Balance if required
		node = node.balance()

		return node
	}

	if key < node.key {
		node.left = node.left.deleteAVL(key)
	} else if key > node.key {
		node.right = node.right.deleteAVL(key)
	}

	// Update the node's height because left or right subtree heights may have changed
	node.height.Store(
		1 + max(node.left.getHeight(), node.right.getHeight()),
	)

	// Balance if required
	node = node.balance()

	return node
}

// Insert adds a new node with the given key and data to the AVL tree. It ensures that the tree
// remains balanced after the insertion.
func (tree *AVL) Insert(key string, data any) {
	tree.root = tree.root.insertAVL(key, data)
}

// insertAVL adds a new node with the given key and data to the tree rooted at the current node.
// It ensures the AVL tree properties are maintained by performing necessary rotations.
func (node *Node) insertAVL(key string, data any) *Node {
	if node == nil {
		return &Node{
			key:    key,
			data:   data,
			height: atomic.Int32{}, // Leaves have a height of 0
		}
	}

	if node.key == key {
		// Nothing further to do so we can safely return
		return node
	}

	if key < node.key {
		node.left = node.left.insertAVL(key, data)
	}

	if key > node.key {
		node.right = node.right.insertAVL(key, data)
	}

	// Update height
	node.height.Store(
		1 + max(node.left.getHeight(), node.right.getHeight()),
	)

	// Balance if required
	node = node.balance()

	return node
}

// findAndRemoveMinimum finds the minimum node from the given root,
// removes it from the tree, and updates the height of each node
// back to the root.
func findAndRemoveMinimum(root *Node) *Node {
	if root.left != nil {
		successorNode := findAndRemoveMinimum(root.left)

		// Delete the successor node
		if successorNode.key == root.left.key {
			root.left = nil
		}

		// Update the node's height
		root.height.Store(
			1 + max(root.left.getHeight(), root.right.getHeight()),
		)

		// TODO: rebalancing is not required because ...

		return successorNode
	}

	return root
}

// balance calculates the node balance factor and applies
// balacing operations if required.
func (node *Node) balance() *Node {
	balanceFactor := node.getBalanceFactor()

	// Conditions under which balanceFactor itself would not lead to a balancing operation:
	//  -1 =< bf <= 1
	if -1 <= balanceFactor && balanceFactor <= 1 {
		return node
	}

	// TODO: Optimize this
	//nolint:gomnd // 2 is the threshold for AVL balancing
	if balanceFactor == 2 {
		// Left-Left ==> Right Rotation
		if node.left.getBalanceFactor() == 1 || node.left.getBalanceFactor() == 0 {
			node = node.rotateRight()
		}

		// Left-right ==> Left Rotation followed by Right Rotation
		if node.left.getBalanceFactor() == -1 {
			node.left = node.left.rotateLeft()
			node = node.rotateRight()
		}
	}

	if balanceFactor == -2 {
		// Right-Right ==> Left Rotation
		if node.right.getBalanceFactor() == -1 || node.right.getBalanceFactor() == 0 {
			node = node.rotateLeft()
		}

		// Right-Left ==> Right Rotation followed by Left Rotation
		if node.right.getBalanceFactor() == 1 {
			node.right = node.right.rotateRight()
			node = node.rotateLeft()
		}
	}

	return node
}

// rotateRight performs a right rotation on the node. This is used to rebalance the tree in case of
// a left-heavy imbalance.
//
//nolint:revive // so that the diagram makes sense
func (a *Node) rotateRight() *Node {
	/*
					 A
					/ \
				   B   C
				  / \
				 D   E
				/
		 	  F
	*/

	b := a.left
	tmp := b.right

	b.right = a
	a.left = tmp

	// Fix the Node heights from child to parent
	a.height.Store(
		1 + max(
			a.left.getHeight(),
			a.right.getHeight(),
		),
	)
	b.height.Store(
		1 + max(
			b.left.getHeight(),
			b.right.getHeight(),
		),
	)

	return b
}

// rotateLeft performs a left rotation on the node. This is used to rebalance the tree in case of a
// right-heavy imbalance.
//
//nolint:revive // so that the diagram makes sense
func (a *Node) rotateLeft() *Node {
	/*
			 A
			/ \
		   B   C
			  / \
			 F   D
				  \
				   E
	*/

	c := a.right
	tmp := c.left

	c.left = a
	a.right = tmp

	// Fix the Node heights from child to parent
	a.height.Store(
		1 + max(
			a.left.getHeight(),
			a.right.getHeight(),
		),
	)
	c.height.Store(
		1 + max(
			c.left.getHeight(),
			c.right.getHeight(),
		),
	)

	return c
}
