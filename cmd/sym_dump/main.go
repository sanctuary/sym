package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mewmew/sym"
)

func main() {
	flag.Parse()
	for _, path := range flag.Args() {
		f, err := sym.ParseFile(path)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		fmt.Println(f)
	}
}
