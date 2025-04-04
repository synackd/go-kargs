# go-kargs: parse and manipulate kernel command line arguments

<!-- Text width is 80, only use spaces and use 4 spaces instead of tabs -->
<!-- vim: set et sta tw=80 ts=4 sw=4 sts=0: -->

Read, set, delete, then write back out kernel command line arguments.

```go
package main

import (
	"fmt"

	kargs "github.com/synackd/go-kargs"
)

func main() {
	// Parse kernel command line arguments
	kargsIn := `nomodeset root=live:https://172.16.0.254/boot-images/compute/base/test console=tty0,115200 console=ttyS0,115200 printk.devkmsg=ratelimit printk.time=1`
	k := kargs.NewKargs([]byte(kargsIn))
	fmt.Println(k)

	// Get all values for an argument
	valConsole, isSetConsole := k.GetKarg("console")
	if isSetConsole {
		fmt.Printf("console: %v\n", valConsole)
	} else {
		fmt.Println("console not set")
	}

	// Works even with a single value
	valRoot, isSetRoot := k.GetKarg("root")
	if isSetRoot {
		fmt.Printf("root: %v\n", valRoot)
	} else {
		fmt.Println("root not set")
	}

	// Get all arguments for a module
	fmt.Println("args for printk " + k.FlagsForModule("printk"))

	// Override multiple args
	if err := k.SetKarg("console", "ttyS1,155200n8"); err != nil {
		fmt.Println("params with new console settings: " + k.String())
	}
}
```

Output:

```
nomodeset root=live:https://172.16.0.254/boot-images/compute/base/test console=tty0,115200 console=ttyS0,115200 printk.devkmsg=ratelimit printk.time=1
console: [tty0,115200 ttyS0,115200]
root: [live:https://172.16.0.254/boot-images/compute/base/test]
args for printk devkmsg=ratelimit time=1
```

## Installation

```
go get github.com/synackd/go-kargs
```

Import as `kargs "github.com/synackd/go-kargs"`

## Documentation

See https://pkg.go.dev/github.com/synackd/go-kargs
