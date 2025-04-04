// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs_test

import (
	"fmt"

	kargs "github.com/synackd/go-kargs"
)

func ExampleNewKargs() {
	cmdline := `nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1`

	// Parse kernel command line arguments
	k := kargs.NewKargs([]byte(cmdline))
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

func ExampleNewKargsEmpty() {
	k := kargs.NewKargsEmpty()
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

func ExampleKargs_String() {
	cmdline := `nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1`
	k := kargs.NewKargs([]byte(cmdline))
	fmt.Println(k.String())

	// Output:
	// nomodeset root=live:https://example.tld/image.squashfs console=tty0,115200n8 console=ttyS0,115200n8 printk.devkmsg=ratelimit printk.time=1
}

func ExampleKargs_ContainsKarg() {
	cmdline := `key1 key2=val`
	k := kargs.NewKargs([]byte(cmdline))

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

func ExampleKargs_GetKarg() {
	cmdline := `nomodeset console=tty0,115200n8 console=ttyS0,115200n8 root=live:https://example.tld/image.squashfs`
	k := kargs.NewKargs([]byte(cmdline))

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

func ExampleKargs_SetKarg_createReplace() {
	k := kargs.NewKargsEmpty()

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

func ExampleKargs_SetKarg_replaceMultiple() {
	cmdline := `console=tty0,115200n8 console=ttyS0,115200n8`
	k := kargs.NewKargs([]byte(cmdline))

	err := k.SetKarg("console", "tty1,115200n8")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	// Output:
	// console=tty1,115200n8
}

func ExampleKargs_DeleteKarg() {
	k := kargs.NewKargs([]byte("noval key=val"))
	err := k.DeleteKarg("key")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	// Output:
	// noval
}

func ExampleKargs_DeleteKargByValue() {
	cmdline := `key=val1 key=val2 key=val3`
	k := kargs.NewKargs([]byte(cmdline))

	err := k.DeleteKargByValue("key", "val2")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println(k)

	// Output:
	// key=val1 key=val3
}
