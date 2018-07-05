package sym

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/lunixbochs/struc"
	"github.com/pkg/errors"
)

// A Symbol is a PS1 symbol.
type Symbol struct {
	// Symbol header.
	Hdr *SymbolHeader
	// Symbol body.
	Body SymbolBody
}

// String returns the string representation of the symbol.
func (sym *Symbol) String() string {
	return fmt.Sprintf("%v %v", sym.Hdr, sym.Body)
}

// Size returns the size of the symbol in bytes.
func (sym *Symbol) Size() int {
	return binary.Size(*sym.Hdr) + sym.Body.Size()
}

// A SymbolHeader is a PS1 symbol header.
type SymbolHeader struct {
	// Address or value of symbol.
	Value uint32 `struc:"uint32,little"`
	// Symbol kind; specifies type of symbol body.
	Kind Kind `struc:"uint8,little"`
}

// String returns the string representation of the symbol header.
func (hdr *SymbolHeader) String() string {
	return fmt.Sprintf("$%08x %v", hdr.Value, hdr.Kind)
}

// SymbolBody is the sum-type of all symbol bodies.
type SymbolBody interface {
	// Size returns the size of the symbol body in bytes.
	Size() int
}

// parseSymbol parses and returns a PS1 symbol.
func parseSymbol(r io.Reader) (*Symbol, error) {
	// Parse symbol header.
	sym := &Symbol{}
	hdr, err := parseSymbolHeader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sym.Hdr = hdr

	// Parse symbol body.
	body, err := parseSymbolBody(r, hdr.Kind)
	if err != nil {
		return sym, errors.WithStack(err)
	}
	sym.Body = body
	return sym, nil
}

// parseSymbolHeader parses and returns a PS1 symbol header.
func parseSymbolHeader(r io.Reader) (*SymbolHeader, error) {
	hdr := &SymbolHeader{}
	if err := struc.Unpack(r, &hdr); err != nil {
		return nil, errors.WithStack(err)
	}
	return hdr, nil
}

// parseSymbolBody parses and returns a PS1 symbol body.
func parseSymbolBody(r io.Reader, kind Kind) (SymbolBody, error) {
	parse := func(body SymbolBody) (SymbolBody, error) {
		if err := struc.Unpack(r, body); err != nil {
			return nil, errors.WithStack(err)
		}
		return body, nil
	}
	switch kind {
	case KindOverlay:
		return parse(&Overlay{})
	default:
		return nil, errors.Errorf("support for symbol kind 0x%02X not yet implemented", uint8(kind))
	}
}

// An Overlay specifies the length and id of an file overlay (e.g. a shared
// library).
//
// Value of the symbol header specifies the base address at which the overlay is
// loaded.
type Overlay struct {
	// Overlay length in bytes.
	Length uint32 `struc:"uint32,little"`
	// Overlay ID.
	ID uint32 `struc:"uint32,little"`
}

// String returns the string representation of the overlay symbol.
func (body *Overlay) String() string {
	return fmt.Sprintf("length $%08x id $%x", body.Length, body.ID)
}

// Size returns the size of the symbol body in bytes.
func (body *Overlay) Size() int {
	return binary.Size(*body)
}
