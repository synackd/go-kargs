// Use of this source code is governed by the LICENSE file in this module's root
// directory.

// This file provides shared structural checks for unit and fuzz tests.
// The checks verify that the ordered linked list, canonical-key index, and
// argument count remain synchronized, and detect cycles before further walks.

package kargs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertInvariants verifies the internal consistency of a Kargs after any
// mutation:
//
//   - The forward linked list (via next) and backward linked list (via prev)
//     describe the same sequence of items.
//   - k.list points to the head and k.last points to the tail (both nil when
//     empty).
//   - numParams equals the number of items in the list.
//   - Every item reachable from the list appears in keyMap under its canonical
//     key, and every keyMap entry points to an item that is in the list.
//   - keyMap holds no empty slices (a key is either present with values or
//     absent).
//   - Each keyMap slice contains exactly its list items, in list order.
//
// Cycles and nil map items fail immediately to keep callers from hanging or
// dereferencing invalid pointers while inspecting the structure.
func assertInvariants(t *testing.T, k *Kargs) {
	t.Helper()
	require.NotNil(t, k)

	// Walk forward, collecting items and verifying prev pointers.
	var forward []*kargItem
	seen := make(map[*kargItem]bool)
	var prev *kargItem
	for it := k.list; it != nil; it = it.next {
		require.False(t, seen[it], "cycle in forward list")
		seen[it] = true
		assert.Same(t, prev, it.prev, "prev pointer mismatch while walking forward")
		forward = append(forward, it)
		prev = it
	}

	// Head/tail pointers.
	if len(forward) == 0 {
		assert.Nil(t, k.list, "list head should be nil when empty")
		assert.Nil(t, k.last, "list tail should be nil when empty")
	} else {
		assert.Same(t, forward[0], k.list, "list head pointer mismatch")
		assert.Same(t, forward[len(forward)-1], k.last, "list tail pointer mismatch")
		assert.Nil(t, forward[0].prev, "head should have nil prev")
		assert.Nil(t, forward[len(forward)-1].next, "tail should have nil next")
	}

	// Walk backward from tail and ensure it mirrors the forward walk.
	var backward []*kargItem
	seen = make(map[*kargItem]bool)
	for it := k.last; it != nil; it = it.prev {
		require.False(t, seen[it], "cycle in backward list")
		seen[it] = true
		backward = append(backward, it)
	}
	require.Len(t, backward, len(forward), "backward and forward walks differ in length")
	for i := range forward {
		assert.Same(t, forward[i], backward[len(backward)-1-i], "backward walk order mismatch at %d", i)
	}

	// numParams matches the list length.
	assert.Equal(t, len(forward), k.numParams, "numParams should equal list length")

	// Every list item is present in keyMap exactly where expected.
	inList := make(map[*kargItem]bool, len(forward))
	wantMap := make(map[string][]*kargItem)
	for _, it := range forward {
		inList[it] = true
		canonical := it.karg.CanonicalKey
		assert.Equal(t, canonicalizeKey(it.karg.Key), canonical)
		wantMap[canonical] = append(wantMap[canonical], it)
	}

	// Every keyMap entry is non-empty and points only to list items.
	total := 0
	for canonical, ptrs := range k.keyMap {
		assert.NotEmptyf(t, ptrs, "keyMap has empty slice for key %q", canonical)
		total += len(ptrs)
		for _, p := range ptrs {
			require.NotNil(t, p, "nil item in keyMap")
			assert.Truef(t, inList[p], "keyMap for %q references item not in list", canonical)
			assert.Equalf(t, canonical, p.karg.CanonicalKey, "keyMap key %q holds item with canonical key %q", canonical, p.karg.CanonicalKey)
		}
	}
	assert.Equal(t, len(forward), total, "total keyMap entries should equal list length")
	require.Len(t, k.keyMap, len(wantMap))
	for key, want := range wantMap {
		got := k.keyMap[key]
		require.Len(t, got, len(want), "keyMap membership for %q", key)
		for i := range want {
			assert.Same(t, want[i], got[i], "keyMap order for %q at %d", key, i)
		}
	}
}
