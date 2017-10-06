package main

import (
	"net/http"

	"fmt"

	"github.com/labstack/echo"
)

func main() {
	requestValidator, err := NewRequestValidator("petstore-expanded.json")
	if err != nil {
		panic(err)
	}

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

	e.Logger.Fatal(e.Start("127.0.0.1:8081"))
}
