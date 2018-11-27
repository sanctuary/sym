// The sym_dump tool converts Playstation 1 SYM files to C headers (*.sym ->
// *.h) and scripts for importing symbol information into IDA.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/sanctuary/sym"
	"github.com/sanctuary/sym/csym"
	"github.com/sanctuary/sym/csym/c"
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
		// Output IDA scripts.
		outputIDA bool
		// Merge SYM files.
		merge bool
		// Split output into source files.
		splitSrc bool
		// Output C types.
		outputTypes bool
	)
	flag.BoolVar(&outputC, "c", false, "output C types and declarations")
	flag.StringVar(&outputDir, "dir", dumpDir, "output directory")
	flag.BoolVar(&outputIDA, "ida", false, "output IDA scripts")
	flag.BoolVar(&merge, "merge", false, "merge SYM files")
	flag.BoolVar(&splitSrc, "src", false, "split output into source files")
	flag.BoolVar(&outputTypes, "types", false, "output C types")
	flag.Usage = usage
	flag.Parse()
	if merge && outputIDA {
		log.Fatalf("IDA output not supported in merge mode, as the scripts would be unusable.")
	}

	// Parse SYM files.
	var ps []*csym.Parser
	for _, path := range flag.Args() {
		// Parse SYM file.
		f, err := sym.ParseFile(path)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		switch {
		case outputC, outputIDA:
			// Parse C types and declarations.
			p := csym.NewParser()
			if merge {
				ps = append(ps, p)
			}
			p.ParseTypes(f.Syms)
			p.ParseDecls(f.Syms)
			// Output once for each files if not in merge mode.
			if !merge {
				if err := dump(p, outputDir, outputC, outputTypes, outputIDA, splitSrc, merge); err != nil {
					log.Fatalf("%+v", err)
				}
			}
		case outputTypes:
			// Parse C types.
			p := csym.NewParser()
			if merge {
				ps = append(ps, p)
			}
			p.ParseTypes(f.Syms)
			// Output once for each files if not in merge mode.
			if !merge {
				if err := dump(p, outputDir, outputC, outputTypes, outputIDA, splitSrc, merge); err != nil {
					log.Fatalf("%+v", err)
				}
			}
		default:
			// Output in Psy-Q DUMPSYM.EXE format.
			// Note, we never merge the Psy-Q output.
			fmt.Print(f)
		}
	}
	// Output the merge of all files if in merge mode.
	if merge {
		skipAddrCheck := true
		p := pruneDuplicates(ps, skipAddrCheck)
		if err := dump(p, outputDir, outputC, outputTypes, outputIDA, splitSrc, merge); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// pruneDuplicates prunes duplicates declarations of the parser, optionally
// ignoring differences in address.
func pruneDuplicates(ps []*csym.Parser, skipAddrCheck bool) *csym.Parser {
	dst := csym.NewParser()
	enumPresent := make(map[string]bool)
	structPresent := make(map[string]bool)
	unionPresent := make(map[string]bool)
	typeDefPresent := make(map[string]bool)
	// Add unique predeclared identifiers.
	if def, ok := ps[0].Types["bool"]; ok {
		dst.Types["bool"] = def
	}

	// placeholder type name to make types match even when typename differ.
	const placeholder = "placeholder"
	fakeEnum := 0
	fakeStruct := 0
	fakeUnion := 0
	for pnum, p := range ps {
		// Add unique enums.
		for _, tag := range p.EnumTags {
			t := p.Enums[tag]
			fake := strings.Contains(t.Tag, "fake")
			if fake {
				t.Tag = placeholder
			}
			s := t.Def()
			if fake {
				tag = fmt.Sprintf("enum_fake_%d_%d", pnum, fakeEnum)
				fakeEnum++
				t.Tag = tag
			}
			if !enumPresent[s] {
				if !fake {
					tag = fmt.Sprintf("%s_dup_%d", tag, pnum)
					t.Tag = tag
				}
				dst.Enums[tag] = t
				dst.EnumTags = append(dst.EnumTags, tag)
			}
			enumPresent[s] = true
		}
		// Add unique structs.
		for _, tag := range p.StructTags {
			t := p.Structs[tag]
			fake := strings.Contains(t.Tag, "fake")
			if fake {
				t.Tag = placeholder
			}
			s := t.Def()
			if fake {
				tag = fmt.Sprintf("struct_fake_%d_%d", pnum, fakeStruct)
				fakeStruct++
				t.Tag = tag
			}
			if !structPresent[s] {
				if !fake {
					if _, ok := dst.Structs[tag]; ok {
						tag = fmt.Sprintf("%s_dup_%d", tag, pnum)
						t.Tag = tag
					}
				}
				dst.Structs[tag] = t
				dst.StructTags = append(dst.StructTags, tag)
			}
			structPresent[s] = true
		}
		// Add unique unions.
		for _, tag := range p.UnionTags {
			t := p.Unions[tag]
			fake := strings.Contains(t.Tag, "fake")
			if fake {
				t.Tag = placeholder
			}
			s := t.Def()
			if fake {
				tag = fmt.Sprintf("union_fake_%d_%d", pnum, fakeUnion)
				fakeUnion++
				t.Tag = tag
			}
			if !unionPresent[s] {
				if !fake {
					if _, ok := dst.Unions[tag]; ok {
						tag = fmt.Sprintf("%s_dup_%d", tag, pnum)
						t.Tag = tag
					}
				}
				dst.Unions[tag] = t
				dst.UnionTags = append(dst.UnionTags, tag)
			}
			unionPresent[s] = true
		}
		// Add unique typedefs.
		for _, def := range p.Typedefs {
			s := def.Def()
			if _, ok := typeDefPresent[s]; !ok {
				dst.Typedefs = append(dst.Typedefs, def)
			}
			typeDefPresent[s] = true
		}
	}
	// Sort types by tag.
	sort.Strings(dst.EnumTags)
	sort.Strings(dst.StructTags)
	sort.Strings(dst.UnionTags)
	less := func(i, j int) bool {
		ti := dst.Typedefs[i].(*c.VarDecl)
		tj := dst.Typedefs[j].(*c.VarDecl)
		return ti.Name < tj.Name
	}
	sort.Slice(dst.Typedefs, less)
	return dst
}

// dump dumps the declarations of the parser to the given output directory, in
// the format specified.
func dump(p *csym.Parser, outputDir string, outputC, outputTypes, outputIDA, splitSrc, merge bool) error {
	switch {
	case outputC:
		// Output C types and declarations.
		if err := initOutputDir(outputDir); err != nil {
			return errors.WithStack(err)
		}
		if err := dumpTypes(p, outputDir); err != nil {
			return errors.WithStack(err)
		}
		if splitSrc {
			if err := dumpSourceFiles(p, outputDir); err != nil {
				return errors.WithStack(err)
			}
		} else {
			if err := dumpDecls(p, outputDir); err != nil {
				return errors.WithStack(err)
			}
		}
	case outputTypes:
		// Output C types.
		if err := initOutputDir(outputDir); err != nil {
			return errors.WithStack(err)
		}
		if err := dumpTypes(p, outputDir); err != nil {
			return errors.WithStack(err)
		}
	case outputIDA:
		// Output IDA scripts.
		if err := initOutputDir(outputDir); err != nil {
			return errors.WithStack(err)
		}
		if err := dumpIDAScripts(p, outputDir); err != nil {
			return errors.WithStack(err)
		}
		// Delete bool and __int64 types as they cause issues with IDA.
		delete(p.Types, "bool")
		for i, def := range p.Typedefs {
			if v, ok := def.(*c.VarDecl); ok {
				if v.Name == "__int64" {
					defs := append(p.Typedefs[:i], p.Typedefs[i+1:]...)
					p.Typedefs = defs
					break
				}
			}
		}
		delete(p.Types, "__int64")
		if err := dumpTypes(p, outputDir); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
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
