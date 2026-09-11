// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanonicalizeKey(t *testing.T) {
	checks := [][]string{
		// Input, expected output
		{"", ""},
		{`with-hyphens`, `with_hyphens`},
		{`with_underscores`, `with_underscores`},
		{`with-many-hyphens`, `with_many_hyphens`},
		{`mix-ed_and_mix-ed`, `mix_ed_and_mix_ed`},
		{`-leading`, `_leading`},
		{`trailing-`, `trailing_`},
		{`---`, `___`},
		{`___`, `___`},
	}
	for _, check := range checks {
		in := check[0]
		want := check[1]
		have := canonicalizeKey(in)
		assert.Equal(t, want, have)
	}
}

func TestCheckKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"empty", "", false},
		{"hyphens", "valid-key", false},
		{"underscores", "valid_key", false},
		{"dotted", "module.flag", false},
		{"space", "key with space", true},
		{"tab", "key\twith\ttab", true},
		{"newline", "key\nwith\nnewline", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkKey(tt.key)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidKey))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDequote(t *testing.T) {
	checks := [][]string{
		// Input, expected output
		{``, ``},
		{`"`, `"`},
		{`'`, `'`},
		{`""`, ``},
		{`''`, ``},
		{`no quotes`, `no quotes`},
		{`"ended double quotes"`, `ended double quotes`},
		{`'ended single quotes'`, `ended single quotes`},
		{`"unterminated double`, `"unterminated double`},
		{`unterminated double"`, `unterminated double"`},
		{`"mismatched'`, `"mismatched'`},
		{`\"escaped ended double quotes\"`, `\"escaped ended double quotes\"`},
		{`\'escaped ended single quotes\'`, `\'escaped ended single quotes\'`},
		{`o"bscure double quotes"`, `o"bscure double quotes"`},
		{`o'bscure single quotes'`, `o'bscure single quotes'`},
		// Escape-handling within a quoted string.
		{`"a\"b"`, `a"b`},
		{`"a\\b"`, `a\\b`},
		{`"a\nb"`, `a\nb`},
		{`"escaped \" quote"`, `escaped " quote`},
		{`"double \\\" escape"`, `double \\" escape`},
		{`"trailing\\"`, `trailing\\`},
	}
	for _, check := range checks {
		in := check[0]
		want := check[1]
		have := dequote(in)
		assert.Equal(t, want, have)
	}
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

func TestEnquote(t *testing.T) {
	checks := [][]string{
		// Input, expected output
		{``, ``},
		{`no-spaces-no-quotes`, `no-spaces-no-quotes`},
		{`"no-spaces-double-end-quotes"`, `"no-spaces-double-end-quotes"`},
		{`'no-spaces-single-end-quotes'`, `'no-spaces-single-end-quotes'`},
		{`spaces no quotes`, `"spaces no quotes"`},
		{`"spaces double end quotes"`, `"spaces double end quotes"`},
		{`'spaces single end quotes'`, `'spaces single end quotes'`},
		{`spaces" obscure double quotes"`, `"spaces\" obscure double quotes\""`},
		{`spaces' obscure single quotes'`, `"spaces' obscure single quotes'"`},
	}
	for _, check := range checks {
		in := check[0]
		want := check[1]
		have := enquote(in)
		assert.Equal(t, want, have)
	}
}

func TestDoParse_empty(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"spaces", "     "},
		{"tabs", "\t\t"},
		{"mixed whitespace", " \t\n "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			doParse(tt.in, func(flag, key, canonicalKey, value, trimmedValue string) {
				called = true
			})
			assert.False(t, called, "handler should not be called for whitespace-only input")
		})
	}
}

func TestDoParse_tokens(t *testing.T) {
	type token struct {
		flag, key, canonicalKey, value, trimmedValue string
	}
	tests := []struct {
		name string
		in   string
		want []token
	}{
		{
			name: "valueless and valued",
			in:   "noval key=val",
			want: []token{
				{"noval", "noval", "noval", "", ""},
				{"key=val", "key", "key", "val", "val"},
			},
		},
		{
			name: "empty value after equals",
			in:   "key=",
			want: []token{
				{"key=", "key", "key", "", ""},
			},
		},
		{
			name: "embedded equals in value",
			in:   "key=a=b=c",
			want: []token{
				{"key=a=b=c", "key", "key", "a=b=c", "a=b=c"},
			},
		},
		{
			name: "collapses consecutive separators",
			in:   "  a    b  ",
			want: []token{
				{"a", "a", "a", "", ""},
				{"b", "b", "b", "", ""},
			},
		},
		{
			name: "quoted value with spaces stays one token",
			in:   `key="value with spaces"`,
			want: []token{
				{`key="value with spaces"`, "key", "key", `"value with spaces"`, "value with spaces"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []token
			doParse(tt.in, func(flag, key, canonicalKey, value, trimmedValue string) {
				got = append(got, token{flag, key, canonicalKey, value, trimmedValue})
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseToStruct_emptyAndWhitespace(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"spaces", "   "},
		{"tabs", "\t\t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := parseToStruct(tt.in)
			require.NotNil(t, k)
			assert.Equal(t, 0, k.numParams)
			assert.Nil(t, k.list)
			assert.Nil(t, k.last)
			assert.Empty(t, k.keyMap)
			assertInvariants(t, k)
		})
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
	for km := range k.keyMap {
		keyLen, exists := expKeyLens[km]
		assert.True(t, exists)
		assert.Len(t, k.keyMap[km], keyLen)
	}

	// Make sure there aren't any extra keys in key map
	for km := range expKeyLens {
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
