package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mewkiz/pkg/term"
	"github.com/sanctuary/sym"
	"github.com/sanctuary/sym/internal/c"
)

// Prefix added to duplicate symbols.
const duplicatePrefix = "_duplicate_"

var (
	// TODO: remove debug output once C output is mature.
	dbg = log.New(os.Stderr, term.Cyan("dbg:")+" ", log.Lshortfile)
)

// dumpC outputs the SYM file as C headers.
func dumpC(f *sym.File) error {
	// Add scaffolding types for structs, unions and enums, so they may be
	// referrenced before defined.
	p := newParser()
	// TODO: parse all symbols.
	var (
		uniqueStruct = make(map[string]bool)
		uniqueUnion  = make(map[string]bool)
		uniqueEnum   = make(map[string]bool)
	)
	for _, s := range f.Syms[:7117] {
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassSTRTAG:
				tag := validName(body.Name)
				if uniqueStruct[tag] {
					tag = duplicatePrefix + tag
				}
				uniqueStruct[tag] = true
				t := &c.StructType{
					Size: body.Size,
					Tag:  tag,
				}
				p.structs[tag] = t
				p.structTags = append(p.structTags, tag)
			case sym.ClassUNTAG:
				tag := validName(body.Name)
				if uniqueUnion[tag] {
					tag = duplicatePrefix + tag
				}
				uniqueUnion[tag] = true
				t := &c.UnionType{
					Size: body.Size,
					Tag:  tag,
				}
				p.unions[tag] = t
				p.unionTags = append(p.unionTags, tag)
			case sym.ClassENTAG:
				tag := validName(body.Name)
				if uniqueEnum[tag] {
					tag = duplicatePrefix + tag
				}
				uniqueEnum[tag] = true
				t := &c.EnumType{
					Tag: tag,
				}
				p.enums[tag] = t
				p.enumTags = append(p.enumTags, tag)
			}
		}
	}
	p.structs["__vtbl_ptr_type"] = &c.StructType{
		Tag: "__vtbl_ptr_type",
	}
	// Bool used for NULL type.
	def := &c.Typedef{
		Type: c.Int,
		Name: "bool",
	}
	p.typedefs = append(p.typedefs, def)
	p.types["bool"] = def
	// Parse symbols.
	// TODO: parse all symbols.
	for i := 0; i < len(f.Syms[:7117]); i++ {
		s := f.Syms[i]
		// TODO: remove debug output once C output is mature.
		fmt.Fprintln(os.Stderr, "sym:", s)
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassSTRTAG:
				n := p.parseClassSTRTAG(body, f.Syms[i+1:])
				i += n
			case sym.ClassUNTAG:
				n := p.parseClassUNTAG(body, f.Syms[i+1:])
				i += n
			case sym.ClassTPDEF:
				p.parseClassTPDEF(body.Type, body.Name, nil, "")
			case sym.ClassENTAG:
				n := p.parseClassENTAG(body, f.Syms[i+1:])
				i += n
			default:
				dbg.Printf("support for class %q not yet implemented", body.Class)
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassTPDEF:
				p.parseClassTPDEF(body.Type, body.Name, body.Dims, body.Tag)
			default:
				dbg.Printf("support for class %q not yet implemented", body.Class)
			}
		case *sym.Overlay:
		// nothing to do.
		default:
			dbg.Printf("support for symbol body %T not yet implemented", body)
		}
	}
	p.dump()
	return nil
}

// parser tracks type information used for parsing.
type parser struct {
	// structs maps from struct tag to struct type.
	structs map[string]*c.StructType
	// unions maps from union tag to union type.
	unions map[string]*c.UnionType
	// enums maps from enum tag to enum type.
	enums map[string]*c.EnumType
	// types maps from type name to underlying type definition.
	types map[string]*c.Typedef
	// Struct tags in order of occurrence in SYM file.
	structTags []string
	// Union tags in order of occurrence in SYM file.
	unionTags []string
	// Enum tags in order of occurrence in SYM file.
	enumTags []string
	// Type definitions in order of occurrence in SYM file.
	typedefs []*c.Typedef
	// Tracks unique enum member names.
	uniqueEnumMember map[string]bool
}

// newParser returns a new parser.
func newParser() *parser {
	return &parser{
		structs:          make(map[string]*c.StructType),
		unions:           make(map[string]*c.UnionType),
		enums:            make(map[string]*c.EnumType),
		types:            make(map[string]*c.Typedef),
		uniqueEnumMember: make(map[string]bool),
	}
}

// dump outputs the type information recorded by the parser as a C header.
func (p *parser) dump() {
	// Print predeclared identifiers.
	def := p.types["bool"]
	fmt.Println(def.Def())
	// Print enums.
	for _, tag := range p.enumTags {
		t := p.enums[tag]
		fmt.Printf("%s;\n", t.Def())
	}
	// Print structs.
	for _, tag := range p.structTags {
		t := p.structs[tag]
		fmt.Printf("%s;\n", t.Def())
	}
	// Print unions.
	for _, tag := range p.unionTags {
		t := p.unions[tag]
		fmt.Printf("%s;\n", t.Def())
	}
	// Print typedefs.
	for _, def := range p.typedefs {
		fmt.Println(def.Def())
	}
}

