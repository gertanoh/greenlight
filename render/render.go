package render

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed "templates"
var renderTemplateFS embed.FS

var pathToTemplates = "templates/"

func Template(templateFile string, templateName string, data interface{}) (*bytes.Buffer, error) {
	// Use the ParseFS() method to parse the required template file from the embedded
	// file system.
	tmpl, err := template.New("render").ParseFS(renderTemplateFS, pathToTemplates+templateFile)
	if err != nil {
		return nil, err
	}

	// Execute the named template "templateName", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
