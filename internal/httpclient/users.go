package httpclient

import (
	"pf-service/internal/users"

	"github.com/go-resty/resty/v2"
	"pf-service/config"
)

type PFUsersResponse struct {
	Data []users.PFUser `json:"data"`
}

func FetchAllUsers(token string) ([]users.PFUser, error) {
	client := resty.New()

	var resp PFUsersResponse

	_, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-PF-Client", config.AppConfig.PFAPIKey).
		SetResult(&resp).
		Get(config.AppConfig.PFAPIUrl + "/users")

	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

