// Package sym provides access to Playstation 1 symbol files (*.SYM).
package sym

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lunixbochs/struc"
	"github.com/pkg/errors"
)

// A File is PS1 symbol file.
type File struct {
	// File header.
	Hdr *FileHeader
	// Symbols.
	Syms []*Symbol
}

// String returns the string representation of the symbol file.
func (f *File) String() string {
	buf := &strings.Builder{}
	offset := 0
	buf.WriteString(f.Hdr.String())
	offset += binary.Size(*f.Hdr)
	for _, sym := range f.Syms {
		fmt.Fprintf(buf, "%06x: %s\n", offset, sym)
		offset += sym.Size()
	}
	return buf.String()
}

// A FileHeader is a PS1 symbol file header.
type FileHeader struct {
	// File signature; MND.
	Signature [3]byte `struc:"[3]byte"`
	// File format version.
	Version uint8 `struc:"uint8,little"`
	// Target unit.
	TargetUnit uint32 `struc:"uint32,little"`
}

// String returns the string representation of the symbol file header.
func (hdr *FileHeader) String() string {
	const format = `
Header : %s version %d
Target unit %d
`
	return fmt.Sprintf(format, hdr.Signature, hdr.Version, hdr.TargetUnit)
}

// ParseFile parses the given PS1 symbol file.
func ParseFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	return Parse(f)
}

// ParseBytes parses the given PS1 symbol file, reading from b.
func ParseBytes(b []byte) (*File, error) {
	return Parse(bytes.NewReader(b))
}

// Parse parses the given PS1 symbol file, reading from r.
func Parse(r io.Reader) (*File, error) {
	// Parse file header.
	f := &File{}
	br := bufio.NewReader(r)
	hdr, err := parseFileHeader(br)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	f.Hdr = hdr

	// Parse symbols.
	for {
		sym, err := parseSymbol(br)
		if err != nil {
			if errors.Cause(err) == io.EOF {
				break
			}
			return f, errors.WithStack(err)
		}
		f.Syms = append(f.Syms, sym)
	}
	return f, nil
}

// parseFileHeader parses and returns a PS1 symbol file header.
func parseFileHeader(r io.Reader) (*FileHeader, error) {
	hdr := &FileHeader{}
	if err := struc.Unpack(r, &hdr); err != nil {
		return nil, errors.WithStack(err)
	}
	// Verify Smacker signature.
	switch string(hdr.Signature[:]) {
	case "MND":
		// valid signature.
	default:
		return nil, errors.Errorf(`invalid SYM signature; expected "MND", got %q`, hdr.Signature)
	}
	return hdr, nil
}
