package sym

//go:generate stringer -linecomment -type Class

// Class specifies the class of a definition.
type Class uint16

// Definition classes.
const (
	ClassStructTag Class = 0x000A // STRTAG
	ClassTypedef   Class = 0x000D // TPDEF
)
