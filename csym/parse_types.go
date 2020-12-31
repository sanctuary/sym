package csym

import (
	"fmt"
	"strings"

	"github.com/sanctuary/sym"
	"github.com/sanctuary/sym/csym/c"
)

// ParseTypes parses the SYM types into the equivalent C types.
func (p *Parser) ParseTypes(syms []*sym.Symbol) {
	p.initTaggedTypes(syms)
	// Parse symbols.
	for i := 0; i < len(syms); i++ {
		s := syms[i]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassSTRTAG:
				n := p.parseStructTag(body, syms[i+1:])
				i += n
			case sym.ClassUNTAG:
				n := p.parseUnionTag(body, syms[i+1:])
				i += n
			case sym.ClassENTAG:
				n := p.parseEnumTag(body, syms[i+1:])
				i += n
			case sym.ClassTPDEF:
				// TODO: Replace with parseDef?
				p.parseTypedef(body.Type, nil, "", body.Name)
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassTPDEF:
				// TODO: Replace with parseDef?
				p.parseTypedef(body.Type, body.Dims, body.Tag, body.Name)
			}
		}
	}
}

// initTaggedTypes adds scaffolding types for structs, unions and enums.
func (p *Parser) initTaggedTypes(syms []*sym.Symbol) {
	// Psy-Q SDK outputs several types including bool as NULL, most likely
	// the object was a bool so assume that.
	boolDef := &c.VarDecl{
		Class: c.Typedef,
		Var: c.Var{
			Type: c.Int,
			Name: "bool",
		},
	}
	p.Types["bool"] = boolDef
	// Add scaffolding types for structs, unions and enums, so they may be
	// referrenced before defined.
	vtblPtrType := &c.StructType{
		Tag: "__vtbl_ptr_type",
	}
	p.Structs["__vtbl_ptr_type"] = vtblPtrType
	p.StructTags = append(p.StructTags, "__vtbl_ptr_type")
	var (
		structTags = make(map[string]bool)
		unionTags  = make(map[string]bool)
		enumTags   = make(map[string]bool)
	)
	for _, s := range syms {
		switch body := s.Body.(type) {
		case *sym.Def:
			tag := validName(body.Name)
			switch body.Class {
			case sym.ClassSTRTAG:
				tag = uniqueTag(tag, structTags)
				t := &c.StructType{
					Size: body.Size,
					Tag:  tag,
				}
				p.Structs[tag] = t
				p.StructTags = append(p.StructTags, tag)
			case sym.ClassUNTAG:
				tag = uniqueTag(tag, unionTags)
				t := &c.UnionType{
					Size: body.Size,
					Tag:  tag,
				}
				p.Unions[tag] = t
				p.UnionTags = append(p.UnionTags, tag)
			case sym.ClassENTAG:
				tag = uniqueTag(tag, enumTags)
				t := &c.EnumType{
					Tag: tag,
				}
				p.Enums[tag] = t
				p.EnumTags = append(p.EnumTags, tag)
			}
		}
	}
}

// parseStructTag parses a struct tag sequence of symbols.
func (p *Parser) parseStructTag(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseStruct {
		panic(fmt.Errorf("support for base type %q not yet implemented", base))
	}
	tag := validName(body.Name)
	t := findStruct(p, tag, body.Size)
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOS:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Var: c.Var{
						Type: p.parseType(body.Type, nil, ""),
						Name: validName(body.Name),
					},
				}
				t.Fields = append(t.Fields, field)
			case sym.ClassFIELD:
				// TODO: Figure out what FIELD represents. Use method for now.
				method := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Var: c.Var{
						Type: p.parseType(body.Type, nil, ""),
						Name: validName(body.Name),
					},
				}
				t.Methods = append(t.Methods, method)
			default:
				panic(fmt.Errorf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassMOS:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Var: c.Var{
						Type: p.parseType(body.Type, body.Dims, body.Tag),
						Name: validName(body.Name),
					},
				}
				t.Fields = append(t.Fields, field)
			case sym.ClassEOS:
				return n + 1
			default:
				panic(fmt.Errorf("support for class %q not yet implemented", body.Class))
			}
		}
	}
	panic("unreachable")
}

// parseUnionTag parses a union tag sequence of symbols.
func (p *Parser) parseUnionTag(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseUnion {
		panic(fmt.Errorf("support for base type %q not yet implemented", base))
	}
	tag := validName(body.Name)
	t := findUnion(p, tag, body.Size)
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOU:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Var: c.Var{
						Type: p.parseType(body.Type, nil, ""),
						Name: validName(body.Name),
					},
				}
				t.Fields = append(t.Fields, field)
			default:
				panic(fmt.Errorf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassMOU:
				field := c.Field{
					Offset: s.Hdr.Value,
					Size:   body.Size,
					Var: c.Var{
						Type: p.parseType(body.Type, body.Dims, body.Tag),
						Name: validName(body.Name),
					},
				}
				t.Fields = append(t.Fields, field)
			case sym.ClassEOS:
				return n + 1
			default:
				panic(fmt.Errorf("support for class %q not yet implemented", body.Class))
			}
		}
	}
	panic("unreachable")
}

