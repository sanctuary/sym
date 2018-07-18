package sym

import (
	"strings"
)

// Type specifies the type of a definition.
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
	BaseNull  Base = 0x0 // NULL
	BaseVoid  Base = 0x1 // VOID
	BaseChar  Base = 0x2 // CHAR
	BaseShort Base = 0x3 // SHORT
	BaseInt   Base = 0x4 // INT
	BaseLong  Base = 0x5 // LONG
	// TODO: 0x6
	// TODO: 0x7
	BaseStruct Base = 0x8 // STRUCT
	BaseUnion  Base = 0x9 // UNION
	BaseEnum   Base = 0xA // ENUM
	// TODO: Figure out what MOE is.
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
