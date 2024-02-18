// Package naive provides a basic implementation of an AVL (Adelson-Velsky and Landis) tree.
// It includes functionalities to insert new nodes and retrieve keys in sorted order.
package naive

// Node represents a single node within an AVL tree.
// It contains the key, associated data, height of the node, and pointers to the left and right child nodes.
type Node struct {
	Key    string // Key is the unique identifier for the node.
	Data   *Data  // Data points to the associated data of the node.
	Height int    // Height is the height of the node within the tree.
	Left   *Node  // Left points to the left child node.
	Right  *Node  // Right points to the right child node.
}

// AVL represents an AVL tree with a pointer to the root node.
type AVL struct {
	Root *Node // Root points to the root node of the AVL tree.
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewAVL() *AVL {
	return &AVL{}
}

// Insert adds a new node with the given key and data to the AVL tree.
// It ensures that the tree remains balanced after the insertion.
func (tree *AVL) Insert(key string, data *Data) {
	tree.Root = tree.Root.insert(key, data)
}

// GetKeys retrieves all keys from the AVL tree in sorted descending order.
func (tree *AVL) GetKeys() []string {
	keys := tree.Root.inorderDesc([]string{})
	return keys
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

// rotateRight performs a right rotation on the node.
// This is used to rebalance the tree in case of a left-heavy imbalance.
func (A * Node) rotateRight() *Node {
	/*
			 A
			/ \
		   B   C
		  / \
		 D   E
		/
 	  F
	*/

	B := A.Left
	tmp := B.Right

	B.Right = A
	A.Left = tmp

	// Fix the Node heights from child to parent
	A.Height = 1 + max(A.Left.getHeight(), A.Right.getHeight())
	B.Height = 1 + max(B.Left.getHeight(), B.Right.getHeight())

	return B
}

// rotateLeft performs a left rotation on the node.
// This is used to rebalance the tree in case of a right-heavy imbalance.
func (A *Node) rotateLeft() *Node {
	/*
		 A
		/ \
	   B   C
		  / \
		 F   D
			  \
			   E
	*/

	C := A.Right
	tmp := C.Left

	C.Left = A
	A.Right = tmp

	// Fix the Node heights from child to parent
	A.Height = 1 + max(A.Left.getHeight(), A.Right.getHeight())
	C.Height = 1 + max(C.Left.getHeight(), C.Right.getHeight())

	return C
}

// insert adds a new node with the given key and data to the tree rooted at the current node.
// It ensures the AVL tree properties are maintained by performing necessary rotations.
func (root *Node) insert(key string, data *Data) *Node  {
	
	if root == nil {
		return &Node{
			Key: key,
			Data: data,
			Height: 0, // Leaves have a height of 0
		}
	}

	if root.Key == key {
		// Nothing further to do so we can safely return
		return root
	}

	if root.Key > key{
		root.Right = root.Right.insert(key, data)
	}

	if root.Key < key {
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

	// TODO: Optimize this
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