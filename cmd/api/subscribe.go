package main

import (
	"fmt"
	"net/http"
)

func (app *application) createSubscription(w http.ResponseWriter, r *http.Request) {
	app.logger.PrintInfo("Inside create", nil)
	user := app.contextGetUser(r)

	app.logger.PrintInfo(fmt.Sprintf("%d", user.ID), nil)
	err := app.models.Permissions.AddForUser(user.ID, "movies:write")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteSubscription(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	err := app.models.Permissions.RemoveForUser(user.ID, "movies:write")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
