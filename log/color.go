package log

import "fmt"

var (
	_ ansiColorString = (*AnsiColorGrey)(nil)
	_ ansiColorString = (*AnsiColorRed)(nil)
	_ ansiColorString = (*AnsiColorGreen)(nil)
	_ ansiColorString = (*AnsiColorBlue)(nil)
	_ ansiColorString = (*AnsiColorYellow)(nil)
	_ ansiColorString = (*AnsiColorWhite)(nil)
)

type AnsiColorGrey string

func (grey AnsiColorGrey) AnsiColorString() string {
	return fmt.Sprintf("\x1b[90m%s\x1b[0m", grey)
}

type AnsiColorRed string

func (red AnsiColorRed) AnsiColorString() string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", red)
}

type AnsiColorGreen string

func (green AnsiColorGreen) AnsiColorString() string {
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", green)
}

type AnsiColorBlue string

func (blue AnsiColorBlue) AnsiColorString() string {
	return fmt.Sprintf("\x1b[34m%s\x1b[0m", blue)
}

type AnsiColorYellow string

func (yellow AnsiColorYellow) AnsiColorString() string {
	return fmt.Sprintf("\x1b[33m%s\x1b[0m", yellow)
}

type AnsiColorWhite string

func (white AnsiColorWhite) AnsiColorString() string {
	return fmt.Sprintf("\x1b[37m%s\x1b[0m", white)
}
