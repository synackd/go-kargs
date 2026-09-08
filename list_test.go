// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildList constructs a Kargs whose linked list and keyMap contain one item
// per provided key, in order. Each item's value is empty. It is used to
// exercise the unlink/substitute helpers directly.
func buildList(keys ...string) (*Kargs, []*kargItem) {
	k := NewKargsEmpty()
	items := make([]*kargItem, 0, len(keys))
	for _, key := range keys {
		canonical := canonicalizeKey(key)
		it := &kargItem{karg: Karg{Key: key, CanonicalKey: canonical, Raw: key}}
		if k.list == nil {
			k.list = it
			k.last = it
		} else {
			it.prev = k.last
			k.last.next = it
			k.last = it
		}
		k.keyMap[canonical] = append(k.keyMap[canonical], it)
		k.numParams++
		items = append(items, it)
	}
	return k, items
}

func TestKargs_unlink_nil(t *testing.T) {
	k := NewKargsEmpty()
	err := k.unlink(nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNilPtr))
}

func TestKargs_unlink_positions(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string
		removeAt int
		wantSeq  []string
	}{
		{"sole", []string{"a"}, 0, nil},
		{"head", []string{"a", "b", "c"}, 0, []string{"b", "c"}},
		{"middle", []string{"a", "b", "c"}, 1, []string{"a", "c"}},
		{"tail", []string{"a", "b", "c"}, 2, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, items := buildList(tt.keys...)
			target := items[tt.removeAt]

			require.NoError(t, k.unlink(target))
			// Keep bookkeeping consistent so invariants hold.
			canonical := target.karg.CanonicalKey
			delete(k.keyMap, canonical)
			k.numParams--

			assert.Nil(t, target.prev)
			assert.Nil(t, target.next)

			var seq []string
			for it := k.list; it != nil; it = it.next {
				seq = append(seq, it.karg.Key)
			}
			assert.Equal(t, tt.wantSeq, seq)
			assertInvariants(t, k)
		})
	}
}

func TestKargs_substitute_nil(t *testing.T) {
	k, items := buildList("a")
	newItem := &kargItem{karg: Karg{Key: "x", CanonicalKey: "x", Raw: "x"}}

	err := k.substitute(nil, newItem)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNilPtr))

	err = k.substitute(items[0], nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNilPtr))
}

func TestKargs_substitute_positions(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		at      int
		wantSeq []string
	}{
		{"sole", []string{"a"}, 0, []string{"x"}},
		{"head", []string{"a", "b", "c"}, 0, []string{"x", "b", "c"}},
		{"middle", []string{"a", "b", "c"}, 1, []string{"a", "x", "c"}},
		{"tail", []string{"a", "b", "c"}, 2, []string{"a", "b", "x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, items := buildList(tt.keys...)
			old := items[tt.at]
			newItem := &kargItem{karg: Karg{Key: "x", CanonicalKey: "x", Raw: "x"}}

			require.NoError(t, k.substitute(old, newItem))
			// Keep bookkeeping consistent so invariants hold.
			k.keyMap[old.karg.CanonicalKey] = nil
			delete(k.keyMap, old.karg.CanonicalKey)
			k.keyMap["x"] = append(k.keyMap["x"], newItem)

			assert.Nil(t, old.prev)
			assert.Nil(t, old.next)

			var seq []string
			for it := k.list; it != nil; it = it.next {
				seq = append(seq, it.karg.Key)
			}
			assert.Equal(t, tt.wantSeq, seq)
			assertInvariants(t, k)
		})
	}
}

func TestKarg_String(t *testing.T) {
	karg := Karg{Key: "key", CanonicalKey: "key", Raw: "key=val", Value: "val"}
	assert.Equal(t, "key=val", karg.String())
}
