package view

import (
	"html/template"
	"net/http"
)

// Template for the blocked message served at root "/"
var tmpl = template.Must(template.New("blocked").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Blocked</title>
</head>
<body>
    <h1>This page has been blocked by Leo</h1>
    <p>You tried to access <strong>some domain</strong>, but it is blocked on this network.</p>
</body>
</html>
`))

// Handler that serves the blocked message on root "/"
func BlockedPageHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain") // Get the blocked domain from query parameters

	err := tmpl.Execute(w, struct{ Domain string }{Domain: domain})
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}
