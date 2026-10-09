package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/helper"
	"siakad-mini/middleware"
	"siakad-mini/route"
)

func NewApp(
	logger *slog.Logger,
	deps route.Dependencies,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "SIAKAD Mini API"),
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    2 * 1024 * 1024, // 2 MB
	})

	middleware.Register(app, GetEnv("ALLOWED_ORIGINS", "*"))
	route.Register(app, deps)

	// Fallback 404 untuk rute yang tidak ditemukan
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("Endpoint tidak ditemukan")
	})

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):
			// Error yang sudah terdefinisi dari helper
		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "Ukuran request body melebihi batas maksimal",
			}
		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Code:    "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		if appErr.Status >= fiber.StatusInternalServerError {
			causeMsg := ""
			if cause := appErr.Unwrap(); cause != nil {
				causeMsg = cause.Error()
			} else {
				causeMsg = appErr.Error()
			}

			logger.Error("server_error",
				slog.String("path", c.Path()),
				slog.String("method", c.Method()),
				slog.Int("status", appErr.Status),
				slog.String("error", causeMsg),
			)
		}

		resp := model.WebResponse{
			Success: false,
			Message: appErr.Message,
		}

		if len(appErr.Errors) > 0 {
			resp.Errors = appErr.Errors
		}

		return c.Status(appErr.Status).JSON(resp)
	}
}
