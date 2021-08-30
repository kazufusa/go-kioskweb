# go-kioskweb

## Description

go-kioskweb opens browsers in kiosk mode.
This package has following features.

1. go-kioskweb supports for IE, Edge, Google Chrome, Firefox
2. **In the startup of Windows 10 Tablet Mode**, it is possible to launch and **display** a kiosk mode browser.

## Usage

```go
// +build windows

package main

import (
	"context"
	"log"
	"time"

	"github.com/kazufusa/go-kioskweb"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := kioskweb.Open(
		"https://github.com",
		kioskweb.Config{Browser: kioskweb.IE, WaitCtx: ctx},
	)
	if err != nil {
		log.Fatal(err)
	}
}
```
