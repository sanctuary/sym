package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mewkiz/pkg/term"
	"github.com/sanctuary/sym"
	"github.com/sanctuary/sym/internal/c"
)

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
	for _, s := range f.Syms[:7117] {
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassSTRTAG:
				t := &c.StructType{
					Size: body.Size,
					Tag:  body.Name,
				}
				p.structs[t.Tag] = t
			case sym.ClassUNTAG:
				t := &c.UnionType{
					Size: body.Size,
					Tag:  body.Name,
				}
				p.unions[t.Tag] = t
			case sym.ClassENTAG:
				t := &c.EnumType{
					Tag: body.Name,
				}
				p.enums[t.Tag] = t
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
	// Type definitions in order of occurrence in SYM file.
	typedefs []*c.Typedef
}

// newParser returns a new parser.
func newParser() *parser {
	return &parser{
		structs: make(map[string]*c.StructType),
		unions:  make(map[string]*c.UnionType),
		enums:   make(map[string]*c.EnumType),
		types:   make(map[string]*c.Typedef),
	}
}

// dump outputs the type information recorded by the parser as a C header.
func (p *parser) dump() {
	for _, def := range p.typedefs {
		fmt.Println(def.Def())
	}
}

// parseClassSTRTAG parses a struct tag sequence of symbols.
func (p *parser) parseClassSTRTAG(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseStruct {
		panic(fmt.Sprintf("support for base type %q not yet implemented", base))
	}
	t, ok := p.structs[body.Name]
	if !ok {
		panic(fmt.Sprintf("unable to locate struct %q", body.Name))
	}
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOS:
				field := &c.Field{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				t.Fields = append(t.Fields, field)
			case sym.ClassFIELD:
				// TODO: Figure out how to handle FIELD. For now, parse as MOS.
				field := &c.Field{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				t.Fields = append(t.Fields, field)
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassMOS:
				field := &c.Field{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
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
	t, ok := p.unions[body.Name]
	if !ok {
		panic(fmt.Sprintf("unable to locate union %q", body.Name))
	}
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOU:
				field := &c.Field{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				t.Fields = append(t.Fields, field)
			default:
				panic(fmt.Sprintf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassMOU:
				field := &c.Field{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
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
	t, ok := p.enums[body.Name]
	if !ok {
		panic(fmt.Sprintf("unable to locate enum %q", body.Name))
	}
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOE:
				member := &c.EnumMember{
					Value: s.Hdr.Value,
					Name:  body.Name,
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
	base := t.Base()
	u := p.parseBase(base, tag)
	return parseMods(u, t.Mods(), dims)
}

// parseBase parses the SYM base type into the equivalent C type.
func (p *parser) parseBase(base sym.Base, tag string) c.Type {
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
	for _, mod := range mods {
		switch mod {
		case sym.ModPointer:
			t = &c.PointerType{Elem: t}
		case sym.ModFunction:
			// TODO: Figure out how to store params and variadic.
			t = &c.FuncType{
				RetType: t,
			}
		case sym.ModArray:
			for _, dim := range dims {
				t = &c.ArrayType{
					Elem: t,
					Len:  int(dim),
				}
			}
		}
	}
	return t
}
