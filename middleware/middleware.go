package middleware

import (
	"github.com/alexraskin/trakt-tv-now-playing/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func AdminAuth() fiber.Handler {
	return keyauth.New(keyauth.Config{
		KeyLookup: "header:X-Admin-Key",
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			return key == models.Credentials.AdminKey, nil
		},
	})
}
