# go-kargs: parse and manipulate kernel command line arguments

<!-- Text width is 80, only use spaces and use 4 spaces instead of tabs -->
<!-- vim: set et sta tw=80 ts=4 sw=4 sts=0: -->

[![Latest release](https://img.shields.io/github/v/release/synackd/go-kargs)](https://github.com/synackd/go-kargs/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/synackd/go-kargs.svg)](https://pkg.go.dev/github.com/synackd/go-kargs)
[![Test](https://github.com/synackd/go-kargs/actions/workflows/test.yml/badge.svg)](https://github.com/synackd/go-kargs/actions/workflows/test.yml)
[![Coverage](https://coveralls.io/repos/github/synackd/go-kargs/badge.svg?branch=main)](https://coveralls.io/github/synackd/go-kargs?branch=main)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/synackd/go-kargs/badge)](https://scorecard.dev/viewer/?uri=github.com/synackd/go-kargs)

<details>
<summary>Additional project checks</summary>

**Quality**

[![Lint](https://github.com/synackd/go-kargs/actions/workflows/lint.yml/badge.svg)](https://github.com/synackd/go-kargs/actions/workflows/lint.yml)
[![Release](https://github.com/synackd/go-kargs/actions/workflows/release.yml/badge.svg)](https://github.com/synackd/go-kargs/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/synackd/go-kargs)](https://goreportcard.com/report/github.com/synackd/go-kargs)

**Security**

[![CodeQL](https://github.com/synackd/go-kargs/actions/workflows/codeql.yaml/badge.svg)](https://github.com/synackd/go-kargs/actions/workflows/codeql.yaml)
[![Vulnerability Check](https://github.com/synackd/go-kargs/actions/workflows/govulncheck.yaml/badge.svg)](https://github.com/synackd/go-kargs/actions/workflows/govulncheck.yaml)

</details>
<br/>

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

## Development

This library targets Go 1.17 and later. A `Makefile` provides the local checks
that mirror the CI workflows (run `make help` for the full list):

```
make test        # run unit tests
make race        # run unit tests with the race detector
make coverage    # run tests and print a coverage summary
make vet         # run go vet
make lint        # run golangci-lint
make govulncheck # run govulncheck
make check       # run all of the above
```

## Documentation

See https://pkg.go.dev/github.com/synackd/go-kargs
