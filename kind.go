package sym

//go:generate stringer -linecomment -type Kind

// Kind specifies the kind of a symbol.
type Kind uint8

// Symbol kinds.
const (
	KindName1      Kind = 0x01 // 1
	KindName2      Kind = 0x02 // 2
	KindIncSLD     Kind = 0x80 // 80
	KindIncSLDByte Kind = 0x82 // 82
	KindIncSLDWord Kind = 0x84 // 84
	KindSetSLD     Kind = 0x86 // 86
	KindSetSLD2    Kind = 0x88 // 88
	KindEndSLD     Kind = 0x8A // 8a
	KindDef        Kind = 0x94 // 94 Def
	KindDef2       Kind = 0x96 // 96 Def2
	KindOverlay    Kind = 0x98 // overlay
)
