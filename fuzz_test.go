// Use of this source code is governed by the LICENSE file in this module's root
// directory.

// This file exercises parsing and mutation through native Go fuzz tests.
// Arbitrary inputs check parser and setter behavior, while bounded edit
// sequences are compared with an independent ordered model. Each target
// checks list/map consistency and key aliases using state local to its input.

package kargs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// snapshotKargs verifies internal consistency and returns arguments in list
// order. It checks pointers before traversing, so a corrupt list fails instead
// of hanging a fuzz worker.
func snapshotKargs(t *testing.T, k *Kargs) []Karg {
	t.Helper()
	assertInvariants(t, k)
	require.False(t, t.Failed(), "inconsistent Kargs")
	var items []Karg
	for it := k.list; it != nil; it = it.next {
		items = append(items, it.karg)
	}
	return items
}

// checkFuzzLookups verifies that GetKarg and ContainsKarg agree with the ordered
// items for every present key, including its hyphen and underscore aliases.
func checkFuzzLookups(t *testing.T, k *Kargs, items []Karg) {
	t.Helper()
	want := make(map[string][]string)
	for _, item := range items {
		want[item.CanonicalKey] = append(want[item.CanonicalKey], item.Value)
	}
	for key, values := range want {
		for _, alias := range []string{key, strings.ReplaceAll(key, "_", "-")} {
			got, present := k.GetKarg(alias)
			require.True(t, present)
			require.Equal(t, values, got)
			require.True(t, k.ContainsKarg(alias))
		}
	}
}

// FuzzParse checks that arbitrary command line bytes produce consistent state,
// preserve lookup order, and remain stable when serialized and parsed again.
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"", " \t\n", "console=tty0 console=ttyS0 quiet",
		"with-dashes=a with_dashes=b key=a=b =empty-key flag=",
		`key="value with spaces" other='single quotes'`,
		`key="`, `key='`, `key="escaped \" quote" tail`,
		"a\u2003b\u00a0c", "key=“Unicode quotes” tail",
		"\x00 key=\x00", "\xff=\xfe", `"key quotes" key=\\`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		k := NewKargs(input)
		items := snapshotKargs(t, k)
		checkFuzzLookups(t, k, items)
		roundTrip := NewKargs([]byte(k.String()))
		require.Equal(t, items, snapshotKargs(t, roundTrip))
		require.Equal(t, k.String(), roundTrip.String())
	})
}

// FuzzSetKarg exercises arbitrary keys and values in empty and populated lists.
// Invalid keys must leave state unchanged; successful sets must collapse a key
// to one value at its first occurrence without changing unrelated arguments.
func FuzzSetKarg(f *testing.F) {
	for _, seed := range [][2]string{
		{"", ""}, {"target", "new"}, {"target", ""},
		{"with-dashes", "value with spaces"}, {"with_dashes", `'quoted'`},
		{"bad key", "value"}, {"bad\tkey", "value"}, {"bad\nkey", "value"},
		{"x", `"`}, {"y", `"escaped \" quote"`},
		{"key=equals", "a=b"}, {"\xff", "\xfe\x00"}, {"\u2003", "\t\n"},
	} {
		f.Add(seed[0], seed[1])
	}
	f.Fuzz(func(t *testing.T, key, value string) {
		for _, initial := range []string{"", "x=1 target=old with-dashes=a target=older with_dashes=b y=2"} {
			k := NewKargs([]byte(initial))
			before := snapshotKargs(t, k)
			rawBefore := k.String()
			err := k.SetKarg(key, value)
			if strings.ContainsAny(key, " \n\t") {
				require.ErrorIs(t, err, ErrInvalidKey)
				require.Equal(t, before, snapshotKargs(t, k))
				require.Equal(t, rawBefore, k.String())
				continue
			}
			require.NoError(t, err)
			canonical := strings.ReplaceAll(key, "-", "_")
			var wantKeys []string
			inserted := false
			for _, item := range before {
				if item.CanonicalKey == canonical {
					if !inserted {
						wantKeys = append(wantKeys, key)
						inserted = true
					}
				} else {
					wantKeys = append(wantKeys, item.Key)
				}
			}
			if !inserted {
				wantKeys = append(wantKeys, key)
			}
			items := snapshotKargs(t, k)
			var gotKeys []string
			var unchanged []Karg
			var wantUnchanged []Karg
			for _, item := range items {
				gotKeys = append(gotKeys, item.Key)
				if item.CanonicalKey != canonical {
					unchanged = append(unchanged, item)
				}
			}
			for _, item := range before {
				if item.CanonicalKey != canonical {
					wantUnchanged = append(wantUnchanged, item)
				}
			}
			require.Equal(t, wantKeys, gotKeys)
			require.Equal(t, wantUnchanged, unchanged)
			values, present := k.GetKarg(key)
			require.True(t, present)
			require.Equal(t, []string{dequote(value)}, values)
			checkFuzzLookups(t, k, items)
		}
	})
}

