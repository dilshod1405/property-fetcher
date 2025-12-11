package httpclient

import (
	"fmt"
    "pf-service/config"

    "github.com/go-resty/resty/v2"
)

type TokenResponse struct {
    AccessToken string `json:"accessToken"`
    ExpiresIn   int    `json:"expiresIn"`
}

func GetJWTToken() (string, error) {
	client := resty.New()

	var resp TokenResponse
	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-PF-Client", config.AppConfig.PFAPIKey).
		SetBody(map[string]string{
			"apiKey":    config.AppConfig.PFAPIKey,
			"apiSecret": config.AppConfig.PFAPISecret,
		}).
		SetResult(&resp).
		Post(config.AppConfig.PFAPIUrl + "/auth/token")

	if err != nil {
		return "", err
	}

	if r.StatusCode() >= 300 {
		return "", fmt.Errorf("token error: %s", r.String())
	}

	return resp.AccessToken, nil
}
