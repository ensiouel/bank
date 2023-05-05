package apilayer

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	http   *http.Client
	apikey string
}

func New(apikey string) *Client {
	return &Client{
		http:   http.DefaultClient,
		apikey: apikey,
	}
}

type ConvertResponse struct {
	Date       string `json:"date"`
	Historical bool   `json:"historical"`
	Info       struct {
		Quote     float64 `json:"quote"`
		Timestamp int     `json:"timestamp"`
	} `json:"info"`
	Query struct {
		Amount int    `json:"amount"`
		From   string `json:"from"`
		To     string `json:"to"`
	} `json:"query"`
	Result  float64 `json:"result"`
	Success bool    `json:"success"`
}

func (client *Client) Convert(amount float64, from, to string) (float64, error) {
	u, _ := url.Parse("https://api.apilayer.com/currency_data/convert")

	v := url.Values{}
	v.Set("from", from)
	v.Set("to", to)
	v.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	u.RawQuery, _ = url.PathUnescape(v.Encode())

	request, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return 0.0, err
	}

	request.Header.Set("apikey", client.apikey)

	response, err := client.http.Do(request)
	if err != nil {
		return 0.0, err
	}

	var convertResponse ConvertResponse
	err = json.NewDecoder(response.Body).Decode(&convertResponse)
	if err != nil {
		return 0.0, err
	}

	if convertResponse.Success == false {
		return 0.0, errors.New("failed to convert")
	}

	return convertResponse.Result, nil
}
