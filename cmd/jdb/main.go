package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gmh123521/java-dev-bootstrap/internal/cli"
)

func main() {
	if err := cli.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "错误：", err)
		os.Exit(1)
	}
}
