package main

import (
	"os"

	"github.com/jiangjiali/vault/command"
)

func main() {
	os.Exit(command.Run(os.Args[1:]))
}
