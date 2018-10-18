package sym

//go:generate stringer -linecomment -type Class

// Class specifies the class of a definition.
type Class uint16

// Definition classes.
const (
	// Storage class auto.
	ClassAUTO Class = 0x0001 // AUTO
	// Storage class extern.
	ClassEXT Class = 0x0002 // EXT
	// Storage class static.
	ClassSTAT Class = 0x0003 // STAT
	// Storage class register.
	ClassREG Class = 0x0004 // REG
	// TODO: Figure out when LABEL is used.
	ClassLABEL Class = 0x0006 // LABEL
	// Member of struct.
	ClassMOS Class = 0x0008 // MOS
	// Function parameter passed on stack.
	ClassARG Class = 0x0009 // ARG
	// Struct tag.
	ClassSTRTAG Class = 0x000A // STRTAG
	// Member of union.
	ClassMOU Class = 0x000B // MOU
	// Union tag.
	ClassUNTAG Class = 0x000C // UNTAG
	// Storage class typedef.
	ClassTPDEF Class = 0x000D // TPDEF
	// Enum tag.
	ClassENTAG Class = 0x000F // ENTAG
	// Member of enum.
	ClassMOE Class = 0x0010 // MOE
	// Function parameter passed in register.
	ClassREGPARM Class = 0x0011 // REGPARM
	// TODO: Figure out when FIELD is used.
	ClassFIELD Class = 0x0012 // FIELD
	// End of symbol.
	ClassEOS Class = 0x0066 // EOS
	Class103 Class = 0x0067 // 103
)
