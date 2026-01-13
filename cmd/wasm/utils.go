package main

import (
	"bytes"
	"context"

	tmpls "github.com/turnforge/lilbattle/web/templates"
)

func renderPanelTemplate(_ context.Context, templatefile string, data any) (content string) {
	tmpl, err := tmpls.Templates.Loader.Load(templatefile, "")
	if err == nil {
		buf := bytes.NewBufferString("")
		err = tmpls.Templates.RenderHtmlTemplate(buf, tmpl[0], "", data, nil)
		if err == nil {
			content = buf.String()
		}
	}
	if err != nil {
		panic(err)
	}
	return
}
