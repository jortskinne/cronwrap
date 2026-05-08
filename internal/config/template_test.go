package config

import (
	"testing"
)

func TestApplyTemplateDefaults_EmptyTemplate(t *testing.T) {
	c := &TemplateConfig{}
	ApplyTemplateDefaults(c)
	if c.Template != "" {
		t.Errorf("expected empty template to remain empty, got %q", c.Template)
	}
}

func TestApplyTemplateDefaults_PreservesCustomTemplate(t *testing.T) {
	custom := "Job {{.Command}} finished with {{.ExitCode}}"
	c := &TemplateConfig{Template: custom}
	ApplyTemplateDefaults(c)
	if c.Template != custom {
		t.Errorf("expected template to be preserved, got %q", c.Template)
	}
}

func TestTemplateConfig_EmptyMeansDefault(t *testing.T) {
	c := TemplateConfig{}
	if c.Template != "" {
		t.Errorf("zero value should be empty string")
	}
}

func TestTemplateConfig_CustomValue(t *testing.T) {
	c := TemplateConfig{Template: "{{.Status}}"}
	if c.Template != "{{.Status}}" {
		t.Errorf("unexpected template value: %q", c.Template)
	}
}
