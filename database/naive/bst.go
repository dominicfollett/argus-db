// Package naive provides a basic implementation of a BST (binary search tree).
// It includes functionality to insert new nodes and retrieve keys in sorted order.
package naive

import (
	"sync"
	"sync/atomic"
	"time"
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

// newBSTNode creates and returns a new instance of a BST node with the given key and data.
func newBSTNode(key string, tokens int) *Node {
	return &Node{
		key:  key,
		lock: sync.Mutex{},
		data: &Data{ // TODO
			tokens: tokens,
			time:   time.Now().String(),
		},
		height: atomic.Int32{},
	}
}

// Insert adds a new node with the given key and data to the BST tree.
// This function will also return an integer representing whether the rate limiting call is allowable
func (tree *BST) Insert(key string, tokens int, capacity int) {

	tree.rootLock.Lock()

	if tree.root == nil {
		node := newBSTNode(key, tokens)
		tree.root = node
		tree.rootLock.Unlock()
		return
	}

	// tree.rootLock will be released through hand-over-hand locking
	tree.root.insertBST(&tree.rootLock, key, tokens, capacity)
}

// updateHeight atomically updates the height of the node based on the height of its left and right children.
func (node *Node) updateHeight(left_height int32, right_height int32) {

	old_height := node.getHeight()
	new_height := 1 + max(left_height, right_height)

	for new_height > old_height {
		if node.height.CompareAndSwap(old_height, new_height) {
			break
		}

		// Something wrote to height before this thread could therefore atomic read the latest height
		old_height = node.getHeight()
	}
}

// insertBST performs a recursive search and when applicable insertion of a new node with the given key and data to the BST tree.
// This method is thread-safe and uses hand-over-hand locking to ensure that the tree is properly locked during the insertion process.
// The method returns the height of the node after the insertion.
func (node *Node) insertBST(parentLock *sync.Mutex, key string, tokens int, capacity int) int32 {

	// Try and obtain this node's lock
	node.lock.Lock()

	// Good now release the prior lock
	parentLock.Unlock()

	// Critical section
	if node.key == key {

		// TODO: Update Data & implement token bucket
		node.data.time = time.Now().String()
		node.data.tokens = tokens

		// Might as well update the height of this node
		node.updateHeight(node.left.getHeight(), node.right.getHeight())

		// Release this node's lock because we're done with it
		node.lock.Unlock()
		return node.getHeight()
	}

	var left_height int32
	var right_height int32

	if key < node.key {
		if node.left == nil {
			node.left = newBSTNode(key, tokens)

			// We have to update this node's height because we've just performed an insertion
			node.updateHeight(node.left.getHeight(), node.right.getHeight())

			node.lock.Unlock()
			return node.getHeight() // or just new_height?
		} else {
			right_height = node.right.getHeight()

			// node.lock will be released in the recursive call
			left_height = node.left.insertBST(&node.lock, key, tokens, capacity)
		}
	}

	if key > node.key {
		if node.right == nil {
			node.right = newBSTNode(key, tokens)

			// We have to update this node's height because we've just performed an insertion
			node.updateHeight(node.left.getHeight(), node.right.getHeight())

			node.lock.Unlock()
			return node.getHeight() // or just new_height?
		} else {
			left_height = node.left.getHeight()

			// node.lock will be released in the recursive call
			right_height = node.right.insertBST(&node.lock, key, tokens, capacity)
		}
	}

	node.updateHeight(left_height, right_height)

	// TODO: Calculate balance factor, and atomically update the global counter
	// balance_factor := absInt32(left_height - right_height)
	return node.getHeight()
}
