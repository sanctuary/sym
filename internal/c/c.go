// Package c provides an AST for a subset of C.
package c

import (
	"fmt"
	"strings"
)

// A VarDecl is a variable declaration.
type VarDecl struct {
	// Address (optional).
	Addr uint32
	// Size (optional).
	Size uint32
	// Underlying variable.
	Var
}

// String returns the string representation of the variable declaration.
func (v *VarDecl) String() string {
	buf := &strings.Builder{}
	if v.Addr > 0 {
		fmt.Fprintf(buf, "// address: 0x%08X\n", v.Addr)
	}
	if v.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", v.Size)
	}
	fmt.Fprintf(buf, "extern %s", v.Var)
	return buf.String()
}

// A FuncDecl is a function declaration.
type FuncDecl struct {
	// Address (optional).
	Addr uint32
	// Size (optional).
	Size uint32
	// Start line number.
	LineStart uint32
	// End line number.
	LineEnd uint32
	// Underlying function variable.
	Var
	// Scope blocks.
	Blocks []*Block
}

// String returns the string representation of the function declaration.
func (f *FuncDecl) String() string {
	buf := &strings.Builder{}
	if f.Addr > 0 {
		fmt.Fprintf(buf, "// address: 0x%08X\n", f.Addr)
	}
	if f.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", f.Size)
	}
	buf.WriteString(f.Var.String())
	// TODO: Output local variables.
	return buf.String()
}

// A Block encapsulates a block scope.
type Block struct {
	// Start line number (relative to the function).
	LineStart uint32
	// End line number (relative to the function).
	LineEnd uint32
	// Local variables.
	Locals []Var
}
