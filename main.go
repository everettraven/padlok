package main

import (
	"log"

	"github.com/everettraven/padlok/pkg/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
