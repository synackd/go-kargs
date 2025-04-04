// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalizeKey(t *testing.T) {
	checks := [][]string{
		// Input, expected output
		[]string{`with-hyphens`, `with_hyphens`},
		[]string{`with_underscores`, `with_underscores`},
	}
	for _, check := range checks {
		in := check[0]
		want := check[1]
		have := canonicalizeKey(in)
		assert.Equal(t, have, want)
	}
}

func TestDequote(t *testing.T) {
	checks := [][]string{
		// Input, expected output
		[]string{`no quotes`, `no quotes`},
		[]string{`"ended double quotes"`, `ended double quotes`},
		[]string{`'ended single quotes'`, `ended single quotes`},
		[]string{`\"escaped ended double quotes\"`, `\"escaped ended double quotes\"`},
		[]string{`\'escaped ended single quotes\'`, `\'escaped ended single quotes\'`},
		[]string{`o"bscure double quotes"`, `o"bscure double quotes"`},
		[]string{`o'bscure single quotes'`, `o'bscure single quotes'`},
	}
	for _, check := range checks {
		in := check[0]
		want := check[1]
		have := dequote(in)
		assert.Equal(t, have, want)
	}
}

func TestEnquote(t *testing.T) {
	checks := [][]string{
		// Input, expected output
		[]string{`no-spaces-no-quotes`, `no-spaces-no-quotes`},
		[]string{`"no-spaces-double-end-quotes"`, `"no-spaces-double-end-quotes"`},
		[]string{`'no-spaces-single-end-quotes'`, `'no-spaces-single-end-quotes'`},
		[]string{`spaces no quotes`, `"spaces no quotes"`},
		[]string{`"spaces double end quotes"`, `"spaces double end quotes"`},
		[]string{`'spaces single end quotes'`, `'spaces single end quotes'`},
		[]string{`spaces" obscure double quotes"`, `"spaces\" obscure double quotes\""`},
		[]string{`spaces' obscure single quotes'`, `"spaces' obscure single quotes'"`},
	}
	for _, check := range checks {
		in := check[0]
		want := check[1]
		have := enquote(in)
		assert.Equal(t, have, want)
	}
}

func TestParseToStruct(t *testing.T) {
	in := `noval dup=val1 dup=val2 nondup=val with-dashes with-dashes-val=val`
	expNumKargs := 6
	expNumKeys := 5

	// Order matters
	expKargs := []Karg{
		{CanonicalKey: "noval", Key: "noval", Raw: "noval", Value: ""},
		{CanonicalKey: "dup", Key: "dup", Raw: "dup=val1", Value: "val1"},
		{CanonicalKey: "dup", Key: "dup", Raw: "dup=val2", Value: "val2"},
		{CanonicalKey: "nondup", Key: "nondup", Raw: "nondup=val", Value: "val"},
		{CanonicalKey: "with_dashes", Key: "with-dashes", Raw: "with-dashes", Value: ""},
		{CanonicalKey: "with_dashes_val", Key: "with-dashes-val", Raw: "with-dashes-val=val", Value: "val"},
	}
	// Maps key to expected number of values for the key
	expKeyLens := map[string]int{
		"noval":           1,
		"dup":             2,
		"nondup":          1,
		"with_dashes":     1,
		"with_dashes_val": 1,
	}

	k := parseToStruct(in)

	// Make sure struct is not nil
	assert.NotNil(t, k)

	// Make sure number of kargs matches count in 'in' string
	assert.Equal(t, k.numParams, expNumKargs)

	// Make sure key map has expected number of keys
	assert.Len(t, k.keyMap, expNumKeys)

	// Make sure present keys in key map are expected and have expected number
	// of values
	for km, _ := range k.keyMap {
		keyLen, exists := expKeyLens[km]
		assert.True(t, exists)
		assert.Len(t, k.keyMap[km], keyLen)
	}

	// Make sure there aren't any extra keys in key map
	for km, _ := range expKeyLens {
		_, exists := k.keyMap[km]
		assert.True(t, exists)
	}

	// Make sure linked list is structured as expected
	var last *kargItem
	for i, llTracker := 0, k.list; llTracker != nil; i, last, llTracker = i+1, llTracker, llTracker.next {
		assert.Equal(t, llTracker.karg, expKargs[i])
	}
	// Make sure last pointer in linked list actually points to last item
	assert.Equal(t, last, k.last)
}

func TestDoParse(t *testing.T) {
	in := `noval dup=val1 dup=val2 nondup=val with-dashes with-dashes-val=val "key quotes" \"key escaped quotes\" vq="value quotes" veq=\"value escaped quotes\"`
	expKargs := []Karg{
		{CanonicalKey: "noval", Key: "noval", Raw: "noval", Value: ""},
		{CanonicalKey: "dup", Key: "dup", Raw: "dup=val1", Value: "val1"},
		{CanonicalKey: "dup", Key: "dup", Raw: "dup=val2", Value: "val2"},
		{CanonicalKey: "nondup", Key: "nondup", Raw: "nondup=val", Value: "val"},
		{CanonicalKey: "with_dashes", Key: "with-dashes", Raw: "with-dashes", Value: ""},
		{CanonicalKey: "with_dashes_val", Key: "with-dashes-val", Raw: "with-dashes-val=val", Value: "val"},
		{CanonicalKey: `"key quotes"`, Key: `"key quotes"`, Raw: `"key quotes"`, Value: ""},
		{CanonicalKey: `\"key escaped quotes\"`, Key: `\"key escaped quotes\"`, Raw: `\"key escaped quotes\"`, Value: ""},
		{CanonicalKey: "vq", Key: "vq", Raw: `vq="value quotes"`, Value: `"value quotes"`},
		{CanonicalKey: "veq", Key: "veq", Raw: `veq=\"value escaped quotes\"`, Value: `\"value escaped quotes\"`},
	}
	idx := 0
	doParse(in, func(flag, key, canonicalKey, value, trimmedValue string) {
		assert.Equal(t, expKargs[idx].Raw, flag, "raw values mismatch")
		assert.Equal(t, expKargs[idx].Key, key, "keys mismatch")
		assert.Equal(t, expKargs[idx].CanonicalKey, canonicalKey, "canonical keys mismatch")
		assert.Equal(t, expKargs[idx].Value, value, "values mismatch")
		idx++
	})
}
