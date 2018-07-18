// The cel_dump tool converts Playstation 1 SYM files to C headers
// (*.sym -> *.h).
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sanctuary/sym"
)

func main() {
	// Command line flags.
	var (
		// Output C headers.
		outputC bool
	)
	flag.BoolVar(&outputC, "c", false, "output C headers")
	flag.Parse()
	for _, path := range flag.Args() {
		f, err := sym.ParseFile(path)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		switch {
		case outputC:
			// Output C headers.
			if err := dumpC(f); err != nil {
				log.Fatalf("%+v", err)
			}
		default:
			// Output in Psy-Q DUMPSYM.EXE format.
			fmt.Println(f)
		}
	}
}
