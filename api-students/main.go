package main 

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app :=fiber.New()

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
