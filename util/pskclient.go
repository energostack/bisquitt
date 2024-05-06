package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GetPSKKeyFromAPI retrieves a pre-shared key from the PSKKeys API.
func GetPSKKeyFromAPI(hint string, endpoint string, username string, password string, timeout time.Duration, logger Logger) ([]byte, bool) {
	req, err := http.NewRequest("GET", fmt.Sprintf(endpoint+"/%s", hint), nil)
	if err != nil {
		logger.Error("Error in creating request: %s", err)
		return nil, false
	}

	req.SetBasicAuth(username, password)
	client := &http.Client{}
	client.Timeout = timeout
	resp, err := client.Do(req)

	if err != nil {
		logger.Error("Error in sending request: %s", err)
		return nil, false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		logger.Debug("ID not found")
		return nil, false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error in reading response body: %s", err)
		return nil, false
	}

	type Response struct {
		Client string `json:"client"`
		PSK    []byte `json:"psk"`
	}

	var response Response

	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Error("Error in unmarshalling response body: %s", err)
		return nil, false
	}

	return response.PSK, true
}
