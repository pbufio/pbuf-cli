package main

import (
	"log"

	"github.com/pbufio/pbuf-cli/cmd"
)

func main() {
	err := cmd.NewRootCmd().Execute()
	if err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
