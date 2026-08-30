package ui

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

type Renderer struct {
	pageCache      map[string]*template.Template
	componentCache map[string]*template.Template
	bundle         *i18n.Bundle
}

func NewRenderer() *Renderer {
	return &Renderer{
		pageCache:      map[string]*template.Template{},
		componentCache: map[string]*template.Template{},
		bundle:         i18n.NewBundle(language.Russian),
	}
}

func (r *Renderer) InitUI() {
	base := template.Must(template.ParseFS(templatesFS, "templates/base.html", "templates/partials/sidebar.html", "templates/partials/row.html"))

	r.pageCache["base"] = base
	r.pageCache["properties"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/properties.html"))
	r.pageCache["draft_order"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/draft_order.html"))
	r.pageCache["confirmed_order"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/confirmed_order.html"))
	r.pageCache["storage"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/storage.html"))
	r.pageCache["orders"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/orders.html"))

	rows := template.Must(template.ParseFS(templatesFS, "templates/partials/row.html"))
	r.componentCache["item-row"] = rows
	r.componentCache["property-row"] = rows
	r.componentCache["calculator_row"] = rows

	// Single parse loads all definitions: "calculator_update", "property_item", and "calculator_rows"
	calcTmpl := template.Must(template.ParseFS(templatesFS, "templates/draft_order.html"))

	r.componentCache["calculator_update"] = calcTmpl
	r.componentCache["property_item"] = calcTmpl
	r.componentCache["calculator_rows"] = calcTmpl
	r.componentCache["calculator_tbody_oob"] = template.Must(calcTmpl.ParseFS(templatesFS, "templates/partials/row.html"))
}
func (r *Renderer) Content(w http.ResponseWriter, req *http.Request, pageName string, data any) {

	tmpl, exist := r.pageCache[pageName]
	if !exist {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if req.Header.Get("HX-Request") == "true" {
		tmpl.ExecuteTemplate(w, "content", r.buildTemplateData(req, data))
		return
	}

	tmpl.ExecuteTemplate(w, "base", data)
}

func (r *Renderer) Component(w http.ResponseWriter, req *http.Request, componentName string, data any) {
	tmpl, exist := r.componentCache[componentName]
	if !exist {
		http.Error(w, "Component not found", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, componentName, r.buildTemplateData(req, data))
}

func StaticHandler() http.Handler {
	return http.FileServer(http.FS(staticFS))
}

func (r *Renderer) buildTemplateData(req *http.Request, data any) map[string]any {
	lang, ok := req.Context().Value("lang").(string)
	if !ok {
		lang = "ru"
	}

	localizer := i18n.NewLocalizer(r.bundle, lang)

	translateFunc := func(messageID string) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID})
	}

	return map[string]any{
		"Data": data,
		"T":    translateFunc,
	}
}
