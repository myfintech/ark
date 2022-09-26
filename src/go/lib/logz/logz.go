package logz

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/pkg/errors"

	"github.com/muesli/termenv"
)

// Fields type, used to pass to `WithFields`.
type Fields = map[string]interface{}

var ctpl *template.Template
var logLineTemplate = `{{ $fieldKeyColor := .FieldKeyColor -}}
{{ $fieldValueColor := .FieldValueColor -}}
{{ $level := Color .LevelBackgroundColor .Level -}}
{{ $timestamp := Color .TimestampBackgroundColor .Timestamp -}}
{{ $level }} {{ $timestamp }}{{ print " " -}}
{{ range $key, $value := .Fields }}{{ Color $fieldKeyColor $key }}={{ Color $fieldValueColor (printf "%s" $value) -}}{{print " "}}{{ end -}} 
{{ Color .Message }}
`

type templateData struct {
	LevelBackgroundColor     string
	Level                    string
	TimestampBackgroundColor string
	Timestamp                string
	Message                  string
	Fields                   Fields
	FieldKeyColor            string
	FieldValueColor          string
}

func init() {
	tmpl, err := template.New("tpl").
		Funcs(termenv.TemplateFuncs(termenv.ColorProfile())).
		Parse(logLineTemplate)

	if err != nil {
		panic(err)
	}
	ctpl = tmpl
}

func LevelColor(level Level) string {
	switch level {
	case TraceLevel:
		return "#3a3a3a"
	case DebugLevel:
		return "#3a3a3a"
	case InfoLevel:
		return "#008080"
	case WarnLevel:
		return "#fff000"
	case ErrorLevel:
		return "#ff0000"
	default:
		return "#3a3a3a"
	}
}

// DefaultFormatter a simple Entry format function
func DefaultFormatter(entry Entry) ([]byte, error) {
	buff := new(bytes.Buffer)

	if entry.Caller != "" {
		entry.Fields["caller"] = entry.Caller
	}

	err := ctpl.Execute(buff, templateData{
		LevelBackgroundColor:     LevelColor(entry.Level),
		Level:                    entry.Level.String(),
		TimestampBackgroundColor: "#3F6DAA",
		Timestamp:                entry.Timestamp.Format(time.RFC3339),

		Fields:          entry.Fields,
		FieldValueColor: "#444444",
		FieldKeyColor:   "#00875f",

		Message: entry.Message,
	})

	if err != nil {
		panic(err)
	}

	return buff.Bytes(), nil
}

var newLine = "\n"

// JsonFormatter formats log line as json
func JsonFormatter(entry Entry) ([]byte, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return append(data, newLine...), nil
}

// SuggestedFilePath returns the suggested path for a log file based on the host operating system
func SuggestedFilePath(tool, filename string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "linux":
		return filepath.Join("/var/log", tool, filename), nil
	case "darwin":
		return filepath.Join(home, "Library/Logs", tool, filename), nil
	default:
		return "", errors.Errorf("%s is not presently a supported OS; please let us know if you'd like to use our tool on your OS!", runtime.GOOS)
	}
}
