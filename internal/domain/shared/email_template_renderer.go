package shared

// EmailTemplateRenderer renders subject/body from template id and vars.
type EmailTemplateRenderer interface {
	Render(templateID string, vars map[string]string) (subject string, body string, err error)
}

