package types

// T defines the type of the configuration option, essential when setting
// flags, converting from interfaces.
type T uint64

const (
	Nul T = iota
	Int
	Float
	String
	Bool
	Duration
)
