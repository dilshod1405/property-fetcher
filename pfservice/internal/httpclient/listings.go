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

	// Debug: Log first listing's media structure to understand API response
	if len(resp.Results) > 0 {
		firstListing := resp.Results[0]
		fmt.Printf("DEBUG: First listing ID: %s\n", firstListing.ID)
		fmt.Printf("DEBUG: Media.Images count: %d\n", len(firstListing.Media.Images))
		if len(firstListing.Media.Images) > 0 {
			fmt.Printf("DEBUG: First image structure - URL: %s\n", firstListing.Media.Images[0].Original.URL)
		} else {
			fmt.Printf("DEBUG: No images in first listing, checking raw response...\n")
			// Log a sample of the response to see structure
			bodyStr := res.String()
			if len(bodyStr) > 500 {
				fmt.Printf("DEBUG: Response sample (first 500 chars): %s\n", bodyStr[:500])
			} else {
				fmt.Printf("DEBUG: Full response: %s\n", bodyStr)
			}
		}
	}

	return &resp, nil
}
