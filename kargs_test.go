// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKargs_AppendKargs_existingVal(t *testing.T) {
	k := NewKargs([]byte(`key=val1 key=val2 key=val3`))

	k.AppendKargs("key=val2")
	assert.Equal(t, 3, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set := k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{"val1", "val2", "val3"}, vals)
}

func TestKargs_AppendKargs_fromEmpty(t *testing.T) {
	k := NewKargsEmpty()

	k.AppendKargs("noval")
	assert.Equal(t, 1, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set := k.GetKarg("noval")
	assert.True(t, set)
	assert.Equal(t, []string{""}, vals)
}

func TestKargs_AppendKargs_fromNonEmpty(t *testing.T) {
	k := NewKargs([]byte("existingarg"))

	k.AppendKargs("noval")
	assert.Equal(t, 2, k.numParams)
	assert.Len(t, k.keyMap, 2)
	vals, set := k.GetKarg("noval")
	assert.True(t, set)
	assert.Equal(t, []string{""}, vals)
}

func TestKargs_AppendKargs_multiple(t *testing.T) {
	k := NewKargs([]byte("key=val1 key=val2"))

	k.AppendKargs("key=val3 key=val4 extra")
	assert.Equal(t, 5, k.numParams)
	assert.Len(t, k.keyMap, 2)
	assert.Equal(t, `key=val1 key=val2 key=val3 key=val4 extra`, k.String())
}

func TestKargs_AppendKargs_novalToVal(t *testing.T) {
	k := NewKargs([]byte(`key`))

	k.AppendKargs("key=val")
	assert.Equal(t, 2, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set := k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{"", "val"}, vals)
}

func TestKargs_ContainsKarg(t *testing.T) {
	k := NewKargs([]byte("test1"))
	assert.True(t, k.ContainsKarg("test1"))
	assert.False(t, k.ContainsKarg("test2"))
}

func TestKargs_DeleteKarg_noValue(t *testing.T) {
	k := NewKargs([]byte("noval key=val"))

	// With no value
	err := k.DeleteKarg("noval")
	assert.NoError(t, err)
	assert.Equal(t, 1, k.numParams)
	assert.Len(t, k.keyMap, 1)
	_, set := k.GetKarg("noval")
	assert.False(t, set)
}

func TestKargs_DeleteKarg_withValue(t *testing.T) {
	k := NewKargs([]byte("noval key=val"))

	// With value
	err := k.DeleteKarg("key")
	assert.NoError(t, err)
	assert.Equal(t, 1, k.numParams)
	assert.Len(t, k.keyMap, 1)
	_, set := k.GetKarg("key")
	assert.False(t, set)
}

func TestKargs_DeleteKarg_nonexistent(t *testing.T) {
	k := NewKargs([]byte("noval key=val"))

	// Test nonexistent
	err := k.DeleteKarg("nonexistent")
	assert.Error(t, err)
}

func TestKargs_DeleteKargByValue_existingValue(t *testing.T) {
	k := NewKargs([]byte("key=val1 key=val2 key=val3"))

	// Test existent value
	err := k.DeleteKargByValue("key", "val2")
	assert.NoError(t, err)
	assert.Equal(t, 2, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set := k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{"val1", "val3"}, vals)
}

func TestKargs_DeleteKargByValue_nonexistentValue(t *testing.T) {
	k := NewKargs([]byte("key=val1 key=val2 key=val3"))

	// Test non-existent value
	err := k.DeleteKargByValue("key", "val4")
	assert.Error(t, err)
}

func TestKargs_DeleteKargByValue_nonexistentKey(t *testing.T) {
	k := NewKargs([]byte("key=val1 key=val2 key=val3"))

	// Test non-existent key
	err := k.DeleteKargByValue("nonexistent", "val")
	assert.Error(t, err)
}

func TestKargs_FlagsForModule_existing(t *testing.T) {
	k := NewKargs([]byte("mod.key1 diffmod diffmod.k1 diffmod.k2=v1 mod.key2=val"))

	// Test existing module kargs
	mods := k.FlagsForModule("mod")
	assert.Equal(t, "key1 key2=val", mods)
}

func TestKargs_FlagsForModule_nonexistent(t *testing.T) {
	k := NewKargs([]byte("mod.key1 diffmod diffmod.k1 diffmod.k2=v1 mod.key2=val"))

	// Test non-existent kargs
	mods := k.FlagsForModule("nonexistent")
	assert.Empty(t, mods)
}

func TestKargs_GetKarg(t *testing.T) {
	k := NewKargs([]byte("noval multkey multkey=val1 multkey=val2 key=val"))

	noval, novalSet := k.GetKarg("noval")
	assert.True(t, novalSet)
	assert.Len(t, noval, 1)
	assert.Empty(t, noval[0])

	keyval, keyvalSet := k.GetKarg("key")
	assert.True(t, keyvalSet)
	assert.Len(t, keyval, 1)
	assert.Equal(t, "val", keyval[0])

	multkey, multkeySet := k.GetKarg("multkey")
	assert.True(t, multkeySet)
	assert.Len(t, multkey, 3)
	assert.Empty(t, multkey[0])
	assert.Equal(t, "val1", multkey[1])
	assert.Equal(t, "val2", multkey[2])
}

func TestKargs_SetKarg_createReplace(t *testing.T) {
	// Test simple creation and replacement
	k := NewKargsEmpty()

	err := k.SetKarg("key", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set := k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{""}, vals)

	err = k.SetKarg("key", "val1")
	assert.NoError(t, err)
	assert.Equal(t, 1, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set = k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{"val1"}, vals)
}

func TestKargs_SetKarg_replaceMultiple(t *testing.T) {
	// Test replacing multiple values
	k := NewKargs([]byte("key=val1 key=val2"))
	assert.Equal(t, 2, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set := k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{"val1", "val2"}, vals)

	err := k.SetKarg("key", "val3")
	assert.NoError(t, err)
	assert.Equal(t, 1, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set = k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{"val3"}, vals)

	// Test unsetting value
	err = k.SetKarg("key", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, k.numParams)
	assert.Len(t, k.keyMap, 1)
	vals, set = k.GetKarg("key")
	assert.True(t, set)
	assert.Equal(t, []string{""}, vals)
}

func TestKargs_String(t *testing.T) {
	cmdline := `nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1`
	k := NewKargs([]byte(cmdline))
	assert.Equal(t, cmdline, k.String())
}

func TestNewKargs(t *testing.T) {
	in := `key1 key2=val`
	k := NewKargs([]byte(in))
	// Since NewKargs calls parseToStruct, more in-depth testing is done for
	// that function. Here, we just make sure the pointer is not nil and
	// that stringifying it matches the input.
	assert.NotNil(t, k)
	assert.Equal(t, in, k.String())
}

func TestNewKargsEmpty(t *testing.T) {
	// Test empty
	emptyK := NewKargsEmpty()
	assert.NotNil(t, emptyK)
	assert.Empty(t, emptyK.numParams)
	assert.Nil(t, emptyK.list)
	assert.Nil(t, emptyK.last)
	assert.Empty(t, emptyK.keyMap)
	assertInvariants(t, emptyK)
}

func TestKargs_AppendKargs_table(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		append     string
		wantString string
		wantKey    string
		wantVals   []string
		wantParams int
	}{
		{
			name:       "append to empty",
			initial:    "",
			append:     `key="val"`,
			wantString: `key="val"`,
			wantKey:    "key",
			wantVals:   []string{"val"},
			wantParams: 1,
		},
		{
			name:       "quoted duplicate not appended",
			initial:    "key=val",
			append:     `key="val"`,
			wantString: "key=val",
			wantKey:    "key",
			wantVals:   []string{"val"},
			wantParams: 1,
		},
		{
			name:       "unquoted duplicate of quoted not appended",
			initial:    `key="val"`,
			append:     "key=val",
			wantString: `key="val"`,
			wantKey:    "key",
			wantVals:   []string{"val"},
			wantParams: 1,
		},
		{
			name:       "different quote styles same content",
			initial:    `key="val with space"`,
			append:     `key='val with space'`,
			wantString: `key="val with space"`,
			wantKey:    "key",
			wantVals:   []string{"val with space"},
			wantParams: 1,
		},
		{
			name:       "different value appended",
			initial:    "key=val",
			append:     "key=val2",
			wantString: "key=val key=val2",
			wantKey:    "key",
			wantVals:   []string{"val", "val2"},
			wantParams: 2,
		},
		{
			name:       "valued appended to valueless",
			initial:    "key",
			append:     "key=val",
			wantString: "key key=val",
			wantKey:    "key",
			wantVals:   []string{"", "val"},
			wantParams: 2,
		},
		{
			name:       "valueless appended to valued",
			initial:    "key=val",
			append:     "key",
			wantString: "key=val key",
			wantKey:    "key",
			wantVals:   []string{"val", ""},
			wantParams: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKargs([]byte(tt.initial))
			k.AppendKargs(tt.append)
			assert.Equal(t, tt.wantString, k.String())
			assert.Equal(t, tt.wantParams, k.numParams)
			vals, set := k.GetKarg(tt.wantKey)
			assert.True(t, set)
			assert.Equal(t, tt.wantVals, vals)
			assertInvariants(t, k)
		})
	}
}

