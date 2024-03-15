// Package naive provides a basic implementation of a BST (binary search tree).
// It includes functionality to insert new nodes and retrieve keys in sorted order.
package naive

import (
	"sync"
	"sync/atomic"
)

// BST represents a BST tree with a pointer to the root node.
type BST struct {
	root     *Node      // Root points to the root node of the AVL tree.
	rootLock sync.Mutex // Lock for the root node
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewBST() *BST {
	return &BST{
		rootLock: sync.Mutex{},
	}
}

// GetKeys retrieves all keys from the BST tree in sorted descending order.
func (tree *BST) GetKeys() []string {
	keys := tree.root.inorderDesc([]string{})
	return keys
}

// Search retrieves the node with the given key from the BST tree. If the node does not exist, it returns nil.
// This function is thread-safe and uses hand-over-hand locking to ensure that the tree is properly
// locked during the search process. However, users of this function MUST release the lock on the
// returned node after they are done with it.
func (tree *BST) Search(key string) *Node {
	tree.rootLock.Lock()

	if tree.root == nil {
		tree.rootLock.Unlock()
		return nil
	}

	return tree.root.searchBST(&tree.rootLock, key)
}

func (node *Node) searchBST(parentLock *sync.Mutex, key string) *Node {
	// Try and obtain this node's lock
	node.lock.Lock()

	// Good now release the prior lock
	parentLock.Unlock()

	if key == node.key {
		return node
	}

	var result *Node

	if key < node.key {
		if node.left == nil {
			return nil
		}
		result = node.left.searchBST(&node.lock, key)
	}

	if key > node.key {
		if node.right == nil {
			return nil
		}
		result = node.right.searchBST(&node.lock, key)
	}

	return result
}

// updateHeight atomically updates the height of the node based on the height of its left and right children.
func (node *Node) updateHeight(leftHeight int32, rightHeight int32) {

	oldHeight := node.getHeight()
	newHeight := 1 + max(leftHeight, rightHeight)

	for newHeight > oldHeight {
		if node.height.CompareAndSwap(oldHeight, newHeight) {
			break
		}

		// Something wrote to height before this thread could therefore atomic read the latest height
		oldHeight = node.getHeight()
	}
}

// newBSTNode creates and returns a new instance of a BST node with the given key.
func newBSTNode(key string) *Node {
	return &Node{
		key:    key,
		lock:   sync.Mutex{},
		data:   nil,
		height: atomic.Int32{},
	}
}

// Insert adds a new node with the given key and data to the BST tree.
// This function is thread-safe.
func (tree *BST) Insert(key string) {

	tree.rootLock.Lock()

	if tree.root == nil {
		tree.root = newBSTNode(key)
		tree.rootLock.Unlock()
		return
	}

	// tree.rootLock will be released through hand-over-hand locking
	_ = tree.root.insertBST(&tree.rootLock, key)
}

// insertBST adds a new node with the given key to the tree and returns the height of the tree and the new node.
// This function is thread-safe and uses hand-over-hand locking to ensure that the tree is properly
// locked during the insertion process.
func (node *Node) insertBST(parentLock *sync.Mutex, key string) int32 {

	// Try and obtain this node's lock
	node.lock.Lock()

	// Good now release the prior lock
	parentLock.Unlock()

	if node.key == key {
		// Might as well update the height of this node
		node.updateHeight(node.left.getHeight(), node.right.getHeight())

		// We have to release the lock on this node because we're done with it
		node.lock.Unlock()
		return node.getHeight()
	}

	var leftHeight int32
	var rightHeight int32

	if key < node.key {
		if node.left == nil {
			node.left = newBSTNode(key)

			// We have to update this node's height because we've just performed an insertion
			node.updateHeight(node.left.getHeight(), node.right.getHeight())

			// We have to release the lock on this node because we're done with it
			node.lock.Unlock()
			return node.getHeight()
		} else {
			rightHeight = node.right.getHeight()

			// node.lock will be released in the recursive call
			leftHeight = node.left.insertBST(&node.lock, key)
		}
	}

	if key > node.key {
		if node.right == nil {
			node.right = newBSTNode(key)

			// We have to update this node's height because we've just performed an insertion
			node.updateHeight(node.left.getHeight(), node.right.getHeight())

			// We have to release the lock on this node because we're done with it
			node.lock.Unlock()
			return node.getHeight()
		} else {
			leftHeight = node.left.getHeight()

			// node.lock will be released in the recursive call
			rightHeight = node.right.insertBST(&node.lock, key)
		}
	}

	node.updateHeight(leftHeight, rightHeight)

	// TODO: Calculate balance factor, and atomically update the global counter
	// balanceFactor := absInt32(leftHeight - rightHeight)
	return node.getHeight()
}
