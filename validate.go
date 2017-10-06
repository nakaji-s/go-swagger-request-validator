package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/labstack/echo"
	"github.com/nakaji-s/runtime"
)

type RequestValidator struct {
	swagger *spec.Swagger
}

var pathParamRe = regexp.MustCompile(`:(.+?)/`)

func NewRequestValidator(filename string) (RequestValidator, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return RequestValidator{}, err
	}

	swagger := spec.Swagger{}
	swagger.UnmarshalJSON(data)
	if err != nil {
		return RequestValidator{}, err
	}

	return RequestValidator{&swagger}, nil
}

func (v RequestValidator) Validate(c echo.Context) error {
	// retrieve paramteters from request
	const sentinel = "/"
	matchedPathWithSentinel := pathParamRe.ReplaceAllString(c.Path()+sentinel, `{$1}/`)
	matchedPath := matchedPathWithSentinel[:len(matchedPathWithSentinel)-1]

	// create swagger path object
	method := c.Request().Method
	path := v.swagger.Paths.Paths[matchedPath]
	var operation *spec.Operation
	switch method {
	case "GET":
		operation = path.Get
	case "PUT":
		operation = path.Put
	case "POST":
		operation = path.Post
	case "DELETE":
		operation = path.Delete
	}
	if operation == nil {
		return c.NoContent(http.StatusNotFound)
	}

	// create requestValidator
	m := map[string]spec.Parameter{}
	for i, param := range operation.OperationProps.Parameters {
		m[fmt.Sprint(i)] = param
	}
	binder := middleware.NewUntypedRequestBinder(m, v.swagger, strfmt.Default)

	// get PathParams from request and set for validate
	pathParams := middleware.RouteParams{}
	for _, paramName := range c.ParamNames() {
		pathParams = append(pathParams, middleware.RouteParam{paramName, c.Param(paramName)})
	}

	// set Params if default value is needed
	data := map[string]interface{}{}

	// validate request and set defalut value to data
	err := binder.Bind(c.Request(), pathParams, runtime.JSONConsumer(), &data)
	if err != nil {
		var out []string

		// filtering error messages
		strs := strings.Split(err.Error(), "\n")
		for _, str := range strs {
			if str != "validation failure list:" {
				out = append(out, str)
			}
		}
		return fmt.Errorf(strings.Join(out, "\n"))
	}

	return nil
}