func TestKargs_DeleteKarg_table(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		deleteKey  string
		wantString string
		wantParams int
		wantErr    bool
	}{
		{"remove all instances", "key=a key=b key=c", "key", "", 0, false},
		{"delete head", "key=a b", "key", "b", 1, false},
		{"delete tail", "a key=b", "key", "a", 1, false},
		{"delete middle", "x key=a y", "key", "x y", 2, false},
		{"hyphen alias to canonical", "with-dashes=val other", "with_dashes", "other", 1, false},
		{"underscore alias to hyphen", "with_dashes=val other", "with-dashes", "other", 1, false},
		{"nonexistent", "a b", "c", "a b", 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKargs([]byte(tt.initial))
			err := k.DeleteKarg(tt.deleteKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrNotExists))
			} else {
				require.NoError(t, err)
				assert.False(t, k.ContainsKarg(tt.deleteKey))
			}
			assert.Equal(t, tt.wantString, k.String())
			assert.Equal(t, tt.wantParams, k.numParams)
			assertInvariants(t, k)
		})
	}
}

func TestKargs_DeleteKargByValue_table(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		key        string
		value      string
		wantString string
		wantParams int
		wantErr    bool
		wantSet    bool
		wantVals   []string
	}{
		{
			name: "delete middle value", initial: "key=val1 key=val2 key=val3",
			key: "key", value: "val2",
			wantString: "key=val1 key=val3", wantParams: 2,
			wantSet: true, wantVals: []string{"val1", "val3"},
		},
		{
			name: "delete last value in slice", initial: "key=val1 key=val2",
			key: "key", value: "val2",
			wantString: "key=val1", wantParams: 1,
			wantSet: true, wantVals: []string{"val1"},
		},
		{
			name: "delete first value in slice", initial: "key=val1 key=val2",
			key: "key", value: "val1",
			wantString: "key=val2", wantParams: 1,
			wantSet: true, wantVals: []string{"val2"},
		},
		{
			name: "delete sole value removes key", initial: "key=val other",
			key: "key", value: "val",
			wantString: "other", wantParams: 1,
			wantSet: false,
		},
		{
			name: "delete first of duplicate values", initial: "key=val key=val",
			key: "key", value: "val",
			wantString: "key=val", wantParams: 1,
			wantSet: true, wantVals: []string{"val"},
		},
		{
			name: "delete via hyphen alias", initial: "with-dashes=val other",
			key: "with_dashes", value: "val",
			wantString: "other", wantParams: 1,
			wantSet: false,
		},
		{
			name: "value not present", initial: "key=val1 key=val2",
			key: "key", value: "nope",
			wantString: "key=val1 key=val2", wantParams: 2, wantErr: true,
			wantSet: true, wantVals: []string{"val1", "val2"},
		},
		{
			name: "key not present", initial: "key=val",
			key: "other", value: "val",
			wantString: "key=val", wantParams: 1, wantErr: true,
			wantSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKargs([]byte(tt.initial))
			err := k.DeleteKargByValue(tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrNotExists))
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantString, k.String())
			assert.Equal(t, tt.wantParams, k.numParams)
			vals, set := k.GetKarg(tt.key)
			assert.Equal(t, tt.wantSet, set)
			if tt.wantSet {
				assert.Equal(t, tt.wantVals, vals)
			}
			assertInvariants(t, k)
		})
	}
}

