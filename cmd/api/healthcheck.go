package main

import (
	"net/http"

	"github.com/henrtytanoh/greenlight/render"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in the response.
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	buf, err := render.Template("healthcheck.tmpl", "healthcheck", env)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		return
	}
}
