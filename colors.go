package logger

type Color int64

const (
	RED Color = iota
	GREEN
	GRAY
	WHITE
	YELLOW
	PURPLE
	BLUE
)

// Returns color as a string
func (c Color) Color() string {
	switch c {
	case RED:
		return "\033[31m"
	case GREEN:
		return "\033[32m"
	case GRAY:
		return "\033[37m"
	case WHITE:
		return "\033[97m"
	case YELLOW:
		return "\033[33m"
	case PURPLE:
		return "\033[35m"
	case BLUE:
		return "\033[34m"
	}
	return ""
}

// colorWrap wraps a string in a color
func colorWrap(c Color, m string) string {
	const Reset = "\033[0m"
	return c.Color() + m + Reset
}
