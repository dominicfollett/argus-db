//nolint:testpackage // Allow tests to access the naive package
package naive

import (
	"sort"
	"testing"
)

//nolint:stylecheck // Allow helper function name
func (node *Node) avlHeightTestHelper(t *testing.T) int32 {
	if node == nil {
		if node.getHeight() != -1 {
			t.Errorf("Expected height of nil node to be -1, got %d", node.getHeight())
		}
		return -1
	}

	leftHeight := node.left.avlHeightTestHelper(t)
	rightHeight := node.right.avlHeightTestHelper(t)

	expectedHeight := 1 + max(leftHeight, rightHeight)

	// Check balance factor is correct
	balanceFactor := leftHeight - rightHeight
	if -1 > balanceFactor || balanceFactor > 1 {
		t.Errorf("%s: balance factor magnitude too great %d", node.key, balanceFactor)
	}

	if absInt32(expectedHeight-node.getHeight()) != 0 {
		t.Errorf(
			"%s: Difference between expected height: %d, and actual height: %d, is greater than zero! "+
				"Left height: %d, Right height: %d",
			node.key,
			expectedHeight,
			node.getHeight(),
			leftHeight,
			rightHeight,
		)
	}

	return expectedHeight
}

func TestAVL(t *testing.T) {
	keys := []string{
		"T", "X", "G", "L", "E", "Q", "M", "H", "O", "I", "B", "Z", "A", "V", "S", "R", "K", "P",
		"C", "D", "U", "F", "N", "W", "Y", "J",
	}
	avl := NewAVL()

	for _, k := range keys {
		avl.Insert(k, nil)
	}

	// Add some duplicates
	avl.Insert("T", nil)
	avl.Insert("D", nil)
	avl.Insert("N", nil)
	avl.Insert("P", nil)

	result := avl.GetKeys()
	sort.Strings(keys)

	if len(keys) != len(result) {
		t.Errorf("Expected and result slices differ in length; expected: %d, got: %d", len(keys), len(result))
	}

	for i, expectedKey := range keys {
		if expectedKey != result[i] {
			t.Errorf("Key mismatch at index %d; expected: %s, got: %s", i, expectedKey, result[i])
		}
	}

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)
}

func TestAVLDelete(t *testing.T) {
	keys := []string{"U", "R", "X", "N", "T", "W", "Y", "M", "P", "S", "V", "Z", "O", "Q"}

	/*
	    	   U
	    	  /  \
	    	 /    \
	    	R      X
	   	   /  \    / \
	      N    T  W   Y
	     / \   /  /    \
	    M  P  S  V      Z
	      / \
	     O   Q
	*/

	avl := NewAVL()

	for _, k := range keys {
		avl.Insert(k, nil)
	}

	// Add some duplicates
	avl.Insert("X", nil)
	avl.Insert("Y", nil)
	avl.Insert("V", nil)
	avl.Insert("U", nil)

	// Delete where node has two children
	avl.Delete("R")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)

	// Delete where node has right child only
	avl.Delete("Y")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)

	// Delete where node has left child only
	avl.Delete("W")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)

	// Delete where node is a leaf
	avl.Delete("O")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)

	expectedOrder := []string{"M", "N", "P", "Q", "S", "T", "U", "V", "X", "Z"}
	result := avl.GetKeys()

	for i, key := range expectedOrder {
		if key != result[i] {
			t.Errorf("Key mismatch: expected: %s, got: %s", key, result[i])
		}
	}
}
