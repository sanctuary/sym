package sym

//go:generate stringer -linecomment -type Kind

// Kind specifies the kind of a symbol.
type Kind uint8

// Symbol kinds.
const (
	KindDef     Kind = 0x94 // 94 Def
	KindDef2    Kind = 0x96 // 96 Def2
	KindOverlay Kind = 0x98 // overlay
)
