package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/sanctuary/sym/internal/c"
)

// --- [ Type definitions ] ----------------------------------------------------

// Type definitions header file name.
const typesName = "types.h"

// dumpTypes outputs the type information recorded by the parser to a C header
// stored in the output directory.
func dumpTypes(p *parser, outputDir string) error {
	// Create output file.
	typesPath := filepath.Join(outputDir, typesName)
	fmt.Println("creating:", typesPath)
	f, err := os.Create(typesPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	// Print predeclared identifiers.
	if def, ok := p.types["bool"]; ok {
		if _, err := fmt.Fprintf(f, "%s;\n\n", def.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	// Print enums.
	for _, tag := range p.enumTags {
		t := p.enums[tag]
		if _, err := fmt.Fprintf(f, "%s;\n\n", t.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	// Print structs.
	for _, tag := range p.structTags {
		t := p.structs[tag]
		if _, err := fmt.Fprintf(f, "%s;\n\n", t.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	// Print unions.
	for _, tag := range p.unionTags {
		t := p.unions[tag]
		if _, err := fmt.Fprintf(f, "%s;\n\n", t.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	// Print typedefs.
	for _, def := range p.typedefs {
		if _, err := fmt.Fprintf(f, "%s;\n\n", def.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// --- [ Global declarations ] -------------------------------------------------

const (
	// Declarations header file name.
	declsName = "decls.h"
	// Overlay header file name format string.
	overlayNameFormat = "overlay_%x.h"
)

// dumpDecls outputs the declarations recorded by the parser to C headers stored
// in the output directory.
func dumpDecls(p *parser, outputDir string) error {
	// Create output file.
	declsPath := filepath.Join(outputDir, declsName)
	fmt.Println("creating:", declsPath)
	f, err := os.Create(declsPath)
	if err != nil {
		return errors.Wrapf(err, "unable to create declarations header %q", declsPath)
	}
	defer f.Close()
	// Store declarations of default binary.
	if err := dumpOverlay(f, p.Overlay); err != nil {
		return errors.WithStack(err)
	}
	// Store declarations of overlays.
	for _, overlay := range p.overlays {
		overlayName := fmt.Sprintf(overlayNameFormat, overlay.ID)
		overlayPath := filepath.Join(outputDir, overlayName)
		fmt.Println("creating:", overlayPath)
		f, err := os.Create(overlayPath)
		if err != nil {
			return errors.Wrapf(err, "unable to create overlay header %q", overlayPath)
		}
		defer f.Close()
		if err := dumpOverlay(f, overlay); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// dumpOverlay outputs the declarations of the overlay, writing to w.
func dumpOverlay(w io.Writer, overlay *Overlay) error {
	// Add types.h include directory.
	if _, err := fmt.Fprintf(w, "#include %q\n\n", typesName); err != nil {
		return errors.WithStack(err)
	}
	if overlay.Addr != 0 || overlay.ID != 0 || overlay.Length != 0 {
		if _, err := fmt.Fprintf(w, "// === [ Overlay ID %x ] ===\n\n", overlay.ID); err != nil {
			return errors.WithStack(err)
		}
	}
	// Print variable declarations.
	for _, v := range overlay.vars {
		if _, err := fmt.Fprintf(w, "%s;\n\n", v.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	// Print function declarations.
	for _, f := range overlay.funcs {
		if _, err := fmt.Fprintf(w, "%s\n\n", f.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// --- [ Source files ] --------------------------------------------------------

// A SourceFile is a source file.
type SourceFile struct {
	// Source file path.
	Path string
	// Variable declarations.
	vars []*c.VarDecl
	// Function declarations.
	funcs []*c.FuncDecl
}

// dumpSourceFiles outputs the source files recorded by the parser to the output
// directory.
func dumpSourceFiles(p *parser, outputDir string) error {
	srcs := getSourceFiles(p)
	for _, src := range srcs {
		// Create source file directory.
		path := strings.ToLower(src.Path)
		path = strings.Replace(path, `\`, "/", -1)
		if strings.HasPrefix(path[1:], ":/") {
			path = path[len("c:/"):]
		}
		path = filepath.Join(outputDir, path)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.WithStack(err)
		}
		fmt.Println("creating:", path)
		f, err := os.Create(path)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()
		if err := dumpSourceFile(f, src); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// dumpSourceFile outputs the declarations of the source file, writing to w.
func dumpSourceFile(w io.Writer, src *SourceFile) error {
	if _, err := fmt.Fprintf(w, "// %s\n\n", src.Path); err != nil {
		return errors.WithStack(err)
	}
	// Add types.h include directory.
	if _, err := fmt.Fprintf(w, "#include %q\n\n", typesName); err != nil {
		return errors.WithStack(err)
	}
	// Handle duplicate identifiers.
	names := make(map[string]bool)
	for _, v := range src.vars {
		if names[v.Name] {
			v.Name = uniqueName(v.Name, v.Addr)
		}
		names[v.Name] = true
	}
	for _, f := range src.funcs {
		if names[f.Name] {
			f.Name = uniqueName(f.Name, f.Addr)
		}
		names[f.Name] = true
	}
	// Print variable declarations.
	for _, v := range src.vars {
		if _, err := fmt.Fprintf(w, "%s;\n\n", v.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	// Print function declarations.
	for _, f := range src.funcs {
		if _, err := fmt.Fprintf(w, "%s\n\n", f.Def()); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// --- [ IDA scripts ] ---------------------------------------------------------

// dumpIDAScripts outputs the declarations recorded by the parser to IDA scripts
// stored in the output directory.
func dumpIDAScripts(p *parser, outputDir string) error {
	// Create scripts for declarations of default binary.
	if err := dumpIDAOverlay(p.Overlay, outputDir); err != nil {
		return errors.WithStack(err)
	}
	// Create scripts for declarations of overlays.
	for _, overlay := range p.overlays {
		if err := dumpIDAOverlay(overlay, outputDir); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// IDA script names.
const (
	// Scripts mapping addresses to identifiers.
	idaIdentsName = "make_psx.py"
	// Scripts adding function signatures to identifiers.
	idaFuncsName = "set_funcs.py"
	// Scripts adding global variable types to identifiers.
	idaVarsName = "set_vars.py"
)

// dumpIDAOverlay outputs the declarations of the overlay to IDA scripts.
func dumpIDAOverlay(overlay *Overlay, outputDir string) error {
	// Create scripts for mapping addresses to identifiers.
	dir := outputDir
	if overlay.ID != 0 {
		overlayDir := fmt.Sprintf("overlay_%x", overlay.ID)
		dir = filepath.Join(outputDir, overlayDir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.WithStack(err)
		}
	}
	identsPath := filepath.Join(dir, idaIdentsName)
	fmt.Println("creating:", identsPath)
	w, err := os.Create(identsPath)
	if err != nil {
		return errors.Wrapf(err, "unable to create declarations IDA script %q", identsPath)
	}
	defer w.Close()
	for _, f := range overlay.funcs {
		if _, err := fmt.Fprintf(w, "MakeNameEx(0x%08X, %q, SN_NOWARN)\n", f.Addr, f.Name); err != nil {
			return errors.WithStack(err)
		}
	}
	for _, v := range overlay.vars {
		if _, err := fmt.Fprintf(w, "MakeNameEx(0x%08X, %q, SN_NOWARN)\n", v.Addr, v.Name); err != nil {
			return errors.WithStack(err)
		}
	}
	// Create scripts for adding function signatures to identifiers.
	funcsPath := filepath.Join(dir, idaFuncsName)
	fmt.Println("creating:", funcsPath)
	w, err = os.Create(funcsPath)
	if err != nil {
		return errors.Wrapf(err, "unable to create function signatures IDA script %q", funcsPath)
	}
	defer w.Close()
	for _, f := range overlay.funcs {
		if _, err := fmt.Fprintf(w, "del_items(0x%08X)\n", f.Addr); err != nil {
			return errors.WithStack(err)
		}
		if _, err := fmt.Fprintf(w, "SetType(0x%08X, %q)\n", f.Addr, f.Var); err != nil {
			return errors.WithStack(err)
		}
	}
	// Create scripts adding global variable types to identifiers.
	varsPath := filepath.Join(dir, idaVarsName)
	fmt.Println("creating:", varsPath)
	w, err = os.Create(varsPath)
	if err != nil {
		return errors.Wrapf(err, "unable to create global variables IDA script %q", varsPath)
	}
	defer w.Close()
	for _, v := range overlay.vars {
		if _, err := fmt.Fprintf(w, "del_items(0x%08X)\n", v.Addr); err != nil {
			return errors.WithStack(err)
		}
		if _, err := fmt.Fprintf(w, "SetType(0x%08X, %q)\n", v.Addr, v.Var); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// ### [ Helper functions ] ####################################################

// getSourceFiles returns the source files recorded by the parser.
func getSourceFiles(p *parser) []*SourceFile {
	// Record source file information from overlays.
	overlays := append(p.overlays, p.Overlay)
	// sources maps from source path to source file.
	sources := make(map[string]*SourceFile)
	for _, overlay := range overlays {
		srcPathFromAddr := make(map[uint32]string)
		for _, line := range overlay.lines {
			srcPathFromAddr[line.Addr] = line.Path
		}
		for _, v := range overlay.vars {
			srcPath := fmt.Sprintf("global_%x.cpp", overlay.ID)
			src, ok := sources[srcPath]
			if !ok {
				src = &SourceFile{
					Path: srcPath,
				}
				sources[srcPath] = src
			}
			src.vars = append(src.vars, v)
		}
		for _, f := range overlay.funcs {
			srcPath, ok := srcPathFromAddr[f.Addr]
			if !ok {
				panic(fmt.Errorf("unable to locate source file of function %q at address 0x%08X", f.Name, f.Addr))
			}
			src, ok := sources[srcPath]
			if !ok {
				src = &SourceFile{
					Path: srcPath,
				}
				sources[srcPath] = src
			}
			src.funcs = append(src.funcs, f)
		}
	}
	var srcs []*SourceFile
	for _, src := range sources {
		srcs = append(srcs, src)
	}
	less := func(i, j int) bool {
		return srcs[i].Path < srcs[j].Path
	}
	sort.Slice(srcs, less)
	return srcs
}
