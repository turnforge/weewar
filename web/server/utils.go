package server

import (
	"net/http"
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
