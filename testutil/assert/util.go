package assert

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/reflects"
	"github.com/gookit/goutil/strutil"
)

// isEmpty value check
func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	return reflects.IsEmpty(reflect.ValueOf(v))
}

func checkEqualArgs(expected, actual any) error {
	if expected == nil && actual == nil {
		return nil
	}

	if reflects.IsFunc(expected) || reflects.IsFunc(actual) {
		return errors.New("cannot take func type as argument")
	}
	return nil
}

// formatUnequalValues takes two values of arbitrary types and returns string
// representations appropriate to be presented to the user.
//
// If the values are not of like type, the returned strings will be prefixed
// with the type name, and the value will be enclosed in parentheses similar
// to a type conversion in the Go grammar.
func formatUnequalValues(expected, actual any) (e string, a string) {
	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		return truncatingFormat(expected), truncatingFormat(actual)
		// return fmt.Sprintf("%T(%s)", expected, truncatingFormat(expected)),
		// 	fmt.Sprintf("%T(%s)", actual, truncatingFormat(actual))
	}

	switch expected.(type) {
	case time.Duration:
		return fmt.Sprintf("%v", expected), fmt.Sprintf("%v", actual)
	}

	return truncatingFormat(expected), truncatingFormat(actual)
}

// truncatingFormat formats the data and truncates it if it's too long.
//
// This helps keep formatted error messages lines from exceeding the
// bufio.MaxScanTokenSize max line length that the go testing framework imposes.
func truncatingFormat(data any) string {
	if data == nil {
		return "<nil>"
	}

	var value string
	switch data.(type) {
	case string:
		value = fmt.Sprintf("string(%q)", data)
	default:
		value = fmt.Sprintf("%T(%v)", data, data)
	}

	// Give us some space the type info too if needed.
	max := bufio.MaxScanTokenSize - 1000
	if len(value) > max {
		value = value[0:max] + "<... truncated>"
	}
	return value
}

func callerInfos() []string {
	num := 3
	skip := 2
	ss := make([]string, 0, num)

	for i := skip; i < skip+num; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			// The breaks below failed to terminate the loop, and we ran off the
			// end of the call stack.
			break
		}

		fc := runtime.FuncForPC(pc)
		if fc == nil {
			continue
		}

		// This is a huge edge case, but it will panic if this is the case
		if file == "<autogenerated>" {
			continue
		}

		fcName := fc.Name()
		if fcName == "testing.tRunner" || strings.Contains(fcName, "goutil/testutil/assert.") {
			continue
		}

		// eg: runtime.goexit
		if strings.HasPrefix(fcName, "runtime.") {
			continue
		}

		filePath := file
		if !ShowFullPath {
			filePath = filepath.Base(filePath)
		}

		ss = append(ss, fmt.Sprintf("%s:%d", filePath, line))
	}

	return ss
}

// refers from stretchr/testify/assert
type labeledText struct {
	label   string
	message string
}

func formatLabeledTexts(lts []labeledText) string {
	labelWidth := 0
	elemSize := len(lts)
	for _, lt := range lts {
		labelWidth = mathutil.MaxInt(len(lt.label), labelWidth)
	}

	var sb strings.Builder
	for i, lt := range lts {
		label := lt.label
		if EnableColor {
			label = color.Green.Sprint(label)
		}

		sb.WriteString("  " + label + strutil.Repeat(" ", labelWidth-len(lt.label)) + ":  ")
		formatMessage(lt.message, labelWidth, &sb)
		if i+1 != elemSize {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func formatMessage(message string, labelWidth int, buf comdef.StringWriteStringer) string {
	for i, scanner := 0, bufio.NewScanner(strings.NewReader(message)); scanner.Scan(); i++ {
		// skip add prefix for first line.
		if i != 0 {
			// +3: is len of ":  "
			_, _ = buf.WriteString("\n  " + strings.Repeat(" ", labelWidth+3))
		}
		_, _ = buf.WriteString(scanner.Text())
	}

	return buf.String()
}