// FuzzMutations compares up to 128 append, set, and delete operations with an
// independent ordered-slice model, checking errors, lookups, serialization,
// and internal consistency after each operation.
func FuzzMutations(f *testing.F) {
	// Each triple is (operation, key/alias, value). Operations are append,
	// set, delete-all, delete-first-by-value. Keys 8..15 are underscore aliases
	// of keys 0..7; value 0 is valueless. Trailing partial triples are ignored.
	for _, seed := range [][]byte{
		{}, {0, 0, 1, 0, 8, 1, 0, 0, 2, 1, 8, 3},
		{0, 0, 1, 0, 1, 2, 0, 2, 3, 2, 0, 0, 2, 2, 0, 0, 3, 4},
		{0, 0, 1, 0, 0, 2, 0, 0, 3, 3, 8, 2, 3, 0, 3, 3, 0, 1, 0, 1, 0},
		{2, 0, 0, 3, 0, 1, 1, 0, 1, 1, 0, 0, 3, 8, 0, 0, 0, 2},
		{0, 0, 1, 0, 1, 2, 0, 0, 3, 0, 2, 4, 1, 8, 5, 1, 2, 0},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		k := NewKargsEmpty()
		var model []Karg
		for offset := 0; offset+2 < len(input) && offset < 128*3; offset += 3 {
			op, keyByte, valueByte := input[offset]%4, input[offset+1], input[offset+2]
			canonical := fmt.Sprintf("key_%d", keyByte%8)
			key := canonical
			if keyByte%16 < 8 {
				key = strings.ReplaceAll(key, "_", "-")
			}
			value := ""
			if valueByte != 0 {
				value = fmt.Sprintf("v%d", valueByte)
			}
			raw := key
			if value != "" {
				raw += "=" + value
			}
			item := Karg{Key: key, CanonicalKey: canonical, Value: value, Raw: raw}
			first := -1
			for i, existing := range model {
				if existing.CanonicalKey == canonical && (op == 1 || op == 2 || existing.Value == value) {
					first = i
					break
				}
			}
			switch op {
			case 0:
				k.AppendKargs(raw)
				if first == -1 {
					model = append(model, item)
				}
			case 1:
				require.NoError(t, k.SetKarg(key, value))
				if first == -1 {
					model = append(model, item)
				} else {
					model[first] = item
					for i := len(model) - 1; i > first; i-- {
						if model[i].CanonicalKey == canonical {
							model = append(model[:i], model[i+1:]...)
						}
					}
				}
			case 2, 3:
				var err error
				if op == 2 {
					err = k.DeleteKarg(key)
				} else {
					err = k.DeleteKargByValue(key, value)
				}
				if first == -1 {
					require.ErrorIs(t, err, ErrNotExists)
				} else {
					require.NoError(t, err)
					if op == 3 {
						model = append(model[:first], model[first+1:]...)
					} else {
						for i := len(model) - 1; i >= 0; i-- {
							if model[i].CanonicalKey == canonical {
								model = append(model[:i], model[i+1:]...)
							}
						}
					}
				}
			}
			got := snapshotKargs(t, k)
			require.Equal(t, len(model), len(got), "operation %d", offset/3)
			for i := range model {
				require.Equal(t, model[i], got[i], "operation %d, item %d", offset/3, i)
			}
			var flags []string
			for _, existing := range model {
				flags = append(flags, existing.Raw)
			}
			require.Equal(t, strings.Join(flags, " "), k.String())
			checkFuzzLookups(t, k, got)
			for i := 0; i < 8; i++ {
				lookup := fmt.Sprintf("key_%d", i)
				var want []string
				for _, existing := range model {
					if existing.CanonicalKey == lookup {
						want = append(want, existing.Value)
					}
				}
				values, present := k.GetKarg(lookup)
				require.Equal(t, want, values)
				require.Equal(t, len(want) > 0, present)
				require.Equal(t, present, k.ContainsKarg(lookup))
			}
		}
	})
}
