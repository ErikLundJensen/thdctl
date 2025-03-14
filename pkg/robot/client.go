package robot

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

var HETZNER_SERVER_URL = "https://robot-ws.your-server.de"

// ClientInterface defines the methods that a client must implement
type ClientInterface interface {
	Get(path string) ([]byte, *HTTPError)
	Post(path string, values url.Values) ([]byte, *HTTPError)
}

type Client struct {
	Username string
	Password string
}

// Invoke GET HTTP request using Hetzner API. Path is added to the base URL.
func (c Client) Get(path string) ([]byte, *HTTPError) {
	return c.MakeRequest("GET", path, nil)
}

// Invoke POST HTTP request using Hetzner API. Path is added to the base URL.
func (c Client) Post(path string, values url.Values) ([]byte, *HTTPError) {
	return c.MakeRequest("POST", path, values)
}

// Invoke Hetzner API. Path is added to the base URL.
func (c Client) MakeRequest(action, path string, values url.Values) ([]byte, *HTTPError) {
	url := fmt.Sprintf("%s/%s", HETZNER_SERVER_URL, path)

	var parameters io.Reader
	if values != nil {
		parameters = strings.NewReader(values.Encode())
	}

	req, err := http.NewRequest(action, url, parameters)
	if err != nil {
		return nil, &HTTPError{0, "failed to create request", err}
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &HTTPError{0, "failed to send request", err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{resp.StatusCode, fmt.Sprintf("%v", resp.Status), nil}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &HTTPError{0, "failed to read response body", err}
	}

	logrus.WithFields(logrus.Fields{
		"action": action,
		"path":   path,
	}).Debugf("Response Body: %s", string(body))
	return body, nil
}
