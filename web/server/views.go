package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	gotl "github.com/panyam/goutils/template"
	oa "github.com/panyam/oneauth"
	tmplr "github.com/panyam/templar"
	svc "github.com/panyam/turnengine/games/weewar/services"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const TEMPLATES_FOLDER = "./web/templates"

// You may have a builder/bundler creating an output folder.  Set that path here.  It can be absolute or relative to
// where the executable will be running from
const DIST_FOLDER = "./web/dist"
const STATIC_FOLDER = "./web/static"

type ViewContext struct {
	AuthMiddleware *oa.Middleware
	ClientMgr      *svc.ClientMgr
	Ctx            context.Context
	Templates      *tmplr.TemplateGroup
}

type ViewMaker func() View

type Copyable interface {
	Copy() View
}

func Copier[V Copyable](v V) ViewMaker {
	return v.Copy
}

type View interface {
	Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool)
}

type RootViewsHandler struct {
	mux     *http.ServeMux
	Context *ViewContext
}

func NewRootViewsHandler(middleware *oa.Middleware, clients *svc.ClientMgr) *RootViewsHandler {
	out := RootViewsHandler{
		mux: http.NewServeMux(),
	}

	templates := tmplr.NewTemplateGroup()
	templates.Loader = (&tmplr.LoaderList{}).AddLoader(tmplr.NewFileSystemLoader(TEMPLATES_FOLDER))
	templates.AddFuncs(gotl.DefaultFuncMap())
	templates.AddFuncs(template.FuncMap{
		"Ctx": func() *ViewContext {
			return out.Context
		},
		"UserInfo": func(userId string) map[string]any {
			// Just a hacky cache
			return map[string]any{
				"FullName":  "XXXX YYY",
				"Name":      "XXXX",
				"AvatarUrl": "/avatar/url",
			}
		},
		"AsHtmlAttribs": func(m map[string]string) template.HTML {
			return `a = 'b' c = 'd'`
		},
		"Indented": func(nspaces int, code string) (formatted string) {
			lines := (strings.Split(strings.TrimSpace(code), "\n"))
			return strings.Join(lines, "<br/>")
		},
		"dset": func(d map[string]any, key string, value any) map[string]any {
			d[key] = value
			return d
		},
		"lset": func(a []any, index int, value any) []any {
			a[index] = value
			return a
		},
		"safeHTMLAttr": func(s string) template.HTMLAttr {
			return template.HTMLAttr(s)
		},
		"ToJson": func(v interface{}) template.JS {
			if v == nil {
				return template.JS("null")
			}
			// Use protojson.Marshal for protobuf types, regular json.Marshal for others
			// Check if it's a protobuf message using proto.Message interface
			if msg, ok := v.(proto.Message); ok {
				jsonBytes, err := protojson.Marshal(msg)
				if err == nil {
					return template.JS(jsonBytes)
				}
				log.Printf("Error marshaling protobuf to JSON: %v", err)
			}
			// Fall back to regular JSON marshaling for non-protobuf types
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				log.Printf("Error marshaling to JSON: %v", err)
				return template.JS("null")
			}
			return template.JS(jsonBytes)
		},
	})
	out.Context = &ViewContext{
		AuthMiddleware: middleware,
		ClientMgr:      clients,
		Templates:      templates,
	}

	// setup routes
	out.setupRoutes()
	return &out
}

func (b *RootViewsHandler) ViewRenderer(view ViewMaker, template string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: ViewRenderer called for path %s", r.URL.Path)
		b.RenderView(view(), template, r, w)
	}
}

