package log

import "testing"

func TestAnsiColorString(t *testing.T) {
	t.Log(AnsiColorBlue("blue").AnsiColorString())
	t.Log(AnsiColorGrey("grey").AnsiColorString())
	t.Log(AnsiColorRed("red").AnsiColorString())
	t.Log(AnsiColorYellow("yellow").AnsiColorString())
}
