// Use of this source code is governed by the LICENSE file in this module's root
// directory.

// Package kargs implements utility routines for parsing kernel command line
// arguments. This includes parsing a command line raw string (represented as a
// byte slice) into a tokenized structure, getting, setting, and deleting
// command line arguments, and writing the structure back into a raw command
// line string, preserving the original order.
//
// Command line argument format is conformant with
// https://www.kernel.org/doc/html/v4.14/admin-guide/kernel-parameters.html,
// which means that, when using the getter/setter functions, 'var_name' and
// 'var-name' are equivalent (though writing the flag back out will use the
// original key format that was read).
package kargs

import (
	"fmt"
	"strings"
)

type Karg struct {
	CanonicalKey string
	Key          string
	Raw          string
	Value        string
}

func (k Karg) String() string {
	return k.Raw
}

// Kargs provides a way to easily parse through kernel command line arguments
type Kargs struct {
	list      *kargItem              // Linked list of all kargs
	last      *kargItem              // Pointer to last karg in linked list
	keyMap    map[string][]*kargItem // Map of karg key to linked list item for faster reference
	numParams int                    // Total kargs count
}

// NewKargs returns a pointer to a Kargs struct parsed from line.
func NewKargs(line []byte) *Kargs {
	return parse(line)
}

// NewKargsEmpty is like NewKargs, but creates a new Kargs that is empty.
func NewKargsEmpty() *Kargs {
	return NewKargs([]byte{})
}

// ContainsKarg verifies that the kernel command line argument identified by key
// has been set, whether it has a value or not.
func (k *Kargs) ContainsKarg(key string) bool {
	_, present := k.GetKarg(key)
	return present
}

// DeleteKarg deletes all instances of key in the kernel command line argument
// list, returning an error if it was not found or a removal error occurs.
func (k *Kargs) DeleteKarg(key string) error {
	canonicalKey := canonicalizeKey(key)
	if _, exists := k.keyMap[key]; exists {
		for _, ptr := range k.keyMap[canonicalKey] {
			if err := remove(ptr); err != nil {
				return fmt.Errorf("failed to delete key %s with value %s: %w", key, ptr.karg.Value, err)
			} else {
				k.numParams--
			}
		}
		delete(k.keyMap, canonicalKey)
	} else {
		return fmt.Errorf("failed to delete key %s: %w", key, ErrNotExists)
	}

	return nil
}

// DeleteKarByValue only deletes the instance of key that has value of value.
func (k *Kargs) DeleteKargByValue(key, value string) error {
	canonicalKey := canonicalizeKey(key)
	if _, exists := k.keyMap[key]; exists {
		for idx, ptr := range k.keyMap[canonicalKey] {
			if value == ptr.karg.Value {
				if err := remove(ptr); err != nil {
					return fmt.Errorf("failed to delete key %s with value %s: %w", key, ptr.karg.Value, err)
				}
				if len(k.keyMap[canonicalKey]) == 1 {
					k.keyMap[canonicalKey] = []*kargItem{}
				} else if idx == len(k.keyMap[canonicalKey])-1 {
					l := len(k.keyMap[canonicalKey]) - 1
					k.keyMap[canonicalKey] = k.keyMap[canonicalKey][:l-1]
				} else if idx == 0 {
					k.keyMap[canonicalKey] = k.keyMap[canonicalKey][1:]
				} else {
					k.keyMap[canonicalKey] = append(k.keyMap[canonicalKey][:idx], k.keyMap[canonicalKey][(idx+1):]...)
				}
				k.numParams--
				return nil
			}
		}
	} else {
		return fmt.Errorf("failed to delete key %s: %w", key, ErrNotExists)
	}

	return fmt.Errorf("could not find value %s for key %s: %w", value, key, ErrNotExists)
}

// FlagsForModule gets all flags for a designated module and returns them as a
// space-seperated string designed to be passed to insmod. Note that similarly
// to flags, module names with - and _ are treated the same.
func (k *Kargs) FlagsForModule(name string) string {
	var ret string
	flagsAdded := make(map[string]bool) // Ensures duplicate flags aren't both added
	// Module flags come as moduleName.flag in /proc/cmdline
	prefix := canonicalizeKey(name) + "."
	first := true
	for llTracker := k.list; llTracker != nil; llTracker = llTracker.next {
		canonicalFlag := canonicalizeKey(llTracker.karg.Key)
		if !flagsAdded[canonicalFlag] && strings.HasPrefix(canonicalFlag, prefix) {
			flagsAdded[canonicalFlag] = true
			if !first {
				ret += " "
			} else {
				first = false
			}
			// They are passed to insmod space seperated as flag=val
			if llTracker.karg.Value == "" {
				ret += strings.TrimPrefix(canonicalFlag, prefix)
			} else {
				ret += strings.TrimPrefix(canonicalFlag, prefix) + "=" + llTracker.karg.Value
			}
		}
	}
	return ret
}

// GetKarg returns the value list of the karg identified by key, as well as
// whether it was set.
func (k *Kargs) GetKarg(key string) ([]string, bool) {
	canonicalKey := canonicalizeKey(key)
	piPtrs, present := k.keyMap[canonicalKey]
	var vals []string
	for _, p := range piPtrs {
		vals = append(vals, p.karg.Value)
	}
	return vals, present
}

// SetKarg sets key to value.
//
// If the key doesn't exist, it is added. If the key exists, its value is set to
// the new value. If the key exists with multiple values, all of the values are
// removed and the first occurrence of the key has its value set to the new
// value.
func (k *Kargs) SetKarg(key, value string) error {
	if err := checkKey(key); err != nil {
		return fmt.Errorf("key check failed: %w", err)
	}
	canonicalKey := canonicalizeKey(key)
	newKarg := Karg{
		Key:          enquote(key),
		CanonicalKey: canonicalKey,
		Value:        dequote(value),
	}
	if value == "" {
		newKarg.Raw = enquote(key)
	} else {
		newKarg.Raw = fmt.Sprintf("%s=%s", key, enquote(value))
	}
	newKargItem := &kargItem{
		karg: newKarg,
	}
	if ptrList, exists := k.keyMap[canonicalKey]; exists {
		// Karg already exists with one or more values. Set the first
		// value to the new one and remove all of the others.
		for pidx, ptr := range ptrList {
			if ptr == nil {
				continue
			}
			if pidx == 0 {
				if ptr.next == nil {
					k.last = newKargItem
				}
				if ptr.prev == nil {
					k.list = newKargItem
				}
				if err := replace(ptr, newKargItem); err != nil {
					return fmt.Errorf("failed to replace existing karg value: %w", err)
				}
				k.keyMap[canonicalKey][pidx] = newKargItem
				k.keyMap[canonicalKey] = []*kargItem{newKargItem}
			} else {
				if err := remove(ptr); err != nil {
					return fmt.Errorf("failed to remove karg: %w", err)
				}
				k.numParams--
			}
		}
	} else {
		// Karg is new. Append it to the end of the list and set the
		// last pointer to it.
		k.keyMap[canonicalKey] = []*kargItem{newKargItem}
		if k.list == nil {
			k.list = newKargItem
			k.last = k.list
		} else {
			k.last.next = newKargItem
			newKargItem.prev = k.last
			k.last = newKargItem
		}
		k.numParams++
	}

	return nil
}

// String returns the karg list in string form, ready to be used as a kernel
// command line argument string.
func (k *Kargs) String() string {
	var s []string
	for llTracker := k.list; llTracker != nil; llTracker = llTracker.next {
		s = append(s, llTracker.karg.String())
	}
	return strings.Join(s, " ")
}
