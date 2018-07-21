package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Type definitions header file name.
const typesName = "types.h"

// dumpTypes outputs the type information recorded by the parser to a C header
// stored in the output directory.
func dumpTypes(p *parser, outputDir string) error {
	// Create output file.
	typesPath := filepath.Join(outputDir, typesName)
	f, err := os.Create(typesPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	// Print predeclared identifiers.
	def := p.types["bool"]
	if _, err := fmt.Fprintf(f, "%s;\n\n", def.Def()); err != nil {
		return errors.WithStack(err)
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

const (
	// Declarations header file name.
	declsName = "decls.h"
	// Overlay header file name format string.
	overlayNameFormat = "overlay_%x.h"
)

// dumpDecls outputs the declarations recorded by the parser to a C header
// stored in the output directory.
func dumpDecls(p *parser, outputDir string) error {
	// Create output file.
	declsPath := filepath.Join(outputDir, declsName)
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
