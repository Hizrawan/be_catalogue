package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"be20250107/internal/app"
	"be20250107/internal/errors"
	"be20250107/internal/middlewares"
	"be20250107/internal/reqdata"
)

type Controller struct {
	App *app.Registry
}

func (c *Controller) Parse(reqPtr reqdata.Form, r *http.Request) error {
	if r.ContentLength == 0 {
		return nil
	}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		// Todo: Handle multipart form data
	} else {
		err := json.NewDecoder(r.Body).Decode(reqPtr)
		if err != nil {
			return errors.ErrMalformedRequest
		}
	}

	return nil 
}

func (c *Controller) Validate(reqPtr reqdata.Form, r *http.Request) error {
	err := c.Parse(reqPtr, r)
	if err != nil {
		return err
	}

	ctx := c.RequestContext(r)
	if ok := reqPtr.Authorized(ctx); !ok {
		panic(errors.ErrForbidden)
	}
	if err = reqPtr.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Controller) RequestContext(r *http.Request) *reqdata.Context {
	var auth reqdata.AuthInformation
	if a, ok := r.Context().Value(middlewares.ContextAuth).(reqdata.AuthInformation); ok {
		auth = a
	}

	return &reqdata.Context{
		App:         c.App,
		Auth:        auth,
		HTTPContext: r.Context(),
	}
}

func (c *Controller) AssertAuthenticated(r *http.Request) reqdata.AuthInformation {
	ctx := c.RequestContext(r)
	if !ctx.Auth.IsLoggedIn() {
		panic(errors.ErrUnauthenticated)
	}
	return ctx.Auth
}

// Forbidden will send a panic of type errors.ErrForbidden. Can be used to stop
// controllers execution in case of lack of access privilege. When the controllers
// is running in the default HTTP server, this error will be automatically
// converted to the proper HTTP 403 response.
func (c *Controller) Forbidden() {
	panic(errors.ErrForbidden)
}
