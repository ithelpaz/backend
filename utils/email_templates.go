package utils

import (
	"bytes"
	"html/template"
 
	"path/filepath"
)

func RenderEmailTemplate(filename string, data any) (string, error) {
	basePath := "templates/email"
	path := filepath.Join(basePath, filename)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}