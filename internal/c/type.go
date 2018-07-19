package c

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Type is a C type.
type Type interface {
	fmt.Stringer
	// Def returns the C syntax representation of the definition of the type.
	Def() string
}

// --- [ Type definition ] -----------------------------------------------------

// Typedef is a type definition.
type Typedef struct {
	// Underlying type.
	Type Type
	// Type name.
	Name string
}

// String returns the string representation of the type definition.
func (t *Typedef) String() string {
	return t.Name
}

// Def returns the C syntax representation of the definition of the type.
func (t *Typedef) Def() string {
	// HACK, but works. The syntax of the C type system is pre-historic.
	v := Var{
		Type: t.Type,
		Name: t.Name,
	}
	return fmt.Sprintf("typedef %s", v)
}

// --- [ Base type ] -----------------------------------------------------------

//go:generate stringer -linecomment -type BaseType

// BaseType is a base type.
type BaseType uint8

// Base types.
const (
	Void   BaseType = iota + 1 // void
	Char                       // char
	Short                      // short
	Int                        // int
	Long                       // long
	UChar                      // unsigned char
	UShort                     // unsigned short
	UInt                       // unsigned int
	ULong                      // unsigned long
)

// Def returns the C syntax representation of the definition of the type.
func (t BaseType) Def() string {
	return t.String()
}

// --- [ Struct type ] ---------------------------------------------------------

// StructType is a structure type.
type StructType struct {
	// Size in bytes (optional).
	Size uint32
	// Structure tag.
	Tag string
	// Structure fields.
	Fields []Field
}

// String returns the string representation of the structure type.
func (t *StructType) String() string {
	return fmt.Sprintf("struct %s", t.Tag)
}

// Def returns the C syntax representation of the definition of the type.
func (t *StructType) Def() string {
	buf := &strings.Builder{}
	if t.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", t.Size)
	}
	if len(t.Tag) > 0 {
		fmt.Fprintf(buf, "struct %s {\n", t.Tag)
	} else {
		buf.WriteString("struct {\n")
	}
	for _, field := range t.Fields {
		if field.Size > 0 {
			fmt.Fprintf(buf, "\t// offset: %04X (%d bytes)\n", field.Offset, field.Size)
		} else if len(t.Fields) > 1 && t.Fields[1].Offset > 0 {
			fmt.Fprintf(buf, "\t// offset: %04X\n", field.Offset)
		}
		fmt.Fprintf(buf, "\t%s;\n", field)
	}
	buf.WriteString("}")
	return buf.String()
}

// --- [ Union type ] ---------------------------------------------------------

// UnionType is a union type.
type UnionType struct {
	// Size in bytes (optional).
	Size uint32
	// Union tag.
	Tag string
	// Union fields.
	Fields []Field
}

// String returns the string representation of the union type.
func (t *UnionType) String() string {
	return fmt.Sprintf("union %s", t.Tag)
}

// Def returns the C syntax representation of the definition of the type.
func (t *UnionType) Def() string {
	buf := &strings.Builder{}
	if t.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", t.Size)
	}
	if len(t.Tag) > 0 {
		fmt.Fprintf(buf, "union %s {\n", t.Tag)
	} else {
		buf.WriteString("union {\n")
	}
	for _, field := range t.Fields {
		if field.Size > 0 {
			fmt.Fprintf(buf, "\t// offset: %04X (%d bytes)\n", field.Offset, field.Size)
		} else if len(t.Fields) > 1 && t.Fields[1].Offset > 0 {
			fmt.Fprintf(buf, "\t// offset: %04X\n", field.Offset)
		}
		fmt.Fprintf(buf, "\t%s;\n", field)
	}
	buf.WriteString("}")
	return buf.String()
}

// --- [ Enum type ] -----------------------------------------------------------

// EnumType is a enum type.
type EnumType struct {
	// Enum tag.
	Tag string
	// Enum members.
	Members []*EnumMember
}

// String returns the string representation of the enum type.
func (t *EnumType) String() string {
	return fmt.Sprintf("enum %s", t.Tag)
}

// Def returns the C syntax representation of the definition of the type.
func (t *EnumType) Def() string {
	buf := &strings.Builder{}
	if len(t.Tag) > 0 {
		fmt.Fprintf(buf, "enum %s {\n", t.Tag)
	} else {
		buf.WriteString("enum {\n")
	}
	less := func(i, j int) bool {
		if t.Members[i].Value == t.Members[j].Value {
			return t.Members[i].Name < t.Members[j].Name
		}
		return t.Members[i].Value < t.Members[j].Value
	}
	sort.Slice(t.Members, less)
	w := tabwriter.NewWriter(buf, 1, 3, 1, ' ', tabwriter.TabIndent)
	for _, member := range t.Members {
		fmt.Fprintf(w, "\t%s\t= %d,\n", member.Name, member.Value)
	}
	if err := w.Flush(); err != nil {
		panic(fmt.Errorf("unable to flush tabwriter; %v", err))
	}
	buf.WriteString("}")
	return buf.String()
}

