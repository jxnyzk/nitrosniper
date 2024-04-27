package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/inancgumus/screen"
	"golang.org/x/term"
)

// don't judge
// don't touch
// don't copy
// don't look
// don't ask

var (
	colorsBulletPoint = "131;131;131"
	colorsGeneral     = "220;220;220"
	colorsVarKey      = "131;131;131"
	colorsVarValue    = "176;176;176"
	colorsSpinner     = "53;96;176"
	loggerMutex       sync.Mutex
)

const (
	// StartSet chars
	startSet = "\x1b["

	// ResetSet close all properties.
	resetSet = "\x1b[0m"

	// Foreground
	fgRGBPfx = "38;2;"

	// Background
	bgRGBPfx = "48;2;"
)

type LogField struct {
	Key   string
	Value string
}

func FieldString(key string, value string) LogField {
	return LogField{
		Key:   key,
		Value: value,
	}
}

func FieldInt(key string, value int) LogField {
	return LogField{
		Key:   key,
		Value: strconv.Itoa(value),
	}
}

func FieldFloat32(key string, value float32) LogField {
	return LogField{
		Key:   key,
		Value: fmt.Sprintf("%f", value),
	}
}

func FieldFloat64(key string, value float64) LogField {
	return LogField{
		Key:   key,
		Value: fmt.Sprintf("%f", value),
	}
}

func FieldAny(key string, value any) LogField {
	return LogField{
		Key:   key,
		Value: fmt.Sprintf("%v", value),
	}
}

// parses if it's a hex to "r;g;b;a"
func parseColorCode(code string) string {
	if code[0] == '#' {
		code = code[1:]
	}

	var r, g, b uint8
	values, err := strconv.ParseUint(code, 16, 32)
	if err != nil {
		return code
	}

	r = uint8(values >> 16)
	g = uint8((values >> 8) & 0xFF)
	b = uint8(values & 0xFF)


	return fmt.Sprintf("%d;%d;%d", int(r), int(g), int(b))
}

func colorStringForeground(code string, str string) string {
	return startSet + fgRGBPfx + parseColorCode(code) + "m" + str + resetSet
}


func resolveFields(fields []LogField) string {
	var ret = ""
	if len(fields) > 0 {
		for i, v := range fields {
			ret += colorStringForeground(colorsVarKey, v.Key) + "=" + THEME+ v.Value
			if i < len(fields) {
				ret += " "
			}
		}
	}

	return ret
}

func clr(color, text string) string {
	return color + text
}

var (
	THEME = "\x1b[38;2;168;70;212m"
	INF = clr(THEME, "+")
	WRN = clr("\x1b[38;5;203m", "!")
	FTL = clr("\x1b[38;5;209m", "-")
	SUC = clr("\x1b[38;5;127m", "*")
)

func helperLog(logType string, description string, fields ...LogField) {
	loggerMutex.Lock()
	
	fmt.Println(clr("\033[90m", time.Now().Format("15:04:05")) + " \033[0m[" + logType +  "\033[0m] " + clr("\033[0m", description) + " " + resolveFields(fields) + strings.Repeat(" ", 40))

	loggerMutex.Unlock() 
}

func Info(description string, fields ...LogField) {
	helperLog(INF, description, fields...)
}

func Warn(description string, fields ...LogField) {
	helperLog(WRN, description, fields...)
}

func Fail(description string, fields ...LogField) {
	helperLog(FTL, description, fields...)
}

func Error(description string, fields ...LogField) {
	helperLog(WRN, description, fields...)
}

func Success(description string, fields ...LogField) {
	helperLog(SUC, description, fields...)
}

func ShowTerminalCursor() {
	fmt.Fprint(os.Stdout, "\x1b[?25h")
}

func HideTerminalCursor() {
	fmt.Fprintf(os.Stdout, "\x1b[?25l")
}

var spinnerIterator int = 0

func CallSpinnerTitle(text string) {
	spinnerIterator++
	var shouldReturn bool = false

	finalChar := "\r"
	terminalW, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil && terminalW > 0 {
		nocolorText := strings.ReplaceAll(text, "\033[90m", "")
		nocolorText = strings.ReplaceAll(nocolorText, "\033[97m", "")
		nocolorText = strings.ReplaceAll(nocolorText, "\x1b[38;2;168;70;212m", "")
		if terminalW <= len(nocolorText)  {
			shouldReturn = true
		}
	} else {
		shouldReturn = true
	}

	if shouldReturn {
		finalChar = "\n"
	}

	if spinnerIterator%400 == 0 {
		shouldReturn = false
	}

	if err == nil && terminalW <= 80 {
		shouldReturn = true

		if spinnerIterator%4000 == 0 {
			shouldReturn = false
		}
	}

	if shouldReturn {
		return
	}

	loggerMutex.Lock()
	fmt.Printf(text + finalChar)
	loggerMutex.Unlock()
}

func PrintLogo(shouldClear bool) {
	if shouldClear {
		screen.Clear()
		screen.MoveTopLeft()
	}
	fmt.Printf(THEME+
		"\n               _        _\n"+
		"             _/|    _   |\\_\n"+
		"           _/_ |   _|\\\\ | _\\\n"+
		"         _/_/| /  /   \\|\\ |\\_\\_\n"+
		"       _/_/  |/  /  _  \\/\\|  \\_\\_ \n"+
		"     _/_/    ||  | | \\o/ ||    \\_\\_\n"+
		"    /_/  | | |\\  | \\_ V  /| | |  \\_\\    \033[97m  Tempo v2.0.0\n"+THEME+
		"   //    ||| | \\_/   \\__/ | |||    \\\\\033[97m     Enleashing The Power"+THEME+"\n"+
		"  // __| ||\\  \\          /  /|| |__ \\\\\n"+
		" //_/ \\|||| \\/\\\\        //\\/ ||||/ \\_\\\\\n"+
		"///    \\\\\\\\/   /        \\   \\////    \\\\\\\n"+
		"|/      \\/    |    |    |     \\/      \\|\n"+
		"            /_|  | |_  \\\n"+
		"           ///_| |_||\\_ \\\n"+
		"          |//||/||\\/||\\ |\n"+
		"           / \\/|||/||/\\/\n"+
		"             /|/\\| \\/\n"+
		"             \\/  |\n\n\n")
}
