package main

import (
	"akatengu/internal/pkg/enumx_tools"
	"log"
)

func main() {
	if err := enumx_tools.Run(); err != nil {
		log.Fatal(err)
	}
}
