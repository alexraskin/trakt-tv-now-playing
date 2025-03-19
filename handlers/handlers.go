package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexraskin/trakt-tv-now-playing/models"
	"github.com/alexraskin/trakt-tv-now-playing/service"
	"github.com/alexraskin/trakt-tv-now-playing/utils"

	"github.com/gofiber/fiber/v2"
)

func HandleCheckWatching(c *fiber.Ctx) error {
	username := c.Params("username")
	format := c.Query("format")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}

	username = strings.ToLower(username)

	if models.AccessToken == "" {
		err := utils.LoadToken()
		if err != nil || time.Now().After(models.TokenExpiry) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Service not authorized. Please contact administrator.",
			})
		}
	}

	url := fmt.Sprintf("%s/users/%s/watching", models.BaseURL, username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", models.AccessToken))
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", os.Getenv("TRAKT_CLIENT_ID"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if resp.StatusCode == http.StatusNotFound {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found. Please check the username.",
		})
	}

	if resp.StatusCode == http.StatusNoContent {
		return c.JSON(fiber.Map{
			"watching": false,
		})
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token. Please send me a message on twitter @notalexraskin",
		})
	}

	var watchingResponse models.TraktWatchingResponse
	if err := json.Unmarshal(body, &watchingResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        fmt.Sprintf("Failed to parse response: %v", err),
			"raw_response": string(body),
		})
	}
	message := fmt.Sprintf("%s - %d", watchingResponse.Movie.Title, watchingResponse.Movie.Year)
	label := "Currently Watching"
	if watchingResponse.Movie == nil {
		message = fmt.Sprintf("%s - %d", watchingResponse.Show.Title, watchingResponse.Show.Year)
		label = "Currently Watching"
	}
	if format == "shields.io" {
		return c.JSON(fiber.Map{
			"SchemaVersion": 1,
			"Label":         label,
			"Message":       message,
		})
	}

	return c.JSON(fiber.Map{
		"watching": true,
		"data":     watchingResponse,
	})
}

func HandleAuthStatus(c *fiber.Ctx) error {
	if models.AccessToken == "" {
		err := utils.LoadToken()
		if err != nil {
			return c.JSON(fiber.Map{
				"status":  "not_authorized",
				"message": "Not authorized",
			})
		}
	}

	if time.Now().Before(models.TokenExpiry) {
		return c.JSON(fiber.Map{
			"status":     "authorized",
			"expires_at": models.TokenExpiry,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "expired",
		"message": "Token expired",
	})
}

func HandleAuth(c *fiber.Ctx) error {
	if models.AccessToken != "" && time.Now().Before(models.TokenExpiry) {
		return c.JSON(fiber.Map{
			"status":     "already_authorized",
			"message":    "Application is already authorized",
			"expires_at": models.TokenExpiry,
		})
	}

	if models.RefreshToken != "" {
		err := service.RefreshAccessToken()
		if err == nil {
			return c.JSON(fiber.Map{
				"status":     "refreshed",
				"message":    "Successfully refreshed access token",
				"expires_at": models.TokenExpiry,
			})
		}
		fmt.Printf("Failed to refresh token: %v\n", err)
		// Fall through to get a new device code
	}

	reqBody := fmt.Sprintf("client_id=%s", models.Credentials.ClientID)
	resp, err := http.Post(models.DeviceCodeURL, "application/x-www-form-urlencoded", strings.NewReader(reqBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer resp.Body.Close()

	var deviceResp models.TraktDeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	models.DeviceCode = deviceResp.DeviceCode
	if err := utils.SaveToken(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Immediately start to poll for the token
	go service.PollForAuthorization(deviceResp.Interval, deviceResp.ExpiresIn)

	return c.JSON(fiber.Map{
		"status":           "pending",
		"message":          "Please visit the verification URL and enter the code. The server will automatically check for authorization.",
		"verification_url": deviceResp.VerificationURL,
		"user_code":        deviceResp.UserCode,
		"expires_in":       deviceResp.ExpiresIn,
		"interval":         deviceResp.Interval,
	})
}

func HandleRefreshToken(c *fiber.Ctx) error {
	if models.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No refresh token available",
		})
	}

	err := service.RefreshAccessToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "Token refreshed successfully",
		"expires_at": models.TokenExpiry,
	})
}
