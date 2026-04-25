package main

import (
	"akatengu/internal/pkg/dbmap_tools"
	"log"
)

func main() {
	if err := dbmap_tools.Run(); err != nil {
		log.Fatal(err)
	}
}