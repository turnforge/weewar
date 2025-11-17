package server

import (
	"log"
	"net/http"

	"github.com/panyam/templar"
)

func ThemeFromRequest(req *http.Request) string {
	// Theme fallback priority: URL param > cookie > default
	queryParams := req.URL.Query()
	theme := queryParams.Get("theme")
	if theme == "" {
		if cookie, err := req.Cookie("assetTheme"); err == nil {
			theme = cookie.Value
		}
	}
	if theme == "" {
		theme = "fantasy" // Default theme
	}
	return theme
}

// SetupTemplates initializes the Templar template group
func SetupTemplates(templatesDir string) (*templar.TemplateGroup, error) {
	// Create a new template group
	group := templar.NewTemplateGroup()

	// Set up the file appitem loader with multiple paths
	group.Loader = templar.NewFileSystemLoader(
		templatesDir,
		templatesDir+"/shared",
		templatesDir+"/components",
	)

	// Preload common templates to ensure they're available
	commonTemplates := []string{
		"base.html",
		"appitems/listing.html",
		"appitems/details.html",
	}

	for _, tmpl := range commonTemplates {
		// Use defer to catch panics from MustLoad
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Template not found (will create): %s", tmpl)
				}
			}()
			group.MustLoad(tmpl, "")
		}()
	}

	return group, nil
}