func (b *RootViewsHandler) RenderView(view View, template string, r *http.Request, w http.ResponseWriter) {
	if template == "" {
		t := reflect.TypeOf(view)
		e := t.Elem()
		template = e.Name()
	}
	log.Printf("DEBUG: Rendering template '%s' for view type %T", template, view)
	err, finished := view.Load(r, w, b.Context)
	if !finished {
		if err != nil {
			log.Println("Error: ", err)
			fmt.Fprint(w, "Error rendering: ", err.Error())
		} else {
			templateFile := template + ".html"
			log.Printf("DEBUG: Loading template file '%s'", templateFile)
			tmpl, err := b.Context.Templates.Loader.Load(templateFile, "")
			if err != nil {
				log.Println("Template Load Error: ", templateFile, err)
				fmt.Fprint(w, "Error rendering: ", err.Error())
			} else {
				log.Printf("DEBUG: Successfully loaded template, rendering...")
				err = b.Context.Templates.RenderHtmlTemplate(w, tmpl[0], template, view, nil)
				if err != nil {
					log.Printf("DEBUG: Template render error: %v", err)
					fmt.Fprint(w, "Template render error: ", err.Error())
				} else {
					log.Printf("DEBUG: Template rendered successfully")
				}
			}
		}
	}
}

func (b *RootViewsHandler) HandleError(err error, w io.Writer) {
	if err != nil {
		fmt.Fprint(w, "Error rendering: ", err.Error())
	}
}

func (n *RootViewsHandler) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("DEBUG: RootViewsHandler received request: %s %s", r.Method, r.URL.Path)
		n.mux.ServeHTTP(w, r)
	})
}

// Here you can setup all your view routes, pages, etc
func (n *RootViewsHandler) setupRoutes() {
	log.Println("DEBUG: Setting up routes...")
	// This is the chance to setup all your routes for your app across various resources etc
	// Typically "/views" is dedicated for returning view fragments - eg via htmx
	n.mux.Handle("/views/", http.StripPrefix("/views", n.setupViewsMux()))

	n.mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(STATIC_FOLDER))))

	// Then setup your "resource" specific endpoints
	n.mux.Handle("/games/", http.StripPrefix("/games", n.setupGamesMux()))
	n.mux.Handle("/worlds/", http.StripPrefix("/worlds", n.setupWorldsMux()))
	
	// Handle no-trailing-slash redirects for convenience
	n.mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/games/", http.StatusMovedPermanently)
	})
	n.mux.HandleFunc("/worlds", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/worlds/", http.StatusMovedPermanently)
	})

	n.mux.HandleFunc("/about", n.ViewRenderer(Copier(&GenericPage{}), "AboutPage"))
	n.mux.HandleFunc("/contact", n.ViewRenderer(Copier(&GenericPage{}), "ContactUsPage"))
	n.mux.HandleFunc("/login", n.ViewRenderer(Copier(&LoginPage{}), ""))
	// n.mux.HandleFunc("/logout", n.onLogout)
	n.mux.HandleFunc("/privacy-policy", n.ViewRenderer(Copier(&PrivacyPolicy{}), ""))
	n.mux.HandleFunc("/terms-of-service", n.ViewRenderer(Copier(&TermsOfService{}), ""))
	log.Println("DEBUG: Registering root path handler")
	n.mux.HandleFunc("/", n.ViewRenderer(Copier(&HomePage{}), ""))
	n.mux.Handle("/{invalidbits}/", http.NotFoundHandler()) // <-- Default 404

	// Alternatively if you have things getting built in a dist folder we could do:
	/*
		r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/" {
				http.Redirect(w, req, "/appitems", http.StatusFound)
				return
			}
			// Serve static files for other root-level paths
			http.FileServer(http.Dir(DIST_FOLDER)).ServeHTTP(w, req)
		})
	*/
}

func (n *RootViewsHandler) setupViewsMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Setup the various views you want to return here
	/*, eg:
	mux.HandleFunc("/ComposerSelectionModal", n.ViewRenderer(Copier(&ComposerSelectionModal{}), ""))
	mux.HandleFunc("/ListTemplatesView", n.ViewRenderer(Copier(&ListTemplatesView{}), ""))
	mux.HandleFunc("/notations/ListView", n.ViewRenderer(Copier(&NotationListView{}), ""))
	*/

	// n.HandleView(Copier(&components.SelectTemplatePage{}), r, w)
	return mux
}
