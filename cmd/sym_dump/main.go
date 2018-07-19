// The sym_dump tool converts Playstation 1 SYM files to C headers (*.sym ->
// *.h) and scripts for importing symbol information into IDA.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sanctuary/sym"
)

func usage() {
	const use = `
Convert Playstation 1 SYM files to C headers (*.sym -> *.h) and scripts for importing symbol information into IDA.
`
	fmt.Println(use[1:])
	flag.PrintDefaults()
}

func main() {
	// Command line flags.
	var (
		// Output C headers.
		outputC bool
	)
	flag.BoolVar(&outputC, "c", false, "output C headers")
	flag.Usage = usage
	flag.Parse()
	for _, path := range flag.Args() {
		f, err := sym.ParseFile(path)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		switch {
		case outputC:
			// Output C headers.
			p := parse(f)
			dumpTypes(p)
			dumpDecls(p)
		default:
			// Output in Psy-Q DUMPSYM.EXE format.
			fmt.Println(f)
		}
	}
}
