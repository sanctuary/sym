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
	// Return type
	RetType Type
	// Function name.
	Name string
	// Function parameters.
	Params []Var
	// Variadic.
	Variadic bool
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
	// int (*)(int a, int b)
	fmt.Fprintf(buf, "%s %s(", f.RetType, f.Name)
	for i, param := range f.Params {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(param.String())
	}
	if f.Variadic {
		if len(f.Params) > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("...")
	}
	buf.WriteString(")")
	return buf.String()
}