// TestKargs_DeleteKargByValue_thenAppend ensures the tail pointer is not left
// stale after deleting the last item, so a subsequent append links correctly.
func TestKargs_DeleteKargByValue_thenAppend(t *testing.T) {
	k := NewKargs([]byte("a=1 b=2"))
	require.NoError(t, k.DeleteKargByValue("b", "2"))
	assertInvariants(t, k)

	k.AppendKargs("c=3")
	assert.Equal(t, "a=1 c=3", k.String())
	assertInvariants(t, k)
}

func TestKargs_SetKarg_table(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		key        string
		value      string
		wantString string
		wantVals   []string
		wantParams int
		wantErr    bool
	}{
		{
			name: "insert into empty", initial: "",
			key: "key", value: "val",
			wantString: "key=val", wantVals: []string{"val"}, wantParams: 1,
		},
		{
			name: "insert into nonempty appends", initial: "a=1",
			key: "b", value: "2",
			wantString: "a=1 b=2", wantVals: []string{"2"}, wantParams: 2,
		},
		{
			name: "replace sole value", initial: "key=old",
			key: "key", value: "new",
			wantString: "key=new", wantVals: []string{"new"}, wantParams: 1,
		},
		{
			name: "collapse multiple values in place", initial: "x key=a key=b y",
			key: "key", value: "new",
			wantString: "x key=new y", wantVals: []string{"new"}, wantParams: 3,
		},
		{
			name: "set to empty value", initial: "key=a key=b",
			key: "key", value: "",
			wantString: "key", wantVals: []string{""}, wantParams: 1,
		},
		{
			name: "value with spaces is quoted", initial: "key=a",
			key: "key", value: "value with spaces",
			wantString: `key="value with spaces"`, wantVals: []string{"value with spaces"}, wantParams: 1,
		},
		{
			name: "value with embedded equals", initial: "key=a",
			key: "key", value: "a=b",
			wantString: "key=a=b", wantVals: []string{"a=b"}, wantParams: 1,
		},
		{
			name: "invalid key returns error", initial: "key=a",
			key: "bad key", value: "v",
			wantString: "key=a", wantParams: 1, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKargs([]byte(tt.initial))
			err := k.SetKarg(tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidKey))
			} else {
				require.NoError(t, err)
				vals, set := k.GetKarg(tt.key)
				assert.True(t, set)
				assert.Equal(t, tt.wantVals, vals)
			}
			assert.Equal(t, tt.wantString, k.String())
			assert.Equal(t, tt.wantParams, k.numParams)
			assertInvariants(t, k)
		})
	}
}

