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
	err := kioskweb.OpenKioskWeb("https://github.com", &kioskweb.Config{Browser: kioskweb.IE, WaitCtx: ctx})
	if err != nil {
		log.Fatal(err)
	}
}
