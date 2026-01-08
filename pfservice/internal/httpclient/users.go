package httpclient

import (
	"fmt"
	"pfservice/internal/users"

	"github.com/go-resty/resty/v2"
	"pfservice/config"
)

type PFUsersResponse struct {
	Data []users.PFUser `json:"data"`
}

func FetchAllUsers(token string) ([]users.PFUser, error) {
	client := resty.New()

	var resp PFUsersResponse

	r, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-PF-Client", config.AppConfig.PFAPIKey).
		SetResult(&resp).
		Get(config.AppConfig.PFAPIUrl + "/users")

	if err != nil {
		return nil, err
	}

	if r.StatusCode() >= 300 {
		return nil, fmt.Errorf("users API error: status %d, body: %s", r.StatusCode(), r.String())
	}

	return resp.Data, nil
}