// parseEnumTag parses an enum tag sequence of symbols.
func (p *Parser) parseEnumTag(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseEnum {
		panic(fmt.Errorf("support for base type %q not yet implemented", base))
	}
	tag := validName(body.Name)
	t := findEnum(p, tag)
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassMOE:
				name := validName(body.Name)
				name = uniqueEnum(name, p.enumMembers)
				member := &c.EnumMember{
					Value: s.Hdr.Value,
					Name:  name,
				}
				t.Members = append(t.Members, member)
			default:
				panic(fmt.Errorf("support for class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassEOS:
				return n + 1
			default:
				panic(fmt.Errorf("support for class %q not yet implemented", body.Class))
			}
		}
	}
	panic("unreachable")
}

// parseTypedef parses a typedef symbol.
func (p *Parser) parseTypedef(t sym.Type, dims []uint32, tag, name string) {
	name = validName(name)
	def := &c.VarDecl{
		Class: c.Typedef,
		Var: c.Var{
			Type: p.parseType(t, dims, tag),
			Name: name,
		},
	}
	p.Typedefs = append(p.Typedefs, def)
	p.Types[name] = def
}

// ### [ Helper functions ] ####################################################

// Duplicate tag format string.
const duplicateTagFormat = "%s_duplicate_%d"

// uniqueTag returns a unique tag based on the given tag and set of present
// tags.
func uniqueTag(tag string, tags map[string]bool) string {
	newTag := tag
	for i := 0; tags[newTag]; i++ {
		newTag = fmt.Sprintf(duplicateTagFormat, tag, i)
	}
	tags[newTag] = true
	return newTag
}

// Duplicate enum member format string.
const duplicateEnumFormat = "%s_DUPLICATE_%d"

// uniqueEnum returns a unique enum member based on the given tag and set of
// present enum members.
func uniqueEnum(name string, members map[string]bool) string {
	newName := name
	for i := 0; members[newName]; i++ {
		newName = fmt.Sprintf(duplicateEnumFormat, name, i)
	}
	members[newName] = true
	return newName
}

// findStruct returns the struct with the given tag and size.
func findStruct(p *Parser, tag string, size uint32) *c.StructType {
	newTag := tag
	for i := 0; ; i++ {
		t, ok := p.Structs[newTag]
		if !ok {
			panic(fmt.Errorf("unable to locate struct %q", tag))
		}
		if t.Size == size && len(t.Fields) == 0 {
			return t
		}
		newTag = fmt.Sprintf(duplicateTagFormat, tag, i)
	}
}

// findUnion returns the union with the given tag and size.
func findUnion(p *Parser, tag string, size uint32) *c.UnionType {
	newTag := tag
	for i := 0; ; i++ {
		t, ok := p.Unions[newTag]
		if !ok {
			panic(fmt.Errorf("unable to locate union %q", tag))
		}
		if t.Size == size && len(t.Fields) == 0 {
			return t
		}
		newTag = fmt.Sprintf(duplicateTagFormat, tag, i)
	}
}

// findEnum returns the enum with the given tag.
func findEnum(p *Parser, tag string) *c.EnumType {
	newTag := tag
	for i := 0; ; i++ {
		t, ok := p.Enums[newTag]
		if !ok {
			panic(fmt.Errorf("unable to locate enum %q", tag))
		}
		if len(t.Members) == 0 {
			return t
		}
		newTag = fmt.Sprintf(duplicateTagFormat, tag, i)
	}
}

// parseType parses the SYM type into the equivalent C type.
func (p *Parser) parseType(t sym.Type, dims []uint32, tag string) c.Type {
	u := p.parseBase(t.Base(), tag)
	return parseMods(u, t.Mods(), dims)
}

// parseBase parses the SYM base type into the equivalent C type.
func (p *Parser) parseBase(base sym.Base, tag string) c.Type {
	tag = validName(tag)
	switch base {
	case sym.BaseNull:
		return p.Types["bool"]
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
		t, ok := p.Structs[tag]
		if !ok {
			panic(fmt.Errorf("unable to locate struct %q", tag))
		}
		return t
	case sym.BaseUnion:
		t, ok := p.Unions[tag]
		if !ok {
			panic(fmt.Errorf("unable to locate union %q", tag))
		}
		return t
	case sym.BaseEnum:
		t, ok := p.Enums[tag]
		if !ok {
			panic(fmt.Errorf("unable to locate enum %q", tag))
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
		panic(fmt.Errorf("base type %q not yet supported", base))
	}
}

// parseMods parses the SYM type modifiers into the equivalent C type modifiers.
func parseMods(t c.Type, mods []sym.Mod, dims []uint32) c.Type {
	j := 0
	for i := len(mods) - 1; i >= 0; i-- {
		mod := mods[i]
		switch mod {
		case sym.ModPointer:
			t = &c.PointerType{Elem: t}
		case sym.ModFunction:
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
