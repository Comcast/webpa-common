package logging

import (
	"bytes"
	"io"
	"github.com/go-kit/kit/log"
	"time"
	"fmt"
	"strings"
	"errors"
	"github.com/ttacon/chalk"
	"github.com/spf13/cast"
	"sort"
)

type TextFormatter struct {
	// Force disabling colors.
	DisableColors bool `json:"disableColors"`

	// Disables the truncation of the level text to 4 characters.
	DisableLevelTruncation bool `json:"disableLevelTruncation"`

	// Disables the sorting of key value pairs
	DisableSorting bool `json:"disableSorting"`

	// Whether the logger's out is to a terminal
	isTerminal bool
}

type reformatLogger struct {
	baseTimestamp time.Time
	w             io.Writer
	formatter     *TextFormatter
}

// NewLogfmtLogger returns a logger that encodes keyvals to the Writer in
// logfmt format. Each log event produces no more than one call to w.Write.
// The passed Writer must be safe for concurrent use by multiple goroutines if
// the returned Logger will be used concurrently.
func NewReformatLogger(w io.Writer, formatter *TextFormatter) log.Logger {
	return log.NewSyncLogger(&reformatLogger{time.Now(), w, formatter})
}

func colorLevel(data reformatData) chalk.Color {

	switch data.level {
	case "debug":
		return chalk.Cyan
	case "info":
		return chalk.Blue
	case "warn":
		return chalk.Yellow
	case "error":
		return chalk.Red
	default:
		return chalk.ResetColor
	}
}

func (t *TextFormatter) isColored() bool {
	return !t.DisableColors
}

func writeColor(buf *bytes.Buffer, color chalk.Color, formatter *TextFormatter, coloredText string) {
	if formatter.isColored() {
		fmt.Fprint(buf, color.Color(coloredText))
	} else {
		fmt.Fprint(buf, coloredText)
	}
}

func writeStyle(buf *bytes.Buffer, style chalk.Style, formatter *TextFormatter, styledText string) {
	if formatter.isColored() {
		fmt.Fprint(buf, style.Style(styledText))
	} else {
		fmt.Fprint(buf, styledText)
	}
}

func (l *reformatLogger) Log(keyvals ...interface{}) error {
	buf := &bytes.Buffer{}
	data := mapKeyVals(keyvals)
	color := colorLevel(data)

	levelText := strings.ToUpper(data.level)
	if !l.formatter.DisableLevelTruncation {
		levelText = levelText[0:4]
	}

	// Write Level aka INFO, WARN
	writeColor(buf, color, l.formatter, levelText)
	fmt.Fprintf(buf, "[%05d] %-44s", int(data.time.Sub(l.baseTimestamp)/time.Second), data.msg)

	// if err, wright it
	if nil != data.error {
		writeStyle(buf, chalk.Black.NewStyle().WithBackground(chalk.Red), l.formatter, "ERR")
		fmt.Fprintf(buf, ":%s ", data.error.Error())
	}

	keys := make([]string, 0, len(data.fieldMap))
	for k := range data.fieldMap {
		keys = append(keys, k)
	}

	if !l.formatter.DisableSorting {
		sort.Strings(keys)
	}

	// Write KeyValue's
	for _, key := range keys {
		writeColor(buf, color, l.formatter, key)
		if printString, err := cast.ToStringE(data.fieldMap[key]); err == nil {
			fmt.Fprintf(buf, "=%s ", printString)
		} else {
			fmt.Fprintf(buf, "=%#v ", data.fieldMap[key])
		}
	}

	// Add newline to the end of the buffer
	//buf.WriteString("\n")
	fmt.Fprintf(buf, "\n")

	// The Logger interface requires implementations to be safe for concurrent
	// use by multiple goroutines. For this implementation that means making
	// only one call to l.w.Write() for each call to Log.
	if _, err := l.w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

type reformatData struct {
	time     time.Time
	msg      string
	level    string
	error    error
	fieldMap map[string]interface{}
}

func mapKeyVals(keyvals []interface{}) reformatData {
	data := reformatData{}
	data.fieldMap = make(map[string]interface{})
	for i := 0; i < len(keyvals)-1; i += 2 {
		if cast.ToString(keyvals[i]) == "level" {
			data.level = cast.ToString(keyvals[i+1])
			continue
		}
		if cast.ToString(keyvals[i]) == "ts" {
			if t, ok := keyvals[i+1].(time.Time); ok {
				data.time = t
			} else if t, err := time.Parse(time.RFC3339, cast.ToString(keyvals[i+1])); err == nil {
				data.time = t
			} else {
				// We failed to parse it this should be close enough
				data.time = time.Now()
			}
			continue
		}
		if cast.ToString(keyvals[i]) == "msg" {
			data.msg = cast.ToString(keyvals[i+1])
			continue
		}
		if cast.ToString(keyvals[i]) == "err" || cast.ToString(keyvals[i]) == "error" {
			if err, ok := keyvals[i+1].(error); ok {
				data.error = err
				continue
			} else if errString := cast.ToString(keyvals[i+1]); len(strings.TrimSpace(errString)) > 0 {
				data.error = errors.New(errString)
				continue
			}

		}
		data.fieldMap[cast.ToString(keyvals[i])] = keyvals[i+1]

	}

	// Set Default values
	if data.level == "" {
		data.level = "info"
	}
	if data.time.IsZero() {
		data.time = time.Now()
	}

	return data
}
