package common

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

const (
	defaultTimestampFormat = time.RFC3339
)

var (
	baseTimestamp      time.Time
	levelTextMaxLength int
)

func init() {
	baseTimestamp = time.Now()

	for _, level := range logrus.AllLevels {
		levelTextLength := utf8.RuneCount([]byte(level.String()))
		if levelTextLength > levelTextMaxLength {
			levelTextMaxLength = levelTextLength
		}
	}
}

type fieldKey string

// FieldMap allows customization of the key names for default fields.
type FieldMap map[fieldKey]string

func (f FieldMap) resolve(key fieldKey) string {
	if k, ok := f[key]; ok {
		return k
	}

	return string(key)
}

// UncoloredTextFormatter formats logs into text
type UncoloredTextFormatter struct {
	// Force quoting of all values
	ForceQuote bool

	// DisableQuote disables quoting for all values.
	// DisableQuote will have a lower priority than ForceQuote.
	// If both of them are set to true, quote will be forced on all values.
	DisableQuote bool

	// Override coloring based on CLICOLOR and CLICOLOR_FORCE. - https://bixense.com/clicolors/
	EnvironmentOverrideColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool

	// TimestampFormat to use for display when a full timestamp is printed.
	// The format to use is the same than for time.Format or time.Parse from the standard
	// library.
	// The standard Library already provides a set of predefined format.
	TimestampFormat string

	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool

	// The keys sorting function, when uninitialized it uses sort.Strings.
	SortingFunc func([]string)

	// Disables the truncation of the level text to 4 characters.
	DisableLevelTruncation bool

	// PadLevelText Adds padding the level text so that all the levels output at the same length
	// PadLevelText is a superset of the DisableLevelTruncation option
	PadLevelText bool

	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool

	// FieldMap allows users to customize the names of keys for default fields.
	// As an example:
	// formatter := &UncoloredTextFormatter{
	//     FieldMap: FieldMap{
	//         FieldKeyTime:  "@timestamp",
	//         FieldKeyLevel: "@level",
	//         FieldKeyMsg:   "@message"}}
	FieldMap FieldMap

	// CallerPrettyfier can be set by the user to modify the content
	// of the function and file keys in the data when ReportCaller is
	// activated. If any of the returned value is the empty string the
	// corresponding key will be removed from fields.
	CallerPrettyfier func(*runtime.Frame) (function string, file string)
}

// Format renders a single log entry
func (f *UncoloredTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	prefixFieldClashes(data, f.FieldMap, entry.HasCaller())
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	if !f.DisableSorting {
		if f.SortingFunc == nil {
			sort.Strings(keys)
		}
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	f.printToBuf(b, entry, keys, data, timestampFormat)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *UncoloredTextFormatter) printToBuf(b *bytes.Buffer, entry *logrus.Entry, keys []string, data logrus.Fields, timestampFormat string) {
	levelText := strings.ToUpper(entry.Level.String())
	if !f.DisableLevelTruncation && !f.PadLevelText {
		levelText = levelText[0:4]
	}
	if f.PadLevelText {
		// Generates the format string used in the next line, for example "%-6s" or "%-7s".
		// Based on the max level text length.
		formatString := "%-" + strconv.Itoa(levelTextMaxLength) + "s"
		// Formats the level text by appending spaces up to the max length, for example:
		// 	- "INFO   "
		//	- "WARNING"
		levelText = fmt.Sprintf(formatString, levelText)
	}

	// Remove a single newline if it already exists in the message to keep
	// the behavior of logrus text_formatter the same as the stdlib log package
	entry.Message = strings.TrimSuffix(entry.Message, "\n")

	caller := ""
	if entry.HasCaller() {
		funcVal := fmt.Sprintf("%s()", entry.Caller.Function)
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)

		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		}

		if fileVal == "" {
			caller = funcVal
		} else if funcVal == "" {
			caller = fileVal
		} else {
			caller = fileVal + " " + funcVal
		}
	}

	switch {
	case f.DisableTimestamp:
		fmt.Fprintf(b, "%s%s %-44s ", levelText, caller, entry.Message)
	case !f.FullTimestamp:
		fmt.Fprintf(b, "%s[%04d]%s %-44s ", levelText, int(entry.Time.Sub(baseTimestamp)/time.Second), caller, entry.Message)
	default:
		fmt.Fprintf(b, "%s[%s]%s %-44s ", levelText, entry.Time.Format(timestampFormat), caller, entry.Message)
	}
	for _, k := range keys {
		v := data[k]
		fmt.Fprintf(b, " %s=", k)
		f.appendValue(b, v)
	}
}

func (f *UncoloredTextFormatter) needsQuoting(text string) bool {
	if f.ForceQuote {
		return true
	}
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	if f.DisableQuote {
		return false
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func (f *UncoloredTextFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

func prefixFieldClashes(data logrus.Fields, fieldMap FieldMap, reportCaller bool) {
	timeKey := fieldMap.resolve(logrus.FieldKeyTime)
	if t, ok := data[timeKey]; ok {
		data["fields."+timeKey] = t
		delete(data, timeKey)
	}

	msgKey := fieldMap.resolve(logrus.FieldKeyMsg)
	if m, ok := data[msgKey]; ok {
		data["fields."+msgKey] = m
		delete(data, msgKey)
	}

	levelKey := fieldMap.resolve(logrus.FieldKeyLevel)
	if l, ok := data[levelKey]; ok {
		data["fields."+levelKey] = l
		delete(data, levelKey)
	}

	logrusErrKey := fieldMap.resolve(logrus.FieldKeyLogrusError)
	if l, ok := data[logrusErrKey]; ok {
		data["fields."+logrusErrKey] = l
		delete(data, logrusErrKey)
	}

	// If reportCaller is not set, 'func' will not conflict.
	if reportCaller {
		funcKey := fieldMap.resolve(logrus.FieldKeyFunc)
		if l, ok := data[funcKey]; ok {
			data["fields."+funcKey] = l
		}
		fileKey := fieldMap.resolve(logrus.FieldKeyFile)
		if l, ok := data[fileKey]; ok {
			data["fields."+fileKey] = l
		}
	}
}
