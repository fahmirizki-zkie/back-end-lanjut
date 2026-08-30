package main

import (
	"github.com/gofiber/fiber/v2"
)

// response berhasil
func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// response daftar
func oklist(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// response 201 created
func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// response 204 no content
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// response gagal
func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}

// response validasi 
func failValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Error:   errs,
	})
}

