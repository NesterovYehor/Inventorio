package ui

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed locales
var localesFS embed.FS

type Renderer struct {
	pageCache      map[string]*template.Template
	componentCache map[string]*template.Template
	bundle         *i18n.Bundle
}

func NewRenderer() *Renderer {
	r := &Renderer{
		pageCache:      map[string]*template.Template{},
		componentCache: map[string]*template.Template{},
		bundle:         i18n.NewBundle(language.Russian),
	}
	r.loadLocales()
	return r
}

func (r *Renderer) loadLocales() {
	files := []string{"locales/en.json", "locales/ru.json", "locales/es.json"}
	for _, path := range files {
		if _, err := r.bundle.LoadMessageFileFS(localesFS, path); err != nil {
			log.Printf("failed to load locale %s: %v", path, err)
		}
	}
}

func (r *Renderer) InitUI() {
	base := template.Must(template.ParseFS(templatesFS, "templates/base.html", "templates/partials/sidebar.html", "templates/partials/row.html"))

	r.pageCache["base"] = base
	r.pageCache["properties"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/properties.html"))
	r.pageCache["draft_order"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/draft_order.html"))
	r.pageCache["confirmed_order"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/confirmed_order.html"))
	r.pageCache["storage"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/storage.html"))
	r.pageCache["arrivals"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/arrivals.html"))
	r.pageCache["orders"] = template.Must(template.Must(base.Clone()).ParseFS(templatesFS, "templates/orders.html"))

	rows := template.Must(template.ParseFS(templatesFS, "templates/partials/row.html"))
	r.componentCache["item-row"] = rows
	r.componentCache["property-row"] = rows
	r.componentCache["calculator_row"] = rows

	// Single parse loads all definitions: "calculator_update" and "calculator_rows"
	calcTmpl := template.Must(template.ParseFS(templatesFS, "templates/draft_order.html"))

	r.componentCache["calculator_update"] = calcTmpl
	r.componentCache["calculator_rows"] = calcTmpl
	r.componentCache["calculator_tbody_oob"] = template.Must(calcTmpl.ParseFS(templatesFS, "templates/partials/row.html"))

	arrivalsTmpl := template.Must(template.ParseFS(templatesFS, "templates/arrivals.html"))
	r.componentCache["arrival_row"] = arrivalsTmpl
	r.componentCache["arrival_property_select"] = arrivalsTmpl
	r.componentCache["arrival_modal"] = arrivalsTmpl
}
func (r *Renderer) Content(w http.ResponseWriter, req *http.Request, pageName string, data any) {

	tmpl, exist := r.pageCache[pageName]
	if !exist {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	tmplData := r.buildTemplateData(req, data)
	if req.Header.Get("HX-Request") == "true" {
		if err := tmpl.ExecuteTemplate(w, "content", tmplData); err != nil {
			log.Printf("failed to render content %s: %v", pageName, err)
		}
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", tmplData); err != nil {
		log.Printf("failed to render base %s: %v", pageName, err)
	}
}

func (r *Renderer) Component(w http.ResponseWriter, req *http.Request, componentName string, data any) {
	tmpl, exist := r.componentCache[componentName]
	if !exist {
		http.Error(w, "Component not found", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, componentName, r.buildTemplateData(req, data)); err != nil {
		log.Printf("failed to render component %s: %v", componentName, err)
	}
}

func StaticHandler() http.Handler {
	return http.FileServer(http.FS(staticFS))
}

type templateData struct {
	Data      any
	Lang      string
	localizer *i18n.Localizer
}

func (td *templateData) T(messageID string) string {
	msg, err := td.localizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		return messageID
	}
	return msg
}

func (td *templateData) Wrap(data any) *templateData {
	return &templateData{
		Data:      data,
		Lang:      td.Lang,
		localizer: td.localizer,
	}
}

func (r *Renderer) buildTemplateData(req *http.Request, data any) *templateData {
	lang := "ru"
	if req != nil {
		if l, ok := req.Context().Value("lang").(string); ok {
			lang = l
		}
	}

	return &templateData{
		Data:      data,
		Lang:      lang,
		localizer: i18n.NewLocalizer(r.bundle, lang),
	}
}
