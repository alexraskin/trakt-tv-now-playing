package utils

import (
	"encoding/json"
	"os"

	"github.com/alexraskin/trakt-tv-now-playing/models"
)

func SaveToken() error {
	tokenFile := os.Getenv("TOKEN_FILE")
	if tokenFile == "" {
		tokenFile = "token.json" // Default value
	}

	tokenData := models.TokenData{
		AccessToken:  models.AccessToken,
		RefreshToken: models.RefreshToken,
		ExpiresAt:    models.TokenExpiry,
		DeviceCode:   models.DeviceCode,
	}

	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenFile, data, 0644)
}

func LoadToken() error {
	tokenFile := os.Getenv("TOKEN_FILE")
	if tokenFile == "" {
		tokenFile = "token.json" // Default value
	}

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return err
	}

	var tokenData models.TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return err
	}

	models.AccessToken = tokenData.AccessToken
	models.RefreshToken = tokenData.RefreshToken
	models.TokenExpiry = tokenData.ExpiresAt
	models.DeviceCode = tokenData.DeviceCode
	return nil
}
