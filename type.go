package sym

import (
	"strings"
)

// Type specifies the type of a definition.
//
// A type is made up of a 4-bit basic type specifyer, and a set of 2-bit type
// modifiers.
//
//    Basic type                                            xxxx
//       Modifier                                        xx
//          Modifier                                  xx
//             Modifier                            xx
//                Modifier                      xx
//                   Modifier                xx
//                      Modifier          xx
//
// Example.
//
//    int * f_0064() {}
//
// Interpretation.
//
//    int                                                   0100
//       function                                        10
//          pointer                                   01
//                                                 00
//                                              00
//                                           00
//                                        00
//
//                                 0x64 = 00 00 00 00 01 10 0100
//
// Example.
//
//    int (*v_0094)();
//
// Interpretation.
//
//    int                                                   0100
//       pointer                                         01
//          function                                  10
//                                                 00
//                                              00
//                                           00
//                                        00
//
//                                 0x94 = 00 00 00 00 10 01 0100
type Type uint16

// String returns a string representation of the type.
func (t Type) String() string {
	tMods := t.Mods()
	mods := make([]string, len(tMods))
	for i, mod := range tMods {
		mods[i] = mod.String()
	}
	ss := append(mods, t.Base().String())
	return strings.Join(ss, " ")
}

//go:generate stringer -linecomment -type Base

// Base is a base type.
type Base uint8

// Base types.
const (
	BaseNull   Base = 0x0 // NULL
	BaseVoid   Base = 0x1 // VOID
	BaseChar   Base = 0x2 // CHAR
	BaseShort  Base = 0x3 // SHORT
	BaseInt    Base = 0x4 // INT
	BaseLong   Base = 0x5 // LONG
	BaseFloat  Base = 0x6 // FLOAT
	BaseDouble Base = 0x7 // DOUBLE
	BaseStruct Base = 0x8 // STRUCT
	BaseUnion  Base = 0x9 // UNION
	BaseEnum   Base = 0xA // ENUM
	// TODO: Figure out what MOE is (member of enum?).
	BaseMOE    Base = 0xB // MOE
	BaseUChar  Base = 0xC // UCHAR
	BaseUShort Base = 0xD // USHORT
	BaseUInt   Base = 0xE // UINT
	BaseULong  Base = 0xF // ULONG
)

// Base returns the base type of the type.
func (t Type) Base() Base {
	return Base(t & 0xF)
}

//go:generate stringer -linecomment -type Mod

// Mod is a type modifier.
type Mod uint8

// Type modifiers.
const (
	ModPointer  Mod = 0x1 // PTR
	ModFunction Mod = 0x2 // FCN
	ModArray    Mod = 0x3 // ARY
)

// Mods returns the modifiers of the type.
func (t Type) Mods() []Mod {
	var mods []Mod
	for i := 0; i < 6; i++ {
		// 0b0000000000110000
		shift := uint16(4 + i*2)
		mask := uint16(0x3) << shift
		modMask := Mod((uint16(t) & mask) >> shift)
		if modMask == 0 {
			continue
		}
		if !(modMask >= 0x1 && modMask <= 0x3) {
			panic("unreachable")
		}
		mods = append(mods, modMask)
	}
	return mods
}
