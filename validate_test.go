package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func TestMain(m *testing.M) {
	requestValidator, _ := NewRequestValidator("petstore-expanded.json")

	// start server
	e := echo.New()

	// body dump
	e.Pre(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		fmt.Println(string(reqBody))
	}))

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
	e.POST("/pets", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	go func() {
		e.Start("127.0.0.1:8081")
	}()

	os.Exit(m.Run())
}

func newString(s string) *string {
	return &s
}

type NewPet struct {
	Name *string `json:"name,omitempty"`
	Tag  *string `json:"tag,omitempty"`
}

func TestValidate(t *testing.T) {
	expect := httpexpect.New(t, "http://127.0.0.1:8081")
	expect.GET("/pets/123").
		Expect().
		Status(http.StatusOK)

	expect.GET("/pets/abc").
		Expect().
		Status(http.StatusBadRequest)

	expect.POST("/pets").WithJSON(NewPet{Name: newString("pochi")}).
		Expect().
		Status(http.StatusOK)
}
