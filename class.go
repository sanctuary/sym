package sym

//go:generate stringer -linecomment -type Class

// Class specifies the class of a definition.
type Class uint16

// Definition classes.
const (
	ClassEXT     Class = 0x0002 // EXT
	ClassSTAT    Class = 0x0003 // STAT
	ClassMOS     Class = 0x0008 // MOS
	ClassSTRTAG  Class = 0x000A // STRTAG
	ClassMOU     Class = 0x000B // MOU
	ClassUNTAG   Class = 0x000C // UNTAG
	ClassTPDEF   Class = 0x000D // TPDEF
	ClassENTAG   Class = 0x000F // ENTAG
	ClassMOE     Class = 0x0010 // MOE
	ClassREGPARM Class = 0x0011 // REGPARM
	ClassFIELD   Class = 0x0012 // FIELD
	ClassEOS     Class = 0x0066 // EOS
)
