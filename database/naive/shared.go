package naive

import (
	"sync"
)

// Shared Data structure stores the Token Bucket particulars
type Data struct {
	Tokens  int
	Time    string
	// and ?
}

// Node represents a single node within a BST.
// It contains the key, associated data, height of the node, and pointers to the left and right child nodes.
type Node struct {
	Lock   sync.Mutex
	Key    string // Key is the unique identifier for the node.
	Data   *Data  // Data points to the associated data of the node.
	Height int    // Height is the height of the node within the tree.
	Left   *Node  // Left points to the left child node.
	Right  *Node  // Right points to the right child node.
}

// The message that is passed over the channel
type Message struct {
	Key string
	Data *Data
}

// inorderDesc traverses the tree in an in-order manner (descending order) and collects the keys.
func (node *Node) inorderDesc(keys []string) []string {
	if node == nil {
		return keys
	}

	keys = node.Right.inorderDesc(keys)
	keys = append(keys, node.Key)
	keys = node.Left.inorderDesc(keys)

	return keys
}

// getHeight returns the height of the node.
// If the node is nil, it returns 0, indicating the height of a non-existent node.
func (node *Node) getHeight() int {
	if node == nil {
		return 0
	}
	return node.Height
}

// getBalanceFactor calculates and returns the balance factor of the node.
// The balance factor is the difference in heights between the left and right subtrees.
func (node *Node) getBalanceFactor() int {
	if node == nil {
		return 0
	}

	return node.Left.getHeight() - node.Right.getHeight()
}