package store

import (
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
)

func TestBuildFingerprints(t *testing.T) {
	tests := []struct {
		name     string
		entries  []Kdtree.KDNodeKey
		expected []string
	}{
		{
			name: "Even number of entries",
			entries: []Kdtree.KDNodeKey{
				{Fingerprint: "a"},
				{Fingerprint: "b"},
				{Fingerprint: "c"},
				{Fingerprint: "d"},
			},
			expected: []string{"abcd", "ab", "cd", "a", "b", "c", "d"},
		},
		{
			name: "Odd number of entries",
			entries: []Kdtree.KDNodeKey{
				{Fingerprint: "a"},
				{Fingerprint: "b"},
				{Fingerprint: "c"},
			},
			expected: []string{"abc", "a", "bc", "", "", "b", "c"},
		},
		{
			name: "5 entries",
			entries: []Kdtree.KDNodeKey{
				{Fingerprint: "a"},
				{Fingerprint: "b"},
				{Fingerprint: "c"},
				{Fingerprint: "d"},
				{Fingerprint: "e"},
			},
			expected: []string{"abcde", "ab", "cde", "a", "b", "c", "de", "", "", "", "", "", "", "d", "e"},
		},
		{
			name: "Single entry",
			entries: []Kdtree.KDNodeKey{
				{Fingerprint: "a"},
			},
			expected: []string{"a"},
		},
		{
			name:     "Empty entries",
			entries:  []Kdtree.KDNodeKey{},
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildFingerprints(tt.entries)
			// if len(result) != len(tt.expected) {
			// 	t.Errorf("Expected length %d, got %d, %s", len(tt.expected), len(result), result)
			// }
			for i := range tt.expected {
				if result[i] != tt.expected[i] {
					t.Errorf("Result: %s\nExpected %s at index %d, got %s", result, tt.expected[i], i, result[i])
				}
			}
		})
	}
}
