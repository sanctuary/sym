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

// parse parses the SYM file into equivalent C declarations.
func parse(f *sym.File) *parser {
	p := newParser()
	p.parseTypes(f.Syms)
	p.parseDecls(f.Syms)
	return p
}

// parseDecls parses the SYM symbols into the equivalent C declarations.
func (p *parser) parseDecls(syms []*sym.Symbol) {
	var curLine Line
	var curOverlay *Overlay
	for i := 0; i < len(syms); i++ {
		s := syms[i]
		switch body := s.Body.(type) {
		case *sym.Name1:
			symbol := &Symbol{
				Addr: s.Hdr.Value,
				Name: body.Name,
			}
			p.symbols = append(p.symbols, symbol)
		case *sym.Name2:
			symbol := &Symbol{
				Addr: s.Hdr.Value,
				Name: body.Name,
			}
			p.symbols = append(p.symbols, symbol)
		case *sym.IncSLD:
			curLine.Line++
			line := &Line{
				Addr: s.Hdr.Value,
				Path: curLine.Path,
				Line: curLine.Line,
			}
			p.lines = append(p.lines, line)
		case *sym.IncSLDByte:
			curLine.Line += uint32(body.Inc)
			line := &Line{
				Addr: s.Hdr.Value,
				Path: curLine.Path,
				Line: curLine.Line,
			}
			p.lines = append(p.lines, line)
		case *sym.IncSLDWord:
			curLine.Line += uint32(body.Inc)
			line := &Line{
				Addr: s.Hdr.Value,
				Path: curLine.Path,
				Line: curLine.Line,
			}
			p.lines = append(p.lines, line)
		case *sym.SetSLD:
			// TODO: reset curLine.Path?
			curLine.Line = body.Line
			line := &Line{
				Addr: s.Hdr.Value,
				Path: curLine.Path,
				Line: curLine.Line,
			}
			p.lines = append(p.lines, line)
		case *sym.SetSLD2:
			curLine = Line{
				Path: body.Path,
				Line: body.Line,
			}
			line := &Line{
				Addr: s.Hdr.Value,
				Path: curLine.Path,
				Line: curLine.Line,
			}
			p.lines = append(p.lines, line)
		case *sym.EndSLD:
			curLine = Line{}
		case *sym.FuncStart:
			n := p.parseFunc(s.Hdr.Value, body, syms[i+1:])
			i += n
		case *sym.Def:
			switch body.Class {
			case sym.ClassEXT, sym.ClassSTAT:
				// TODO: figure out how EXT and STAT differ.
				p.parseClassEXT(s.Hdr.Value, body.Size, body.Type, nil, "", body.Name)
			case sym.ClassMOS, sym.ClassSTRTAG, sym.ClassMOU, sym.ClassUNTAG, sym.ClassTPDEF, sym.ClassENTAG, sym.ClassMOE, sym.ClassFIELD:
				// nothing to do.
			default:
				panic(fmt.Sprintf("support for symbol class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassEXT, sym.ClassSTAT:
				// TODO: figure out how EXT and STAT differ.
				p.parseClassEXT(s.Hdr.Value, body.Size, body.Type, body.Dims, body.Tag, body.Name)
			case sym.ClassMOS, sym.ClassMOU, sym.ClassTPDEF, sym.ClassMOE, sym.ClassFIELD, sym.ClassEOS:
				// nothing to do.
			default:
				panic(fmt.Sprintf("support for symbol class %q not yet implemented", body.Class))
			}
		case *sym.Overlay:
			p.parseOverlay(s.Hdr.Value, body)
		case *sym.SetOverlay:
			overlay, ok := p.overlayIDs[s.Hdr.Value]
			if !ok {
				panic(fmt.Errorf("unable to locate overlay with ID %x", s.Hdr.Value))
			}
			curOverlay = overlay
			_ = curOverlay // TODO: Remove.
		default:
			panic(fmt.Sprintf("support for symbol type %T not yet implemented", body))
		}
	}
}

// parseFunc parses a function sequence of symbols.
func (p *parser) parseFunc(addr uint32, body *sym.FuncStart, syms []*sym.Symbol) (n int) {
	name := validName(body.Name)
	f, ok := p.funcNames[name]
	if !ok {
		panic(fmt.Errorf("unable to locate function %q", name))
	}
	funcType, ok := f.Type.(*c.FuncType)
	if !ok {
		panic(fmt.Errorf("invalid function type; expected *c.FuncType, got %T", f.Type))
	}
	if f.LineStart != 0 {
		// Ignore function duplicate.
		for n = 0; n < len(syms); n++ {
			s := syms[n]
			switch s.Body.(type) {
			case *sym.FuncEnd:
				return n + 1
			}
		}
	}
	f.LineStart = body.Line
	var blocks blockStack
	var curBlock *c.Block
	for n = 0; n < len(syms); n++ {
		s := syms[n]
		switch body := s.Body.(type) {
		case *sym.FuncEnd:
			f.LineEnd = body.Line
			return n + 1
		case *sym.BlockStart:
			if curBlock != nil {
				blocks.push(curBlock)
			}
			block := &c.Block{
				LineStart: body.Line,
			}
			f.Blocks = append(f.Blocks, block)
			curBlock = block
		case *sym.BlockEnd:
			curBlock.LineEnd = body.Line
			if !blocks.empty() {
				curBlock = blocks.pop()
			}
		case *sym.Def:
			switch body.Class {
			case sym.ClassAUTO:
				v := c.Var{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					// Parameter already added for ClassARG.
					//funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassSTAT:
				v := c.Var{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassREG:
				v := c.Var{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					// Parameter already added for ClassARG.
					//funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassLABEL:
				v := c.Var{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassARG:
				v := c.Var{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassREGPARM:
				v := c.Var{
					Type: p.parseType(body.Type, nil, ""),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			default:
				panic(fmt.Errorf("support for symbol class %q not yet implemented", body.Class))
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassAUTO:
				v := c.Var{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					// Parameter already added for ClassARG.
					//funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassSTAT:
				v := c.Var{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassREG:
				v := c.Var{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					// Parameter already added for ClassARG.
					//funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassLABEL:
				v := c.Var{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassARG:
				v := c.Var{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			case sym.ClassREGPARM:
				v := c.Var{
					Type: p.parseType(body.Type, body.Dims, body.Tag),
					Name: body.Name,
				}
				if curBlock != nil {
					curBlock.Locals = append(curBlock.Locals, v)
				} else {
					funcType.Params = append(funcType.Params, v)
				}
			default:
				panic(fmt.Errorf("support for symbol class %q not yet implemented", body.Class))
			}
		default:
			panic(fmt.Errorf("support for symbol type %T not yet implemented", body))
		}
	}
	panic("unreachable")
}

// parseOverlay parses an overlay symbol.
func (p *parser) parseOverlay(addr uint32, body *sym.Overlay) {
	overlay := &Overlay{
		Addr:      addr,
		ID:        body.ID,
		Length:    body.Length,
		varNames:  make(map[string]*c.VarDecl),
		funcNames: make(map[string]*c.FuncDecl),
	}
	p.overlays = append(p.overlays, overlay)
	p.overlayIDs[overlay.ID] = overlay
}

// parseClassEXT parses an EXT symbol.
func (p *parser) parseClassEXT(addr, size uint32, t sym.Type, dims []uint32, tag, name string) {
	name = validName(name)
	if _, ok := p.varNames[name]; ok {
		name = duplicatePrefix + name
	} else if _, ok := p.funcNames[name]; ok {
		name = duplicatePrefix + name
	}
	cType := p.parseType(t, dims, tag)
	if funcType, ok := cType.(*c.FuncType); ok {
		f := &c.FuncDecl{
			Addr: addr,
			Size: size,
			Var: c.Var{
				Type: funcType,
				Name: name,
			},
		}
		p.funcs = append(p.funcs, f)
		p.funcNames[name] = f
	} else {
		v := &c.VarDecl{
			Addr: addr,
			Size: size,
			Var: c.Var{
				Type: cType,
				Name: name,
			},
		}
		p.vars = append(p.vars, v)
		p.varNames[name] = v
	}
}

// parseTypes parses the SYM types into the equivalent C types.
func (p *parser) parseTypes(syms []*sym.Symbol) {
	p.initTaggedTypes(syms)
	// Parse symbols.
	for i := 0; i < len(syms); i++ {
		s := syms[i]
		// TODO: remove debug output once C output is mature.
		//fmt.Fprintln(os.Stderr, "sym:", s)
		switch body := s.Body.(type) {
		case *sym.Def:
			switch body.Class {
			case sym.ClassSTRTAG:
				n := p.parseClassSTRTAG(body, syms[i+1:])
				i += n
			case sym.ClassUNTAG:
				n := p.parseClassUNTAG(body, syms[i+1:])
				i += n
			case sym.ClassTPDEF:
				p.parseClassTPDEF(body.Type, nil, "", body.Name)
			case sym.ClassENTAG:
				n := p.parseClassENTAG(body, syms[i+1:])
				i += n
			default:
				//dbg.Printf("support for class %q not yet implemented", body.Class)
			}
		case *sym.Def2:
			switch body.Class {
			case sym.ClassTPDEF:
				p.parseClassTPDEF(body.Type, body.Dims, body.Tag, body.Name)
			default:
				//dbg.Printf("support for class %q not yet implemented", body.Class)
			}
		case *sym.Overlay:
		// nothing to do.
		default:
			//dbg.Printf("support for symbol body %T not yet implemented", body)
		}
	}
}

// initTaggedTypes adds scaffolding types for structs, unions and enums.
func (p *parser) initTaggedTypes(syms []*sym.Symbol) {
	// Add scaffolding types for structs, unions and enums, so they may be
	// referrenced before defined.
	vtblPtrType := &c.StructType{
		Tag: "__vtbl_ptr_type",
	}
	p.structs["__vtbl_ptr_type"] = vtblPtrType
	p.structTags = append(p.structTags, "__vtbl_ptr_type")
	// Bool used for NULL type.
	boolDef := &c.Typedef{
		Type: c.Int,
		Name: "bool",
	}
	p.types["bool"] = boolDef
	var (
		uniqueStruct = make(map[string]bool)
		uniqueUnion  = make(map[string]bool)
		uniqueEnum   = make(map[string]bool)
	)
	for _, s := range syms {
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
}

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

	// Declarations.
	*Overlay // default binary

	// Overlays.
	overlays []*Overlay
	// overlayIDs maps from overlay ID to overlay.
	overlayIDs map[uint32]*Overlay
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

// newParser returns a new parser.
func newParser() *parser {
	return &parser{
		structs:          make(map[string]*c.StructType),
		unions:           make(map[string]*c.UnionType),
		enums:            make(map[string]*c.EnumType),
		types:            make(map[string]*c.Typedef),
		uniqueEnumMember: make(map[string]bool),
		Overlay: &Overlay{
			varNames:  make(map[string]*c.VarDecl),
			funcNames: make(map[string]*c.FuncDecl),
		},
		overlayIDs: make(map[uint32]*Overlay),
	}
}

// parseClassSTRTAG parses a struct tag sequence of symbols.
func (p *parser) parseClassSTRTAG(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseStruct {
		panic(fmt.Errorf("support for base type %q not yet implemented", base))
	}
	name := validName(body.Name)
	t, ok := p.structs[name]
	if !ok {
		panic(fmt.Errorf("unable to locate struct %q", name))
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
			case sym.ClassMOS, sym.ClassFIELD:
				// TODO: Figure out how to handle FIELD. For now, parse as MOS.
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

// parseClassUNTAG parses a union tag sequence of symbols.
func (p *parser) parseClassUNTAG(body *sym.Def, syms []*sym.Symbol) (n int) {
	if base := body.Type.Base(); base != sym.BaseUnion {
		panic(fmt.Errorf("support for base type %q not yet implemented", base))
	}
	name := validName(body.Name)
	t, ok := p.unions[name]
	if !ok {
		panic(fmt.Errorf("unable to locate union %q", name))
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

// parseClassTPDEF parses a typedef symbol.
func (p *parser) parseClassTPDEF(t sym.Type, dims []uint32, tag, name string) {
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
		panic(fmt.Errorf("support for base type %q not yet implemented", base))
	}
	name := validName(body.Name)
	t, ok := p.enums[name]
	if !ok {
		panic(fmt.Errorf("unable to locate enum %q", name))
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

// ### [ Helper functions ] ####################################################

// parseType parses the SYM type into the equivalent C type.
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
			panic(fmt.Errorf("unable to locate struct %q", tag))
		}
		return t
	case sym.BaseUnion:
		t, ok := p.unions[tag]
		if !ok {
			panic(fmt.Errorf("unable to locate union %q", tag))
		}
		return t
	case sym.BaseEnum:
		t, ok := p.enums[tag]
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

// blockStack is a stack of blocks.
type blockStack []*c.Block

// push pushes the block onto the stack.
func (b *blockStack) push(block *c.Block) {
	*b = append(*b, block)
}

// pop pops the block from the stack.
func (b *blockStack) pop() *c.Block {
	if b.empty() {
		panic("invalid pop call; empty stack")
	}
	block := (*b)[len(*b)-1]
	*b = (*b)[:len(*b)-1]
	return block
}

// empty reports if the stack is empty.
func (b *blockStack) empty() bool {
	return len(*b) == 0
}
