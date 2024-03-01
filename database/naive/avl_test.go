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

	left_height := node.left.avlHeightTestHelper(t)
	right_height := node.right.avlHeightTestHelper(t)

	expected_height := 1 + max(left_height, right_height)

	if absInt32(expected_height-node.getHeight()) != 0 {
		t.Errorf(
			"%s: Difference between expected height: %d, and actual height: %d, is greater than zero! "+
				"Left height: %d, Right height: %d",
			node.key,
			expected_height,
			node.getHeight(),
			left_height,
			right_height,
		)
	}

	return expected_height
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