// TestKargs_SetKarg_headTail exercises replacing the head and the tail so the
// list head/tail pointers are updated correctly, then appends to confirm.
func TestKargs_SetKarg_headTail(t *testing.T) {
	// Replace head.
	k := NewKargs([]byte("key=a b=2"))
	require.NoError(t, k.SetKarg("key", "z"))
	assert.Equal(t, "key=z b=2", k.String())
	assertInvariants(t, k)

	// Replace tail.
	k = NewKargs([]byte("a=1 key=b"))
	require.NoError(t, k.SetKarg("key", "z"))
	assert.Equal(t, "a=1 key=z", k.String())
	k.AppendKargs("c=3")
	assert.Equal(t, "a=1 key=z c=3", k.String())
	assertInvariants(t, k)
}

func TestKargs_FlagsForModule_table(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		module  string
		want    string
	}{
		{"valued and valueless", "mod.key1 diffmod diffmod.k1 diffmod.k2=v1 mod.key2=val", "mod", "key1 key2=val"},
		{"duplicate canonical flags deduped", "mod.foo=1 mod.foo=2", "mod", "foo=1"},
		{"valueless flag", "mod.foo", "mod", "foo"},
		{"hyphen module matches underscore", "mod_name.foo=1", "mod-name", "foo=1"},
		{"underscore module matches hyphen", "mod-name.foo=1", "mod_name", "foo=1"},
		{"prefix does not over-match", "foo.bar=1 fooext.baz=2", "foo", "bar=1"},
		{"nonexistent module", "mod.key1", "nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKargs([]byte(tt.initial))
			assert.Equal(t, tt.want, k.FlagsForModule(tt.module))
		})
	}
}

func TestKargs_GetKarg_missing(t *testing.T) {
	k := NewKargs([]byte("key=val"))
	vals, set := k.GetKarg("missing")
	assert.False(t, set)
	assert.Empty(t, vals)
}

func TestKargs_ContainsKarg_aliases(t *testing.T) {
	k := NewKargs([]byte("with-dashes=val"))
	assert.True(t, k.ContainsKarg("with-dashes"))
	assert.True(t, k.ContainsKarg("with_dashes"))
	assert.False(t, k.ContainsKarg("with_other"))
}

func TestKargs_String_empty(t *testing.T) {
	k := NewKargsEmpty()
	assert.Equal(t, "", k.String())
}
