package naive

import (
	"sync"
	"sync/atomic"
)

// Shared Data structure stores the Token Bucket particulars
type Data struct {
	tokens int
	time   string
	// and ?
}

// Node represents a single node within a BST.
// It contains the key, associated data, height of the node, and pointers to the left and right child nodes.
type Node struct {
	lock   sync.Mutex
	key    string       // Key is the unique identifier for the node.
	data   *Data        // Data points to the associated data of the node.
	height atomic.Int32 // Height is the height of the node within the tree.
	left   *Node        // Left points to the left child node.
	right  *Node        // Right points to the right child node.
}

// The message that is passed over the channel
type Message struct {
	Key  string
	Data *Data
}

// inorderDesc traverses the tree in an in-order manner (descending order) and collects the keys.
func (node *Node) inorderDesc(keys []string) []string {
	if node == nil {
		return keys
	}

	keys = node.left.inorderDesc(keys)
	keys = append(keys, node.key)
	keys = node.right.inorderDesc(keys)

	return keys
}

// getHeight atomically returns the height of the node.
// If the node is nil, it returns -1, indicating the height of a non-existent node.
func (node *Node) getHeight() int32 {
	if node == nil {
		return -1
	}
	return node.height.Load()
}

// getBalanceFactor atomically calculates and returns the balance factor of the node.
// The balance factor is the difference in heights between the left and right subtrees.
func (node *Node) getBalanceFactor() int32 {
	if node == nil {
		return 0
	}

	return node.left.getHeight() - node.right.getHeight()
}

// absInt32 returns the magnitude of the given scalar
func absInt32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
