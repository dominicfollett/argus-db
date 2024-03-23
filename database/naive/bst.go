// Package naive provides a basic implementation of a BST (binary search tree).
// It includes functionality to insert new nodes and retrieve keys in sorted order.
package naive

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var BF_THRESHOLD int32 = 5

// BST represents a BST tree with a pointer to the root node.
type BST struct {
	root             *Node      // Root points to the root node of the AVL tree.
	rootLock         sync.Mutex // Lock for the root node
	balanceFactorSum *atomic.Int64
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewBST() *BST {
	return &BST{
		rootLock:         sync.Mutex{},
		balanceFactorSum: &atomic.Int64{},
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

// InSearch retrieves the node with the given key from the BST tree. If the node does not exist, it creates a new node.
// This function is thread-safe and uses hand-over-hand locking to ensure that the tree is properly
// locked during the search process. However, users of this function MUST release the lock on the
// returned node after they are done with it.
func (tree *BST) InSearch(key string) *Node {

	tree.rootLock.Lock()

	if tree.root == nil {
		tree.root = newBSTNode(key)
		tree.rootLock.Unlock()

		tree.root.lock.Lock()
		return tree.root
	}

	// tree.rootLock will be released through hand-over-hand locking
	_, node, balanceFactor := tree.root.inSearchBST(&tree.rootLock, key)

	// Atomically update the global balance factor sum
	tree.balanceFactorSum.Add(int64(balanceFactor))

	fmt.Printf("Balance factor sum: %d\n", tree.balanceFactorSum.Load())

	return node
}

// inSearchBST retrieves the node with the given key from the BST tree. If the node does not exist, it creates a new node.
// This function is thread-safe and uses hand-over-hand locking to ensure that the tree is properly
// locked during the search process.
func (node *Node) inSearchBST(parentLock *sync.Mutex, key string) (int32, *Node, int32) {

	// Try and obtain this node's lock
	node.lock.Lock()

	// Good now release the prior lock
	parentLock.Unlock()

	if node.key == key {
		// Might as well update the height of this node
		node.updateHeight(node.left.getHeight(), node.right.getHeight())

		// Calculate and add the balance factor
		balanceFactor := absInt32(node.left.getHeight() - node.right.getHeight())
		if balanceFactor < BF_THRESHOLD {
			balanceFactor = 0
		}

		return node.getHeight(), node, balanceFactor
	}

	var leftHeight int32
	var rightHeight int32
	var returnedNode *Node
	var balanceFactor int32

	if key < node.key {
		if node.left == nil {
			node.left = newBSTNode(key)

			// We have to update this node's height because we've just performed an insertion
			node.updateHeight(node.left.getHeight(), node.right.getHeight())

			node.left.lock.Lock()

			// We have to release the lock on this node because we're done with it
			node.lock.Unlock()
			return node.getHeight(), node.left, 0
		} else {
			rightHeight = node.right.getHeight()

			// node.lock will be released in the recursive call
			leftHeight, returnedNode, balanceFactor = node.left.inSearchBST(&node.lock, key)
		}
	}

	if key > node.key {
		if node.right == nil {
			node.right = newBSTNode(key)

			// We have to update this node's height because we've just performed an insertion
			node.updateHeight(node.left.getHeight(), node.right.getHeight())

			node.right.lock.Lock()

			// We have to release the lock on this node because we're done with it
			node.lock.Unlock()
			return node.getHeight(), node.right, 0
		} else {
			leftHeight = node.left.getHeight()

			// node.lock will be released in the recursive call
			rightHeight, returnedNode, balanceFactor = node.right.inSearchBST(&node.lock, key)
		}
	}

	node.updateHeight(leftHeight, rightHeight)

	// Calculate balance factor
	balanceFactorPrime := absInt32(leftHeight - rightHeight)
	if balanceFactorPrime < BF_THRESHOLD {
		balanceFactorPrime = 0
	}

	return node.getHeight(), returnedNode, balanceFactor + balanceFactorPrime
}
