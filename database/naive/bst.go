// Package naive provides a basic implementation of a BST (binary search tree).
// It includes functionalities to insert new nodes and retrieve keys in sorted order.
// It also maintains the worst-case balance factor of the tree as a signal to decide when to swap it with an AVL tree
package naive

// BST represents a BST tree with a pointer to the root node.
type BST struct {
	Root *Node // Root points to the root node of the AVL tree.
}

// NewAVL creates and returns a new instance of an AVL tree.
func NewBST() *BST {
	return &BST{}
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
	tree.Root = tree.Root.insertBST(key, tokens, capacity)
}

func (node *Node) insertBST(key string, tokens int, capacity int) *Node /*allowed int ??*/ {
	
	// TODO: We actually need to synchronize node creation
	if node == nil {
		// perform calculations
		// create Data struct and store data

		// return node
	}

	if node.Key == key {
		// Obtain the data mutex
		node.Data.Lock.Lock()
		// Do calculations
		// return the results

		node.Data.Lock.Unlock()
		return node /*, allowed*/
	}

	if node.Key > key{
		if node.Right == nil {
			// obtain create/insert mutex ??
			node.Right = node.Right.insertBST(key, tokens, capacity)
			// unlock
		} else {
			node.Right = node.Right.insertBST(key, tokens, capacity)
		}
	}

	if node.Key < key {
		if node.Left == nil {
			// obtain create/insert mutex ??
			node.Left = node.Left.insertBST(key, tokens, capacity)
			// unlock
		} else {
			node.Left = node.Left.insertBST(key, tokens, capacity)
		}
	}

	node.Height = 1 + max(node.Left.getHeight(), node.Right.getHeight())
	balanceFactor := node.getBalanceFactor()

	// TODO remove
	println(balanceFactor)

	// TODO Update global balance factor -- atomic cmpSwap?



	return node
}