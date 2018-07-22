package main

import (
	"github.com/sanctuary/sym/internal/c"
)

// parser tracks type information used for parsing.
type parser struct {
	// Type information.

	// structs maps from struct tag to struct type.
	structs map[string]*c.StructType
	// unions maps from union tag to union type.
	unions map[string]*c.UnionType
	// enums maps from enum tag to enum type.
	enums map[string]*c.EnumType
	// types maps from type name to underlying type definition.
	types map[string]c.Type
	// Struct tags in order of occurrence in SYM file.
	structTags []string
	// Union tags in order of occurrence in SYM file.
	unionTags []string
	// Enum tags in order of occurrence in SYM file.
	enumTags []string
	// Type definitions in order of occurrence in SYM file.
	typedefs []c.Type
	// Tracks unique enum member names.
	enumMembers map[string]bool

	// Declarations.
	*Overlay // default binary

	// Overlays.
	overlays []*Overlay
	// overlayIDs maps from overlay ID to overlay.
	overlayIDs map[uint32]*Overlay

	// Current overlay.
	curOverlay *Overlay
}

// newParser returns a new parser.
func newParser() *parser {
	overlay := &Overlay{
		varNames:  make(map[string]*c.VarDecl),
		funcNames: make(map[string]*c.FuncDecl),
	}
	return &parser{
		structs:     make(map[string]*c.StructType),
		unions:      make(map[string]*c.UnionType),
		enums:       make(map[string]*c.EnumType),
		types:       make(map[string]c.Type),
		enumMembers: make(map[string]bool),
		Overlay:     overlay,
		overlayIDs:  make(map[uint32]*Overlay),
		curOverlay:  overlay,
	}
}

// An Overlay is an overlay appended to the end of the executable.
type Overlay struct {
	// Base address at which the overlay is loaded.
	Addr uint32
	// Overlay ID.
	ID uint32
	// Overlay length in bytes.
	Length uint32

	// Variable delcarations.
	vars []*c.VarDecl
	// Function delcarations.
	funcs []*c.FuncDecl
	// varNames maps from variable name to variable declaration.
	varNames map[string]*c.VarDecl
	// funcNames maps from function name to function declaration.
	funcNames map[string]*c.FuncDecl

	// Symbols.
	symbols []*Symbol
	// Source file line numbers.
	lines []*Line
}

// A Symbol associates a symbol name with an address.
type Symbol struct {
	// Symbol address.
	Addr uint32
	// Symbol name.
	Name string
}

// A Line associates a line number in a source file with an address.
type Line struct {
	// Address.
	Addr uint32
	// Source file name.
	Path string
	// Line number.
	Line uint32
}
