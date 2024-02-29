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
	root     *Node // Root points to the root node of the AVL tree.
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

// newBSTNode creates and returns a new instance of a BST node with the given key and data.
func newBSTNode(key string, tokens int) *Node {
	node := &Node{
		key:  key,
		lock: sync.Mutex{},
		data: &Data{ // TODO
			tokens: tokens,
			time:   time.Now().String(),
		},
		height: atomic.Int32{},
	}

	// TODO: Mind BLOWN
	node.height.Store(1)
	return node
}

// Insert adds a new node with the given key and data to the BST tree.
// This function will actually also return an integer representing whether the rate limiting call is allowable
// Access to the *Data struct will be synchronized with a mutex
func (tree *BST) Insert(key string, tokens int, capacity int) {

	tree.rootLock.Lock()

	if tree.root == nil {
		// Create the node here
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

	if left_height == -1 {
		left_height = node.left.getHeight()
	}
	if right_height == -1 {
		right_height = node.right.getHeight()
	}

	old_height := node.getHeight()
	new_height := 1 + max(left_height, right_height)

	// TODO: Can we do away with this for loop?
	for new_height > old_height {
		if node.height.CompareAndSwap(old_height, new_height) {
			break
		}

		// Something wrote to height before this thread could therefore atomic read the latest height
		old_height = node.getHeight()
	}
}

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
		// TODO: does this make any difference?
		node.updateHeight(-1, -1)

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
			// TODO: Assuming the Left Node's height to 1 here makes NO sense
			node.updateHeight(1, -1)

			node.lock.Unlock()
			return node.getHeight() // or just new_height?
		} else {
			// retrieve the right node height and get left node height from recursive call??
			right_height = node.right.getHeight()
			// node.lock will be released in the recusrive call
			left_height = node.left.insertBST(&node.lock, key, tokens, capacity)
		}
	}

	if key > node.key {
		if node.right == nil {
			node.right = newBSTNode(key, tokens)

			// We have to update this node's height because we've just performed an insertion
			// TODO: Assuming the Right Node's height to 1 here makes NO sense
			node.updateHeight(-1, 1)

			node.lock.Unlock()
			return node.getHeight() // or just new_height?
		} else {
			// retrieve the left node height and get right node height from recursive call??
			left_height = node.left.getHeight()
			// node.lock will be released in the recusrive call
			right_height = node.right.insertBST(&node.lock, key, tokens, capacity)
		}
	}

	// TODO: This section is equivalent to node.height.Store(new_height) ??? How???
	node.updateHeight(left_height, right_height)

	//balance_factor := left_height - right_height

	return node.getHeight()
}
