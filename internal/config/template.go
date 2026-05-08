package config

// TemplateConfig holds configuration for custom notification message templates.
type TemplateConfig struct {
	// Template is a Go text/template string used to render notification messages.
	// If empty, the default template is used.
	Template string `toml:"template" yaml:"template"`
}

// ApplyTemplateDefaults sets sensible defaults for TemplateConfig.
// Currently a no-op since an empty Template triggers the built-in default,
// but kept for consistency with other Apply* functions.
func ApplyTemplateDefaults(c *TemplateConfig) {
	// No defaults to apply; empty template uses the built-in default.
}
