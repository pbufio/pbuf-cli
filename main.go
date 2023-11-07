package main

import (
	"log"

	"github.com/pbufio/pbuf-cli/cmd"
)

const (
	noFlags  = 0
	noPrefix = ""
)

func main() {
	log.SetFlags(noFlags)
	log.SetPrefix(noPrefix)

	err := cmd.NewRootCmd().Execute()
	if err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