// ~~~ [ Enum member ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// EnumMember is an enum member.
type EnumMember struct {
	// Enum value.
	Value uint32
	// Enum name.
	Name string
}

// --- [ Pointer type ] --------------------------------------------------------

// PointerType is a pointer type.
type PointerType struct {
	// Element type.
	Elem Type
}

// String returns the string representation of the pointer type.
func (t *PointerType) String() string {
	return fmt.Sprintf("%s*", t.Elem)
}

// Def returns the C syntax representation of the definition of the type.
func (t *PointerType) Def() string {
	return t.String()
}

// --- [ Array type ] ----------------------------------------------------------

// ArrayType is an array type.
type ArrayType struct {
	// Element type.
	Elem Type
	// Array length.
	Len int
}

// String returns the string representation of the array type.
func (t *ArrayType) String() string {
	if t.Len > 0 {
		return fmt.Sprintf("%s[%d]", t.Elem, t.Len)
	}
	return fmt.Sprintf("%s[]", t.Elem)
}

// Def returns the C syntax representation of the definition of the type.
func (t *ArrayType) Def() string {
	return t.String()
}

// --- [ Function type ] -------------------------------------------------------

// FuncType is a function type.
type FuncType struct {
	// Return type.
	RetType Type
	// Function parameters.
	Params []Var
	// Variadic function.
	Variadic bool
}

// String returns the string representation of the function type.
func (t *FuncType) String() string {
	// HACK, but works. The syntax of the C type system is pre-historic.
	v := Var{Type: t}
	return v.String()
}

// Def returns the C syntax representation of the definition of the type.
func (t *FuncType) Def() string {
	return t.String()
}

// ### [ Helper types ] ########################################################

// A Field represents a field in a structure type or union type.
type Field struct {
	// Offset (optional).
	Offset uint32
	// Size in bytes (optional).
	Size uint32
	// Underlying variable.
	Var
}

// A Var represents a variable declaration or function parameter.
type Var struct {
	// Variable type.
	Type Type
	// Variable name.
	Name string
}

// String returns the string representation of the variable.
func (v Var) String() string {
	switch t := v.Type.(type) {
	case *PointerType:
		// HACK, but works. The syntax of the C type system is pre-historic.
		switch t.Elem.(type) {
		case *FuncType, *ArrayType:
			// Add grouping parenthesis.
			v.Name = fmt.Sprintf("(*%s)", v.Name)
		default:
			v.Name = fmt.Sprintf("*%s", v.Name)
		}
		v.Type = t.Elem
		return v.String()
	case *ArrayType:
		// HACK, but works. The syntax of the C type system is pre-historic.
		if t.Len > 0 {
			v.Name = fmt.Sprintf("%s[%d]", v.Name, t.Len)
		} else {
			v.Name = fmt.Sprintf("%s[]", v.Name)
		}
		v.Type = t.Elem
		return v.String()
	case *FuncType:
		// HACK, but works. The syntax of the C type system is pre-historic.
		buf := &strings.Builder{}
		fmt.Fprintf(buf, "%s(", v.Name)
		for i, param := range t.Params {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(param.String())
		}
		if t.Variadic {
			if len(t.Params) > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("...")
		}
		buf.WriteString(")")
		v.Name = buf.String()
		v.Type = t.RetType
		return v.String()
	case *UnionType:
		if isFakeTag(t.Tag) {
			return fmt.Sprintf("%s %s", fakeUnionString(t), v.Name)
		}
		return fmt.Sprintf("%s %s", t, v.Name)
	default:
		return fmt.Sprintf("%s %s", t, v.Name)
	}
}

// fakeUnionString returns the string representation of the given union with a
// fake name.
func fakeUnionString(t *UnionType) string {
	buf := &strings.Builder{}
	if t.Size > 0 {
		fmt.Fprintf(buf, "// size: 0x%X\n", t.Size)
	}
	buf.WriteString("\tunion {\n")
	for _, field := range t.Fields {
		if field.Size > 0 {
			fmt.Fprintf(buf, "\t\t// offset: %04X (%d bytes)\n", field.Offset, field.Size)
		} else if len(t.Fields) > 1 && t.Fields[1].Offset > 0 {
			fmt.Fprintf(buf, "\t\t// offset: %04X\n", field.Offset)
		}
		fmt.Fprintf(buf, "\t\t%s;\n", field)
	}
	buf.WriteString("\t}")
	return buf.String()
}

// isFakeTag reports whether the tag name is fake (generated by the compiler for
// symbols lacking a tag name).
func isFakeTag(tag string) bool {
	if strings.HasPrefix(tag, "_") && strings.HasSuffix(tag, "fake") {
		s := tag[len("_") : len(tag)-len("fake")]
		_, err := strconv.Atoi(s)
		return err == nil
	}
	return false
}
