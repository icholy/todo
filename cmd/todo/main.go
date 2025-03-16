package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/icholy/todo"
)

func main() {
	flag.Parse()
	for _, filename := range flag.Args() {
		source, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		todos, err := todo.Parse(filename, source)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range todos {
			fmt.Printf("%s %s\n", t.Location, t)
		}
	}
}
