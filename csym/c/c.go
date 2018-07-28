// Package c provides an AST for a subset of C.
package c

import (
	"fmt"
	"strings"
)

// A VarDecl is a variable declaration.
type VarDecl struct {
	// Address, frame pointer delta, or register depending on storage class
	// (optional).
	Addr uint32
	// Size (optional).
	Size uint32
	// Storage class.
	Class StorageClass
	// Underlying variable.
	Var
}

// String returns the string representation of the variable declaration.
func (v *VarDecl) String() string {
	return v.Name
}

// Def returns the C syntax representation of the definition of the variable
// declaration.
func (v *VarDecl) Def() string {
	buf := &strings.Builder{}
	switch v.Class {
	case Register:
		fmt.Fprintf(buf, "// register: %d\n", v.Addr)
	default:
		if v.Addr > 0 {
			fmt.Fprintf(buf, "// address: 0x%08X\n", v.Addr)
		}
	}
	if v.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", v.Size)
	}
	if v.Class == 0 {
		fmt.Fprintf(buf, "%s", v.Var)
	} else {
		fmt.Fprintf(buf, "%s %s", v.Class, v.Var)
	}
	return buf.String()
}

//go:generate stringer -linecomment -type StorageClass

// A StorageClass is a storage class.
type StorageClass uint8

// Base types.
const (
	Auto     StorageClass = iota + 1 // auto
	Extern                           // extern
	Static                           // static
	Register                         // register
	Typedef                          // typedef
)

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
	return f.Name
}

// Def returns the C syntax representation of the definition of the function
// declaration.
func (f *FuncDecl) Def() string {
	// TODO: Print storage class.
	buf := &strings.Builder{}
	if f.Addr > 0 {
		fmt.Fprintf(buf, "// address: 0x%08X\n", f.Addr)
	}
	if f.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", f.Size)
	}
	if len(f.Blocks) == 0 {
		fmt.Fprintf(buf, "%s;", f.Var)
		return buf.String()
	}
	fmt.Fprintf(buf, "%s ", f.Var)
	for i, block := range f.Blocks {
		indent := strings.Repeat("\t", i)
		fmt.Fprintf(buf, "%s{\n", indent)
		for _, local := range block.Locals {
			indent := strings.Repeat("\t", i+1)
			l := strings.Replace(local.Def(), "\n", "\n"+indent, -1)
			fmt.Fprintf(buf, "%s%s;\n", indent, l)
		}
	}
	for i := len(f.Blocks) - 1; i >= 0; i-- {
		indent := strings.Repeat("\t", i)
		fmt.Fprintf(buf, "%s}\n", indent)
	}
	return buf.String()
}

// A Block encapsulates a block scope.
type Block struct {
	// Start line number (relative to the function).
	LineStart uint32
	// End line number (relative to the function).
	LineEnd uint32
	// Local variables.
	Locals []*VarDecl
}
