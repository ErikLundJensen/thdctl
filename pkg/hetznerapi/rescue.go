package hetznerapi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/eriklundjensen/thdctl/pkg/robot"
)

type RescueDetails struct {
	ServerIP      string        `json:"server_ip"`
	ServerIPv6Net string        `json:"server_ipv6_net"`
	ServerNumber  int           `json:"server_number"`
	Active        bool          `json:"active"`
	Password      string        `json:"password"`
	AuthorizedKey []interface{} `json:"authorized_key"`
	HostKey       []interface{} `json:"host_key"`
	BootTime      interface{}   `json:"boot_time"`
}

type Rescue struct {
	Rescue RescueDetails `json:"rescue"`
}

func GetRescueSystemDetails(client robot.ClientInterface, serverNumber int) (*Rescue, *robot.HTTPError) {
	path := fmt.Sprintf("boot/%d/rescue", serverNumber)

	body, err := client.Get(path)
	if err != nil {
		return nil, err
	}

	var rescue Rescue
	if err := json.Unmarshal(body, &rescue); err != nil {
		return nil, &robot.HTTPError{StatusCode: 0, Message: "failed to unmarshal response", Err: err}
	}

	return &rescue, nil
}

func EnableRescueSystem(client robot.ClientInterface, serverNumber int) (*Rescue, *robot.HTTPError) {
	path := fmt.Sprintf("boot/%d/rescue", serverNumber)

	data := url.Values{}
	data.Set("os", "linux")

	body, err := client.Post(path, data)
	if err != nil {
		return nil, err
	}

	var rescue Rescue
	if err := json.Unmarshal(body, &rescue); err != nil {
		return nil, &robot.HTTPError{StatusCode: 0, Message: "failed to unmarshal response", Err: err}
	}

	fmt.Println("Parsed Response:", rescue)
	fmt.Println("Rescue system enabled successfully.")
	return &rescue, nil
}
