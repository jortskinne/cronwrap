package notifier

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/exampleorg/cronwrap/internal/runner"
)

// DefaultTemplate is the default notification message template.
const DefaultTemplate = `Job: {{.Command}}
Status: {{.Status}}
Exit Code: {{.ExitCode}}
Duration: {{.Duration}}
{{- if .Output}}
Output:
{{.Output}}
{{- end}}
{{- if .Error}}
Error: {{.Error}}
{{- end}}`

// TemplateData holds the values available to notification templates.
type TemplateData struct {
	Command  string
	Status   string
	ExitCode int
	Duration string
	Output   string
	Error    string
	Time     time.Time
}

// TemplateRenderer renders notification messages from a text/template string.
type TemplateRenderer struct {
	tmpl *template.Template
}

// NewTemplateRenderer parses the provided template string and returns a renderer.
// If tmplStr is empty, the DefaultTemplate is used.
func NewTemplateRenderer(tmplStr string) (*TemplateRenderer, error) {
	if tmplStr == "" {
		tmplStr = DefaultTemplate
	}
	t, err := template.New("notification").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}
	return &TemplateRenderer{tmpl: t}, nil
}

// Render produces a notification message string from a runner.Result.
func (r *TemplateRenderer) Render(result runner.Result) (string, error) {
	status := "success"
	if !result.Success {
		status = "failure"
	}
	data := TemplateData{
		Command:  result.Command,
		Status:   status,
		ExitCode: result.ExitCode,
		Duration: result.Duration.Round(time.Millisecond).String(),
		Output:   result.Output,
		Error:    result.Err,
		Time:     result.StartedAt,
	}
	var buf bytes.Buffer
	if err := r.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template render error: %w", err)
	}
	return buf.String(), nil
}
