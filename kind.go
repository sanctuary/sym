package sym

//go:generate stringer -linecomment -type Kind

// Kind specifies the kind of a symbol.
type Kind uint8

// Symbol kinds.
const (
	KindOverlay Kind = 0x98 // overlay
)
