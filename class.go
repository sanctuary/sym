package sym

//go:generate stringer -linecomment -type Class

// Class specifies the class of a definition.
type Class uint16

// Definition classes.
const (
	ClassMOS    Class = 0x0008 // MOS
	ClassSTRTAG Class = 0x000A // STRTAG
	ClassTPDEF  Class = 0x000D // TPDEF
	ClassENTAG  Class = 0x000F // ENTAG
	ClassMOE    Class = 0x0010 // MOE
	ClassEOS    Class = 0x0066 // EOS
)
