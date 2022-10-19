package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/buglloc/cf-ddns/internal/commands"
)

func main() {
	runtime.GOMAXPROCS(1)

	if err := commands.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
