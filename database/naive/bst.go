// Package naive provides a basic implementation of a BST (binary search tree).
// It includes functionalities to insert new nodes and retrieve keys in sorted order.
// It also maintains the worst-case balance factor of the tree as a signal to decide when to swap it with an AVL tree
package naive

import (
	"sync"
	"time"
)

// BST represents a BST tree with a pointer to the root node.
type BST struct {
	Root *Node // Root points to the root node of the AVL tree.
	RootLock sync.Mutex
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewBST() *BST {
	return &BST{
		RootLock: sync.Mutex{},
	}
}

// GetKeys retrieves all keys from the BST tree in sorted descending order.
func (tree *BST) GetKeys() []string {
	keys := tree.Root.inorderDesc([]string{})
	return keys
}

// Insert adds a new node with the given key and data to the BST tree.
// This function will actually also return an integer representing whether the rate limiting call is allowable
// Access to the *Data struct will be synchronized with a mutex
func (tree *BST) Insert(key string, tokens int, capacity int) {

	tree.RootLock.Lock()

	if tree.Root == nil {
		// Create the node here
		node := &Node{
			Key: key,
			Lock: sync.Mutex{},
			Data: &Data{ // TODO
				Tokens: tokens,
				Time: time.Now().String(),
			},
			Height: 0, // Leaves have a height of 0
		}
		tree.Root = node
		tree.RootLock.Unlock()
		return
	}

	// tree.RootLock will be released through hand-over-hand locking
	tree.Root.insertBST(&tree.RootLock, key, tokens, capacity)
}

func (node *Node) insertBST(parentLock *sync.Mutex, key string, tokens int, capacity int) {

	// Try and obtain this node's lock
	node.Lock.Lock()

	// Good now release the prior lock
	parentLock.Unlock()

	// Critical section
	if node.Key == key {

		// TODO: Update Data & implement token bucket
		node.Data.Time = time.Now().String()
		node.Data.Tokens = tokens

		// Release this node's lock because we're done with it
		node.Lock.Unlock()
		return
	}

	if node.Key > key{
		if node.Right == nil {
			node.Right = &Node{
				Key: key,
				Lock: sync.Mutex{},
				Data: &Data{ // TODO
					Tokens: tokens,
					Time: time.Now().String(),
				},
				Height: 0,
			}

			node.Lock.Unlock()
			return
		} else {
			node.Right.insertBST(&node.Lock, key, tokens, capacity)
		}
	}

	if node.Key < key {
		if node.Left == nil {
			node.Left = &Node{
				Key: key,
				Lock: sync.Mutex{},
				Data: &Data{ // TODO
					Tokens: tokens,
					Time: time.Now().String(),
				},
				Height: 0, // Leaves have a height of 0
			}

			node.Lock.Unlock()
			return
		} else {
			node.Left.insertBST(&node.Lock, key, tokens, capacity)
		}
	}
}
