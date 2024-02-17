package concurrent

import (
	"bytes"
	"sync"
)

// Shared Data structure stores the Token Bucket particulars
type Data struct {
	Lock 	sync.Mutex
	Tokens  int
	Time    string
	// and ?
}

// The message that is passed over the channel
type Message struct {
	Key []byte
	Data *Data
}

type Node struct {
	Key []byte  // Should/Can we make this a fixed size?
	Data *Data
	Height int
	Left  *Node
	Right *Node
}

type AVL struct {
	Root *Node
}

func NewShadowAVL() *AVL {
	return &AVL{}
}

func (tree *AVL) Insert(key []byte, data *Data) {
	// TODO: Is this assignment necessary??
	tree.Root = tree.Root.insert(key, data)
}

func (node *Node) getHeight() int {
	if node == nil {
		return 0
	}
	return node.Height
}

func (node *Node) getBalanceFactor() int {
	if node == nil {
		return 0
	}

	return node.Left.getHeight() - node.Right.getHeight()
}

func (root *Node) insert(key []byte, data *Data) *Node  {
	
	if root == nil {
		return &Node{
			Key: key,
			Data: data,
			Height: 0, // Leaves have a height of 0
		}
	}

	direction := bytes.Compare(root.Key, key)

	// root.Key == key
	if direction == 0 {
		// Nothing further to do so we can safely return
		return root
	}

	// root.Key > key
	if direction == 1 {
		root.Right = root.Right.insert(key, data)
	}

	// root.Key < key
	if direction == -1 {
		root.Left = root.Left.insert(key, data)
	}

	// Update height
	root.Height = 1 + max(root.Left.getHeight(), root.Right.getHeight())

	balanceFactor := root.Left.getHeight() - root.Right.getHeight() 

	// What are the conditions under which balanceFactor itself would not lead to a balancing operation?
	//  -1 =< bf <= 1
	// If it exceeds this are we always guaranteed to have a left AND right child?
	// Or does the sign tell us that a left or right child might exist

	/*
	leftCMP := bytes.Compare(key, root.Left.Key)
	rightCMP := bytes.Compare(key, root.Right.Key)

	// Left Left Case
	if balance > 1 && leftCMP == -1 { // key < root.Left.Key
		// return rightRotate(root)
	}

	// ...
	*/

	return nil
}