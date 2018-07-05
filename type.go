package sym

//go:generate stringer -linecomment -type Type

// Type specifies the type of a definition.
type Type uint16

// Definition types.
const (
	TypeStruct Type = 0x0008 // STRUCT
	TypeUChar  Type = 0x000C // UCHAR
	TypeUShort Type = 0x000D // USHORT
	TypeUInt   Type = 0x000E // UINT
	TypeULong  Type = 0x000F // ULONG
)
