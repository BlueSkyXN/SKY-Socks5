package main

import (
	"context"
	"fmt"
	"os"

	"github.com/BlueSkyXN/SKY-Socks5/internal/pipeline"
)

func main() {
	if err := pipeline.RunCLI(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
