//go:build !race
// +build !race

package naive

import (
	"sort"
	"testing"
)

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

	keys := []string{"T", "X", "G", "L", "E", "Q", "M", "H", "O", "I", "B", "Z", "A", "V", "S", "R", "K", "P", "C", "D", "U", "F", "N", "W", "Y", "J"}
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
		if expectedKey != string(result[i]) {
			t.Errorf("Key mismatch at index %d; expected: %s, got: %s", i, expectedKey, result[i])
		}
	}

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)
}

func TestAVLDelete(t *testing.T) {
	keys := []string{"30", "10", "50", "6", "18", "45", "58", "5", "8", "15", "44", "65", "7", "9"}
	avl := NewAVL()

	for _, k := range keys {
		avl.Insert(k, nil)
	}

	// Add some duplicates
	avl.Insert("50", nil)
	avl.Insert("45", nil)
	avl.Insert("15", nil)
	avl.Insert("30", nil)

	// Delete where node has two children
	avl.Delete("10")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)

	// Delete where node has right child only
	avl.Delete("58")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)

	// Delete where node has left child only
	avl.Delete("45")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)

	// Delete where node is a leaf
	avl.Delete("7")

	// Check the height of each node
	avl.root.avlHeightTestHelper(t)
}
