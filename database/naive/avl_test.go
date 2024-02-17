package naive

import (
	"math/rand"
	"sort"
	"testing"
)

func TestAVL(t *testing.T) {

	keys := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	avl := NewAVL()

	// Shuffle the keys randomly
	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	for _, k := range(keys) {
		avl.Insert([]byte(k), nil)
	}

	result := []string{}
	avl.getKeys(&result)

	sort.Strings(keys)

	if len(keys) != len(result) {
		t.Errorf("Expected the same length: %d, got %d", len(keys), len(result))
	}

	for i, k := range keys {
		if k != result[i] {
			t.Errorf("Expected: %d, got %d", len(keys), len(result))
		}
	}
}