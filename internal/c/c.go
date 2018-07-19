// Package c provides an AST for a subset of C.
package c

import (
	"fmt"
	"strings"
)

// A VarDecl is a variable declaration.
type VarDecl struct {
	// Address (optional).
	Address uint32
	// Size (optional).
	Size uint32
	// Underlying variable.
	Var
}

// String returns the string representation of the variable declaration.
func (v *VarDecl) String() string {
	buf := &strings.Builder{}
	if v.Address > 0 {
		fmt.Fprintf(buf, "// address: 0x%08X\n", v.Address)
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
	Address uint32
	// Size (optional).
	Size uint32
	// Underlying function variable.
	Var
	// Local variables.
	Locals []Var
}

// String returns the string representation of the function declaration.
func (f *FuncDecl) String() string {
	buf := &strings.Builder{}
	if f.Address > 0 {
		fmt.Fprintf(buf, "// address: 0x%08X\n", f.Address)
	}
	if f.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", f.Size)
	}
	buf.WriteString(f.Var.String())
	// TODO: Output local variables.
	return buf.String()
}
