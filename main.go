package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ming Kwan API v1.0")
	})

	log.Fatal(app.Listen(":3000"))
}
