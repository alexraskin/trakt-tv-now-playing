package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexraskin/trakt-tv-now-playing/handlers"
	"github.com/alexraskin/trakt-tv-now-playing/middleware"
	"github.com/alexraskin/trakt-tv-now-playing/models"
	"github.com/alexraskin/trakt-tv-now-playing/service"
	"github.com/alexraskin/trakt-tv-now-playing/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	models.Credentials = models.Config{
		ClientID:     os.Getenv("TRAKT_CLIENT_ID"),
		ClientSecret: os.Getenv("TRAKT_CLIENT_SECRET"),
		AdminKey:     os.Getenv("ADMIN_KEY"),
	}

	if models.Credentials.ClientID == "" || models.Credentials.ClientSecret == "" {
		log.Fatal("TRAKT_CLIENT_ID and TRAKT_CLIENT_SECRET must be set")
	}

	if models.Credentials.AdminKey == "" {
		log.Fatal("ADMIN_KEY must be set")
	}

	err := utils.LoadToken()
	if err == nil && models.AccessToken != "" {
		fmt.Println("Loaded existing access token")

		// Check if token is expired or will expire soon (within 1 hour)
		if time.Now().Add(time.Hour).After(models.TokenExpiry) && models.RefreshToken != "" {
			fmt.Println("Token expired or will expire soon, attempting to refresh...")
			err := service.RefreshAccessToken()
			if err != nil {
				fmt.Printf("Failed to refresh token: %v\n", err)
			} else {
				fmt.Println("Access token refreshed successfully")
			}
		}
	} else {
		fmt.Println("No valid access token found. Use /admin/auth to authorize the application")
	}

	// refresh the token every 23 hours (tokens last 24 hours)
	go func() {
		for {
			// Sleep for 23 hours
			time.Sleep(23 * time.Hour)

			if models.RefreshToken != "" {
				fmt.Println("Performing scheduled token refresh...")
				err := service.RefreshAccessToken()
				if err != nil {
					fmt.Printf("Scheduled token refresh failed: %v\n", err)
				} else {
					fmt.Println("Scheduled token refresh successful")
				}
			}
		}
	}()

	app := fiber.New()

	app.Use(logger.New())

	// be nice to the api
	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded, please try again later",
			})
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Trakt.tv Now Playing API")
	})

	app.Get("/:username", handlers.HandleCheckWatching)

	admin := app.Group("/admin", middleware.AdminAuth())
	admin.Get("/auth", handlers.HandleAuth)
	admin.Get("/status", handlers.HandleAuthStatus)
	admin.Get("/refresh", handlers.HandleRefreshToken)

	app.Listen(":8080")
}
