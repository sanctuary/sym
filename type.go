package sym

import (
	"fmt"
	"strings"
)

// Type specifies the type of a definition.
type Type uint16

// baseName maps from base type to name.
var baseName = map[uint16]string{
	0x0: "NULL",
	0x1: "VOID",
	0x2: "CHAR",
	0x3: "SHORT",
	0x4: "INT",
	0x5: "LONG",
	// TODO: 0x6
	// TODO: 0x7
	0x8: "STRUCT",
	// TODO: 0x9
	0xA: "ENUM",
	0xB: "MOE",
	0xC: "UCHAR",
	0xD: "USHORT",
	0xE: "UINT",
	0xF: "ULONG",
}

// modName maps from type modifier to name.
var modName = map[uint16]string{
	// 0b01
	0x1: "PTR",
	// 0b10
	0x2: "FCN",
	// 0b11
	0x3: "ARY",
}

// String returns a string representation of the type.
func (t Type) String() string {
	baseMask := uint16(t & 0xF)
	base, ok := baseName[baseMask]
	if !ok {
		base = fmt.Sprintf("%04X", baseMask)
	}
	var mods []string
	for i := 0; i < 6; i++ {
		// 0b0000000000110000
		shift := uint16(4 + i*2)
		mask := uint16(0x3) << shift
		modMask := (uint16(t) & mask) >> shift
		if modMask == 0 {
			continue
		}
		if !(modMask >= 0x1 && modMask <= 0x3) {
			panic("unreachable")
		}
		mod := modName[modMask]
		mods = append(mods, mod)
	}
	ss := append(mods, base)
	return strings.Join(ss, " ")
}
