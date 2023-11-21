package render

import (
	"bytes"
	"embed"
	"fmt"
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

	// fmt.Println(buf.String())
	_, err = buf.WriteTo(w)
	if err != nil {
		return nil, err
	}
	fmt.Println(buf.String())

	return buf, nil
}
