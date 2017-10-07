package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/labstack/echo"
)

func TestMain(m *testing.M) {
	requestValidator, _ := NewRequestValidator("petstore-expanded.json")

	// start server
	e := echo.New()

	// original middleware
	e.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if err := requestValidator.Validate(c); err != nil {
					fmt.Println(err)
					return c.String(http.StatusBadRequest, err.Error())
				}

				return next(c)
			}
		},
	)

	e.GET("/pets/:id", func(c echo.Context) error {
		return c.String(http.StatusOK, "pets "+c.Param("id"))
	})

	go func() {
		e.Start("127.0.0.1:8081")
	}()

	os.Exit(m.Run())
}

func TestValidate(t *testing.T) {
	expect := httpexpect.New(t, "http://127.0.0.1:8081")
	expect.GET("/pets/123").
		Expect().
		Status(http.StatusOK)

	expect.GET("/pets/abc").
		Expect().
		Status(http.StatusBadRequest)
}
