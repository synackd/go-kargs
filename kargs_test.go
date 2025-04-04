// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKargs(t *testing.T) {
	in := `key1 key2=val`
	k := NewKargs([]byte(in))
	// Since NewKargs calls parseToStruct, more in-depth testing is done for
	// that function. Here, we just make sure the pointer is not nil and
	// that stringifying it matches the input.
	assert.NotNil(t, k)
	assert.Equal(t, in, k.String())
}

func ExampleNewKargs() {
	cmdline := `nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1`

	// Parse kernel command line arguments
	k := NewKargs([]byte(cmdline))
	fmt.Println(k)

	// Get values
	consoleVals, consoleSet := k.GetKarg("console")
	fmt.Printf("console set: %v; values: %v\n", consoleSet, consoleVals)

	// Get module flags
	modvals := k.FlagsForModule("printk")
	fmt.Printf("printk module values: %v\n", modvals)

	// Output:
	// nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1
	// console set: true; values: [tty0,115200n8 ttyS0,115200n8]
	// printk module values: devkmsg=ratelimit time=1
}

func TestNewKargsEmpty(t *testing.T) {
	// Test empty
	emptyK := NewKargsEmpty()
	assert.NotNil(t, emptyK)
	assert.Empty(t, emptyK.numParams)
	assert.Nil(t, emptyK.list)
	assert.Nil(t, emptyK.last)
	assert.Empty(t, emptyK.keyMap)
}

func ExampleNewKargsEmpty() {
	k := NewKargsEmpty()
	fmt.Printf("%q\n", k)

	err := k.SetKarg("console", "tty0,115200n8")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf("%q\n", k)

	// Output:
	// ""
	// "console=tty0,115200n8"
}

func TestKargs_String(t *testing.T) {
	cmdline := `nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1`
	k := NewKargs([]byte(cmdline))
	assert.Equal(t, cmdline, k.String())
}

func ExampleKargs_String() {
	cmdline := `nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1`
	k := NewKargs([]byte(cmdline))
	fmt.Println(k.String())

	// Output:
	// nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1
}

func TestKargs_ContainsKarg(t *testing.T) {
	k := NewKargs([]byte("test1"))
	assert.True(t, k.ContainsKarg("test1"))
	assert.False(t, k.ContainsKarg("test2"))
}

func ExampleKargs_ContainsKarg() {
	cmdline := `key1 key2=val`
	k := NewKargs([]byte(cmdline))

	kList := []struct {
		key    string
		exists bool
	}{
		{key: "key1", exists: k.ContainsKarg("key1")},
		{key: "key2", exists: k.ContainsKarg("key2")},
		{key: "key3", exists: k.ContainsKarg("key3")},
	}
	for _, v := range kList {
		fmt.Printf("contains %s: %v\n", v.key, v.exists)
	}

	// Output:
	// contains key1: true
	// contains key2: true
	// contains key3: false
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

func ExampleKargs_GetKarg() {
	cmdline := `nomodeset console=tty0,115200n8 console=ttyS0,115200n8 root=live:https://example.tld/image.squashfs`
	k := NewKargs([]byte(cmdline))

	// Get all values of console
	console, _ := k.GetKarg("console")
	fmt.Printf("console: %v\n", console)

	// Get value of single key with a value
	root, _ := k.GetKarg("root")
	fmt.Printf("root: %v\n", root)

	// Get value of single key with no value
	nomodeset, _ := k.GetKarg("nomodeset")
	fmt.Printf("nomodeset: %v\n", nomodeset)

	// Output:
	// console: [tty0,115200n8 ttyS0,115200n8]
	// root: [live:https://example.tld/image.squashfs]
	// nomodeset: []
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

func ExampleKargs_SetKarg_createReplace() {
	k := NewKargsEmpty()

	err := k.SetKarg("key", "")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	err = k.SetKarg("key", "val")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	// Output:
	// key
	// key=val
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

func ExampleKargs_SetKarg_replaceMultiple() {
	cmdline := `console=tty0,115200n8 console=ttyS0,115200n8`
	k := NewKargs([]byte(cmdline))

	err := k.SetKarg("console", "tty1,115200n8")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	// Output:
	// console=tty1,115200n8
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

func ExampleKargs_DeleteKarg() {
	k := NewKargs([]byte("noval key=val"))
	err := k.DeleteKarg("key")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	// Output:
	// noval
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

func ExampleKargs_DeleteKargByValue() {
	cmdline := `key=val1 key=val2 key=val3`
	k := NewKargs([]byte(cmdline))

	err := k.DeleteKargByValue("key", "val2")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	// Output:
	// key=val1 key=val3
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

func ExampleKargs_FlagsForModule() {
	cmdline := `nomodeset printk.devkmsg=ratelimit printk.time=1`
	k := NewKargs([]byte(cmdline))

	fmt.Println(k.FlagsForModule("printk"))

	// Output:
	// devkmsg=ratelimit time=1
}
