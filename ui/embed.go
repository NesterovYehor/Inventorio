package ui

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

var pageCache = map[string]*template.Template{}

func InitUI() {
	base := template.Must(template.ParseFS(templatesFS, "templates/base.html", "templates/partials/sidebar.html"))

	pageCache["properties"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/properties.html"))
}

func Render(w http.ResponseWriter, r *http.Request, pageName string, data any) {
	tmpl, exist := pageCache[pageName]
	if !exist {
		http.Error(w, "Template not found", http.StatusInternalServerError)
	}

	if r.Header.Get("HX-Request") == "true" {
		tmpl.ExecuteTemplate(w, "content", data)
		return
	}

	tmpl.ExecuteTemplate(w, "base", data)

}

func StaticHandler() http.Handler{
	return http.FileServer(http.FS(staticFS))
}
