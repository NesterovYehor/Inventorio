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

var componentCache = map[string]*template.Template{}

func InitUI() {
	base := template.Must(template.ParseFS(templatesFS, "templates/base.html", "templates/partials/sidebar.html", "templates/partials/row.html"))

	pageCache["base"] = base
	pageCache["properties"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/properties.html"))
	pageCache["draft_order"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/draft_order.html"))
	pageCache["confirmed_order"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/confirmed_order.html"))
	pageCache["storage"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/storage.html"))
	pageCache["orders"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/orders.html"))

	rows := template.Must(template.ParseFS(templatesFS, "templates/partials/row.html"))
	componentCache["item-row"] = rows
	componentCache["property-row"] = rows
	componentCache["calculator_row"] = rows

	// Single parse loads all definitions: "calculator_update", "property_item", and "calculator_rows"
	calcTmpl := template.Must(template.ParseFS(templatesFS, "templates/draft_order.html"))

	componentCache["calculator_update"] = calcTmpl
	componentCache["property_item"] = calcTmpl
	componentCache["calculator_rows"] = calcTmpl
	componentCache["calculator_tbody_oob"] = template.Must(calcTmpl.ParseFS(templatesFS, "templates/partials/row.html"))
}
func RenderContent(w http.ResponseWriter, r *http.Request, pageName string, data any) {
	tmpl, exist := pageCache[pageName]
	if !exist {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		tmpl.ExecuteTemplate(w, "content", data)
		return
	}

	tmpl.ExecuteTemplate(w, "base", data)
}

func RenderComponent(w http.ResponseWriter, componentName string, data any) {
	tmpl, exist := componentCache[componentName]
	if !exist {
		http.Error(w, "Component not found", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, componentName, data)
}

func StaticHandler() http.Handler {
	return http.FileServer(http.FS(staticFS))
}
