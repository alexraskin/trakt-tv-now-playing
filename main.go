package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	apiKey := os.Getenv("TRAKT_CLIENT_ID")
	if apiKey == "" {
		log.Fatal("TRAKT_CLIENT_ID is not set")
	}

	app := fiber.New()

	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, OPTIONS",
	}))

	// be nice to the api
	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			cfIP := c.Get("CF-Connecting-IP")
			realIP := c.Get("X-Real-IP")
			forwardedFor := c.Get("X-Forwarded-For")
			fallbackIP := c.IP()
			var clientIP string
			switch {
			case cfIP != "":
				clientIP = cfIP
			case realIP != "":
				clientIP = realIP
			case forwardedFor != "":
				clientIP = strings.Split(forwardedFor, ",")[0] // take the first IP in the chain
			default:
				clientIP = fallbackIP
			}
			return clientIP
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

	app.Get("/:username", func(c *fiber.Ctx) error {
		return handleCheckWatching(c, apiKey)
	})

	app.Listen(":8080")
}

func handleCheckWatching(c *fiber.Ctx, traktApiKey string) error {
	username := c.Params("username")
	format := c.Query("format")

	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.trakt.tv/users/"+username+"/watching", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("trakt-api-key", traktApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trakt-api-version", "2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode == http.StatusNoContent {
		if format == "shields.io" {
			return c.JSON(fiber.Map{
				"schemaVersion": 1,
				"label":         "Currently Watching",
				"message":       "Nothing",
				"color":         "red",
			})
		}
		return c.JSON(fiber.Map{
			"watching": false,
		})
	}

	var watchingResponse TraktWatchingResponse
	if err := json.Unmarshal(body, &watchingResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if format == "shields.io" {
		message := "Nothing"
		if watchingResponse.Movie != nil {
			message = watchingResponse.Movie.Title
		} else if watchingResponse.Show != nil {
			if watchingResponse.Episode != nil {
				message = fmt.Sprintf("%s S%02dE%02d",
					watchingResponse.Show.Title,
					watchingResponse.Episode.Season,
					watchingResponse.Episode.Number)
			} else {
				message = watchingResponse.Show.Title
			}
		}

		return c.JSON(fiber.Map{
			"schemaVersion": 1,
			"label":         "Currently Watching",
			"message":       message,
			"color":         "green",
		})
	}

	return c.JSON(watchingResponse)
}
