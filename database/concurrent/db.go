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

func (A * Node) rotateRight() *Node {
	B := A.Left
	tmp := B.Right

	B.Right = A
	A.Left = tmp

	return B
}

func (A *Node) rotateLeft() *Node {
	C := A.Right
	tmp := C.Left

	C.Left = A
	A.Right = tmp

	return C
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

	// Conditions under which balanceFactor itself would not lead to a balancing operation:
	//  -1 =< bf <= 1
	if -1 <= balanceFactor && balanceFactor <= 1 {
		return root
	}

	if balanceFactor == 2 {
		// Left-Left ==> Right Rotation
		if root.Left.getBalanceFactor() == 1 || root.Left.getBalanceFactor() == 0 {
			root = root.rotateRight()
		}

		// Left-right ==> Left Rotation followed by Right Rotation
		if root.Left.getBalanceFactor() == -1 {
			root.Left = root.Left.rotateLeft()
			root = root.rotateRight()
		}
	}

	if balanceFactor == -2 {
		// Right-Right ==> Left Rotation
		if root.Right.getBalanceFactor() == -1 || root.Right.getBalanceFactor() == 0 {
			root = root.rotateLeft()
		}

		// Right-Left ==> Right Rotation followed by Left Rotation
		if root.Right.getBalanceFactor() == 1 {
			root.Right = root.Right.rotateRight()
			root = root.rotateLeft()
		}
	}

	return root
}