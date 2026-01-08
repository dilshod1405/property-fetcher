package httpclient

import (
	"fmt"
	"pfservice/config"
	"pfservice/internal/property"

	"github.com/go-resty/resty/v2"
)

type ListingsResponse struct {
	Results []property.PFListing `json:"results"`
}

func FetchListings(token string, page int) (*ListingsResponse, error) {
	client := resty.New()

	var resp ListingsResponse

	res, err := client.R().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + token,
			"X-PF-Client":   config.AppConfig.PFAPIKey,
		}).
		SetQueryParams(map[string]string{
			"page":    fmt.Sprintf("%d", page),
			"perPage": "50",
		}).
		SetResult(&resp).
		Get(config.AppConfig.PFAPIUrl + "/listings")

	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 300 {
		return nil, fmt.Errorf("listings API error: status %d, body: %s", res.StatusCode(), res.String())
	}

	fmt.Println("STATUS:", res.Status())
	fmt.Println("BODY:", res.String())

	return &resp, nil
}
