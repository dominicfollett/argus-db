// Package naive provides a basic implementation of a BST (binary search tree).
// It includes functionalities to insert new nodes and retrieve keys in sorted order.
// It also maintains the worst-case balance factor of the tree as a signal to decide when to swap it with an AVL tree
package naive

import (
	"sync"
	"sync/atomic"
	"time"
)

// BST represents a BST tree with a pointer to the root node.
type BST struct {
	root *Node // Root points to the root node of the AVL tree.
	rootLock sync.Mutex
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

// Insert adds a new node with the given key and data to the BST tree.
// This function will actually also return an integer representing whether the rate limiting call is allowable
// Access to the *Data struct will be synchronized with a mutex
func (tree *BST) Insert(key string, tokens int, capacity int) {

	tree.rootLock.Lock()

	if tree.root == nil {
		// Create the node here
		node := &Node{
			key: key,
			lock: sync.Mutex{},
			data: &Data{ // TODO
				tokens: tokens,
				time: time.Now().String(),
			},
			height: atomic.Int32{}, // Leaves have a height of 0
		}
		tree.root = node
		tree.rootLock.Unlock()
		return
	}

	// tree.RootLock will be released through hand-over-hand locking
	tree.root.insertBST(&tree.rootLock, key, tokens, capacity)
}

func (node *Node) insertBST(parentLock *sync.Mutex, key string, tokens int, capacity int) {

	// Try and obtain this node's lock
	node.lock.Lock()

	// Good now release the prior lock
	parentLock.Unlock()

	// Critical section
	if node.key == key {

		// TODO: Update Data & implement token bucket
		node.data.time = time.Now().String()
		node.data.tokens = tokens

		// Release this node's lock because we're done with it
		node.lock.Unlock()
		return
	}

	if node.key > key{
		if node.right == nil {
			node.right = &Node{
				key: key,
				lock: sync.Mutex{},
				data: &Data{ // TODO
					tokens: tokens,
					time: time.Now().String(),
				},
				height: atomic.Int32{},
			}
			node.lock.Unlock()
			return
		} else {
			node.right.insertBST(&node.lock, key, tokens, capacity)
		}
	}

	if node.key < key {
		if node.left == nil {
			node.left = &Node{
				key: key,
				lock: sync.Mutex{},
				data: &Data{ // TODO
					tokens: tokens,
					time: time.Now().String(),
				},
				height: atomic.Int32{}, // Leaves have a height of 0
			}

			node.lock.Unlock()
			return
		} else {
			node.left.insertBST(&node.lock, key, tokens, capacity)
		}
	}

	// Update the Node's height using atomic operations
	old_height := node.getHeight()
	new_height := 1 + max(node.left.getHeight(), node.right.getHeight())

	if new_height > old_height{
		if !node.height.CompareAndSwap(old_height, new_height) {

			// Something wrote to height before this thread could
			old_height := node.getHeight() // atomic read the latest height
			if new_height > old_height {
				node.height.CompareAndSwap(old_height, new_height)
			}
		}
	}

	//balance_factor := node.left.getHeight() - node.right.getHeight()
}
