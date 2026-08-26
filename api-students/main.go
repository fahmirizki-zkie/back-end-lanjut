package main 

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)


func main() {
    app := fiber.New()

    app.Use(func(c *fiber.Ctx) error {
        method := c.Method()

        if method == fiber.MethodPost ||
            method == fiber.MethodPut ||
            method == fiber.MethodPatch {

            contentType := c.Get("Content-Type")

            if !strings.HasPrefix(contentType, "application/json") {
                return fail(
                    c,
                    fiber.StatusUnsupportedMediaType,
                    "Content-Type harus application/json",
                )
            }
        }

        return c.Next()
    })

    api := app.Group("/api/v1")

    students := api.Group("/students")

    students.Get("/", listStudent)
    students.Get("/:id", getStudent)
    students.Post("/", createStudent)
    students.Put("/:id", replaceStudent)
    students.Patch("/:id", patchStudent)
    students.Delete("/:id", deleteStudent)

    log.Fatal(app.Listen(":3000"))
}