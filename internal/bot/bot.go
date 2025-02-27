package bot

import "math"

const (
	// MaxSizeEasy is the maximum board size supported by the built-in easy bot
	MaxSizeEasy uint8 = math.MaxUint8
	// MaxSizeHard is the maximum board size supported by the built-in hard bot
	MaxSizeHard uint8 = math.MaxUint8
	// MaxSizeImpossible is the maximum board size supported by the built-in impossible bot
	MaxSizeImpossible uint8 = 3
	// MaxSizeNormal is the maximum board size supported by the built-in normal bot
	MaxSizeNormal uint8 = math.MaxUint8

	// NameEasy is the name of the built-in easy bot
	NameEasy = "easy"
	// NameHard is the name of the built-in hard bot
	NameHard = "hard"
	// NameImpossible is the name of the built-in impossible bot
	NameImpossible = "impossible"
	// NameNormal is the name of the built-in normal bot
	NameNormal = "normal"
)
