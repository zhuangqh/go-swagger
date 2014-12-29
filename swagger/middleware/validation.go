package middleware

import (
	"net/http"

	"github.com/casualjim/go-swagger/swagger/errors"
	"github.com/casualjim/go-swagger/swagger/httputils"
	"github.com/casualjim/go-swagger/swagger/router"
	"github.com/casualjim/go-swagger/swagger/validate"
)

// NewValidation starts a new validation middleware
func newValidation(context *Context) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {

	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		matched, _ := context.RouteInfo(r)

		result := validateRequest(context, r, matched)
		if result.HasErrors() {
			context.Respond(rw, r, result.Errors[0])
			return
		}

		next(rw, r)
	}
}

type validation struct {
	context *Context
	result  *result
	request *http.Request
	route   *router.MatchedRoute
	bound   map[string]interface{}
}

func validateRequest(context *Context, request *http.Request, route *router.MatchedRoute) *result {
	validate := &validation{context, &result{}, request, route, make(map[string]interface{})}

	validate.contentType()
	validate.responseFormat()
	validate.parameters()

	return validate.result
}

func (v *validation) parameters() {
	for _, param := range v.route.Parameters {
		for _, err := range validate.Parameter(v.request, v.route, v.bound, &param) {
			v.result.AddErrors(err)
		}
	}
}

func (v *validation) contentType() {
	if httputils.CanHaveBody(v.request.Method) {
		ct, _, err := v.context.ContentType(v.request)
		if err != nil {
			v.result.AddErrors(err)
		} else {
			if err := validate.ContentType(v.route.Consumes, ct); err != nil {
				v.result.AddErrors(err)
			}
		}
	}
}

func (v *validation) responseFormat() {
	if str := v.context.ResponseFormat(v.request, v.route.Produces); str == "" {
		// v.result.AddErrors(errors.N)
	}
}

type result struct {
	Errors []errors.Error
}

func (r *result) AddErrors(errors ...errors.Error) {
	r.Errors = append(r.Errors, errors...)
}

func (r *result) IsValid() bool {
	return len(r.Errors) == 0
}

func (r *result) HasErrors() bool {
	return !r.IsValid()
}
