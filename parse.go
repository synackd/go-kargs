// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package kargs

import (
	"fmt"
	"strings"
	"unicode"
)

// Kernel variables must allow '-' and '_' to be equivalent in variable names.
// The canonicalized key will replace '-' with '_' in the keys.
func canonicalizeKey(key string) string {
	return strings.Replace(key, "-", "_", -1)
}

// checkKey checks the given unquoted key for invalid characters and errs if any
// are present. Invalid characters are such whitespace characters as spaces,
// tabs, and newlines.
func checkKey(key string) error {
	invalidKeyChars := " \n\t"
	if strings.ContainsAny(key, invalidKeyChars) {
		return fmt.Errorf("checking key %s: %w", key, ErrInvalidKey)
	}
	return nil
}

// dequote removes single and double quotes that aren't escaped with a
// backslash.
func dequote(line string) string {
	if len(line) == 0 {
		return line
	}

	quotationMarks := `"'`

	var quote byte
	if strings.ContainsAny(string(line[0]), quotationMarks) {
		quote = line[0]
		line = line[1 : len(line)-1]
	}

	var context []byte
	var newLine []byte
	for _, c := range []byte(line) {
		if c == '\\' {
			context = append(context, c)
		} else if c == quote {
			if len(context) > 0 {
				last := context[len(context)-1]
				if last == c {
					context = context[:len(context)-1]
				} else if last == '\\' {
					// Delete one level of backslash
					newLine = newLine[:len(newLine)-1]
					context = []byte{}
				}
			} else {
				context = append(context, c)
			}
		} else if len(context) > 0 && context[len(context)-1] == '\\' {
			// If backslash is being used to escape something other
			// than the quote, ignore it.
			context = []byte{}
		}

		newLine = append(newLine, c)
	}
	return string(newLine)
}

// doParse is a generic parsing function that tokenizes input by spaces,
// honoring quotes (meaning that quoted strings are not split if they have
// spaces). It separates each token into the raw token (flag), the key (left of
// =), the canonicalized key (hyphens turned into underscores), the value (right
// of =), and the trimmedValue (dequoted value). These values are passed to the
// handler function, which is executed for each token.
func doParse(input string, handler func(flag, key, canonicalKey, value, trimmedValue string)) {
	lastQuote := rune(0)
	quotedFieldsCheck := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)
		}
	}

	for _, flag := range strings.FieldsFunc(string(input), quotedFieldsCheck) {
		// Split the flag into a key and value
		split := strings.Index(flag, "=")

		if len(flag) == 0 {
			continue
		}
		var key, value string
		if split == -1 {
			// If no value, leave the flag without =
			key = flag
		} else {
			// If value, flag is key=value
			key = flag[:split]
			value = flag[split+1:]
		}
		canonicalKey := canonicalizeKey(key)
		trimmedValue := dequote(value)

		// Call the passed handler for each token
		handler(flag, key, canonicalKey, value, trimmedValue)
	}
}

// enquote surrounds a string in double quotes if it contains spaces and isn't
// already surrounded by single or double quotes.
func enquote(line string) string {
	quotationMarks := `"'`
	if strings.ContainsAny(line, ` `) {
		if strings.ContainsAny(string(line[0]), quotationMarks) && strings.ContainsAny(string(line[len(line)-1]), quotationMarks) {
			return line
		} else {
			return fmt.Sprintf("%q", line)
		}
	} else {
		return line
	}
}

// parse parses the raw byte slice into a Kargs struct and returns a pointer
// to it.
func parse(raw []byte) *Kargs {
	return parseToStruct(string(raw))
}

// parseToStruct takes a kernel command line string and parses it into a Kargs
// struct, whose pointer is returned.
func parseToStruct(input string) *Kargs {
	var (
		last      *kargItem
		ll        *kargItem
		llTracker     = ll
		numParams int = 0
	)
	keyMap := make(map[string][]*kargItem)
	doParse(input, func(flag, key, canonicalKey, value, trimmedValue string) {
		newKarg := Karg{
			CanonicalKey: canonicalKey,
			Key:          key,
			Raw:          flag,
			Value:        trimmedValue,
		}
		newKargItem := &kargItem{
			karg: newKarg,
		}
		if llTracker == nil {
			// Linked list is empty, create first item
			ll = newKargItem
			llTracker = ll
		} else {
			// Linked list is nonempty, append item and set
			// prev/next pointers
			newKargItem.prev = llTracker
			llTracker.next = newKargItem
			llTracker = llTracker.next
		}
		numParams++
		keyMap[canonicalKey] = append(keyMap[canonicalKey], newKargItem)
		last = newKargItem
	})
	return &Kargs{
		last:      last,
		list:      ll,
		keyMap:    keyMap,
		numParams: numParams,
	}
}
