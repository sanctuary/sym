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
		// Output C types and declarations.
		outputC bool
		// Output C types.
		outputTypes bool
	)
	flag.BoolVar(&outputC, "c", false, "output C types and declarations")
	flag.BoolVar(&outputTypes, "types", false, "output C types")
	flag.Usage = usage
	flag.Parse()
	for _, path := range flag.Args() {
		f, err := sym.ParseFile(path)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		switch {
		case outputC:
			// Output C types and declarations.
			p := newParser()
			p.parseTypes(f.Syms)
			p.parseDecls(f.Syms)
			dumpTypes(p)
			dumpDecls(p)
		case outputTypes:
			// Output C types.
			p := newParser()
			p.parseTypes(f.Syms)
			dumpTypes(p)
		default:
			// Output in Psy-Q DUMPSYM.EXE format.
			fmt.Println(f)
		}
	}
}
