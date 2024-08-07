package handler

import "github.com/gofiber/fiber/v2"

// Healthz of Web Server Go
func Healthz(c *fiber.Ctx) error {
	return c.SendString("WebP Server Go up and running!😁")
}
