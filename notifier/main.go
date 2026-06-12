package main

import (
	"devinsync/routes"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	routes.Init(app)

	log.Fatal(app.Listen(":8080"))
}
