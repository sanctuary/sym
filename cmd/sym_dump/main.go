// The sym_dump tool converts Playstation 1 SYM files to C headers (*.sym ->
// *.h) and scripts for importing symbol information into IDA.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/sanctuary/sym"
)

// usage prints usage information.
func usage() {
	const use = `
Convert Playstation 1 SYM files to C headers (*.sym -> *.h) and scripts for importing symbol information into IDA.
`
	fmt.Println(use[1:])
	flag.PrintDefaults()
}

// Default output directory.
const dumpDir = "_dump_"

func main() {
	// Command line flags.
	var (
		// Output C types and declarations.
		outputC bool
		// Output directory.
		outputDir string
		// Output C types.
		outputTypes bool
	)
	flag.BoolVar(&outputC, "c", false, "output C types and declarations")
	flag.StringVar(&outputDir, "dir", dumpDir, "output directory")
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
			if err := initOutputDir(outputDir); err != nil {
				log.Fatalf("%+v", errors.WithStack(err))
			}
			if err := dumpTypes(p, outputDir); err != nil {
				log.Fatalf("%+v", err)
			}
			if err := dumpDecls(p, outputDir); err != nil {
				log.Fatalf("%+v", err)
			}
		case outputTypes:
			// Output C types.
			p := newParser()
			p.parseTypes(f.Syms)
			if err := initOutputDir(outputDir); err != nil {
				log.Fatalf("%+v", errors.WithStack(err))
			}
			if err := dumpTypes(p, outputDir); err != nil {
				log.Fatalf("%+v", err)
			}
		default:
			// Output in Psy-Q DUMPSYM.EXE format.
			fmt.Print(f)
		}
	}
}

// initOutputDir initializes the output directory.
func initOutputDir(outputDir string) error {
	// Only remove output directory if set to default. Otherwise, let user remove
	// output directory as a safety precaution.
	if outputDir == dumpDir {
		if err := os.RemoveAll(outputDir); err != nil {
			return errors.WithStack(err)
		}
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
