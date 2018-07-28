// Package csym translates Playstation 1 symbol information to C declarations.
package csym

import (
	"github.com/sanctuary/sym/csym/c"
)

// Parser tracks type information used for parsing.
type Parser struct {
	// Type information.

	// structs maps from struct tag to struct type.
	Structs map[string]*c.StructType
	// unions maps from union tag to union type.
	Unions map[string]*c.UnionType
	// enums maps from enum tag to enum type.
	Enums map[string]*c.EnumType
	// types maps from type name to underlying type definition.
	Types map[string]c.Type
	// Struct tags in order of occurrence in SYM file.
	StructTags []string
	// Union tags in order of occurrence in SYM file.
	UnionTags []string
	// Enum tags in order of occurrence in SYM file.
	EnumTags []string
	// Type definitions in order of occurrence in SYM file.
	Typedefs []c.Type
	// Tracks unique enum member names.
	enumMembers map[string]bool

	// Declarations.
	*Overlay // default binary

	// Overlays.
	Overlays []*Overlay
	// overlayIDs maps from overlay ID to overlay.
	overlayIDs map[uint32]*Overlay

	// Current overlay.
	curOverlay *Overlay
}

// NewParser returns a new parser.
func NewParser() *Parser {
	overlay := &Overlay{
		varNames:  make(map[string]*c.VarDecl),
		funcNames: make(map[string]*c.FuncDecl),
	}
	return &Parser{
		Structs:     make(map[string]*c.StructType),
		Unions:      make(map[string]*c.UnionType),
		Enums:       make(map[string]*c.EnumType),
		Types:       make(map[string]c.Type),
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
	Vars []*c.VarDecl
	// Function delcarations.
	Funcs []*c.FuncDecl
	// varNames maps from variable name to variable declaration.
	varNames map[string]*c.VarDecl
	// funcNames maps from function name to function declaration.
	funcNames map[string]*c.FuncDecl

	// Symbols.
	symbols []*Symbol
	// Source file line numbers.
	Lines []*Line
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
