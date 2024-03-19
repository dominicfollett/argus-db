package naive

type NaiveDB struct {
	bst      *BST
	avl      *AVL
	callback func(data any, params any) (any, any, error)
}

func NewDB(callback func(data any, params any) (any, any, error)) *NaiveDB {
	// Start the go routines that manage the BST and AVL trees, etc

	// Start the go routine that listens on the AVL channel and manages the AVL tree

	// We need a parent thread that monitors the BST's memory usage, and the balance factor metric
	// If the memory usage is too high, we need to block, swap out the BST and AVL trees, and write the AVL tree to disk
	// If the balance factor is too high, we need to block, swap out the BST and AVL trees, and instantiate a new AVL tree

	bst := NewBST()
	avl := NewAVL()

	// TODO we need to save the channel for the AVL tree in this struct
	return &NaiveDB{
		bst:      bst,
		avl:      avl,
		callback: callback,
	}
}

func (db *NaiveDB) Calculate(key string, params any) (any, error) {
	node := db.bst.InSearch(key)

	// We must absolutely unlock the node before we return
	defer node.lock.Unlock()

	data, result, err := db.callback(node.data, params)
	if err != nil {
		node.lock.Unlock()
		return false, err
	}

	node.data = data

	return result, nil
}
