package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	todo "github.com/icholy/todo"
)

type TodoJSON struct {
	Location    string
	Line        string
	Description string
	Attributes  map[string]string
}

func main() {
	flag.Parse()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	for _, filename := range flag.Args() {
		source, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		todos, err := todo.Parse(ctx, filename, source, nil)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range todos {
			todojson := TodoJSON{
				Location:    fmt.Sprintf("%s:%d", t.Location.File, t.Location.Line),
				Line:        t.Line,
				Description: t.Description,
				Attributes:  map[string]string{},
			}
			for _, a := range t.Attributes {
				todojson.Attributes[a.Key] = a.Value
			}
			if err := enc.Encode(todojson); err != nil {
				log.Fatal(err)
			}
		}
	}
}