// parseClassSTRTAG parses a struct tag sequence of symbols.
func (p *parser) parseClassSTRTAG(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseStruct {
		panic(fmt.Sprintf("support for base type %q not yet implemented", base))
	}
	name := validName(body.Name)
	t, ok := p.structs[name]
	if !ok {
		panic(fmt.Sprintf("unable to locate struct %q", name))
	}
	if len(t.Fields) > 0 {
		log.Printf("duplicate struct tag %q symbol", name)
		dupTag := duplicatePrefix + name
		t = &c.StructType{
			Size: body.Size,
			Tag:  dupTag,
		}
		p.structs[dupTag] = t
	}
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOS:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Type:   p.parseType(body.Type, nil, ""),
					Name:   validName(body.Name),
				}
				t.Fields = append(t.Fields, field)
			case sym.ClassFIELD:
				// TODO: Figure out how to handle FIELD. For now, parse as MOS.
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Type:   p.parseType(body.Type, nil, ""),
					Name:   validName(body.Name),
				}
				t.Fields = append(t.Fields, field)
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassMOS:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Type:   p.parseType(body.Type, body.Dims, body.Tag),
					Name:   validName(body.Name),
				}
				t.Fields = append(t.Fields, field)
			case sym.ClassEOS:
				return n + 1
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		}
	}
	panic("unreachable")
}

// parseClassUNTAG parses a union tag sequence of symbols.
func (p *parser) parseClassUNTAG(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseUnion {
		panic(fmt.Sprintf("support for base type %q not yet implemented", base))
	}
	name := validName(body.Name)
	t, ok := p.unions[name]
	if !ok {
		panic(fmt.Sprintf("unable to locate union %q", name))
	}
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOU:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Type:   p.parseType(body.Type, nil, ""),
					Name:   validName(body.Name),
				}
				t.Fields = append(t.Fields, field)
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassMOU:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Type:   p.parseType(body.Type, body.Dims, body.Tag),
					Name:   validName(body.Name),
				}
				t.Fields = append(t.Fields, field)
			case sym.ClassEOS:
				return n + 1
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		}
	}
	panic("unreachable")
}

// parseClassTPDEF parses a typedef symbol.
func (p *parser) parseClassTPDEF(t sym.Type, name string, dims []uint32, tag string) {
	name = validName(name)
	def := &c.Typedef{
		Type: p.parseType(t, dims, tag),
		Name: name,
	}
	p.typedefs = append(p.typedefs, def)
	p.types[name] = def
}

// parseClassENTAG parses an enum tag sequence of symbols.
func (p *parser) parseClassENTAG(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseEnum {
		panic(fmt.Sprintf("support for base type %q not yet implemented", base))
	}
	name := validName(body.Name)
	t, ok := p.enums[name]
	if !ok {
		panic(fmt.Sprintf("unable to locate enum %q", name))
	}
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOE:
				name := validName(body.Name)
				if p.uniqueEnumMember[name] {
					name = strings.ToUpper(duplicatePrefix) + name
				}
				p.uniqueEnumMember[name] = true
				member := &c.EnumMember{
					Value: s.Hdr.Value,
					Name:  name,
				}
				t.Members = append(t.Members, member)
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassEOS:
				return n + 1
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		}
	}
	panic("unreachable")
}

// ### [ Helper functions ] ####################################################

// parseBase parses the SYM type into the equivalent C type.
func (p *parser) parseType(t sym.Type, dims []uint32, tag string) c.Type {
	u := p.parseBase(t.Base(), tag)
	return parseMods(u, t.Mods(), dims)
}

// parseBase parses the SYM base type into the equivalent C type.
func (p *parser) parseBase(base sym.Base, tag string) c.Type {
	tag = validName(tag)
	switch base {
	case sym.BaseNull:
		return p.types["bool"]
	case sym.BaseVoid:
		return c.Void
	case sym.BaseChar:
		return c.Char
	case sym.BaseShort:
		return c.Short
	case sym.BaseInt:
		return c.Int
	case sym.BaseLong:
		return c.Long
	case sym.BaseStruct:
		t, ok := p.structs[tag]
		if !ok {
			panic(fmt.Sprintf("unable to locate struct %q", tag))
		}
		return t
	case sym.BaseUnion:
		t, ok := p.unions[tag]
		if !ok {
			panic(fmt.Sprintf("unable to locate union %q", tag))
		}
		return t
	case sym.BaseEnum:
		t, ok := p.enums[tag]
		if !ok {
			panic(fmt.Sprintf("unable to locate enum %q", tag))
		}
		return t
	//case sym.BaseMOE:
	case sym.BaseUChar:
		return c.UChar
	case sym.BaseUShort:
		return c.UShort
	case sym.BaseUInt:
		return c.UInt
	case sym.BaseULong:
		return c.ULong
	default:
		panic(fmt.Sprintf("base type %q not yet supported", base))
	}
}

// parseMods parses the SYM type modifiers into the equivalent C type modifiers.
func parseMods(t c.Type, mods []sym.Mod, dims []uint32) c.Type {
	j := 0
	// TODO: consider rewriting c.Type.Mods to calculate mask from right to left
	// instead of left to right.
	for i := len(mods) - 1; i >= 0; i-- {
		mod := mods[i]
		switch mod {
		case sym.ModPointer:
			t = &c.PointerType{Elem: t}
		case sym.ModFunction:
			// TODO: Figure out how to store params and variadic.
			t = &c.FuncType{
				RetType: t,
			}
		case sym.ModArray:
			t = &c.ArrayType{
				Elem: t,
				Len:  int(dims[j]),
			}
			j++
		}
	}
	return t
}

// validName returns a valid C identifier based on the given name.
func validName(name string) string {
	f := func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9':
			return r
		default:
			return '_'
		}
	}
	return strings.Map(f, name)
}
