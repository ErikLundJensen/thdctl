package hetznerapi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/eriklundjensen/thdctl/pkg/robot"
)

type FirewallRule struct {
	Direction string `json:"direction"`
	SrcIP     string `json:"src_ip"`
	DstIP     string `json:"dst_ip"`
	Protocol  string `json:"protocol"`
	SrcPort   string `json:"src_port"`
	DstPort   string `json:"dst_port"`
	Action    string `json:"action"`
	TCPFlags  string `json:"tcp_flags"`
}

type FirewallSet struct {
	ServerIP                 string         `json:"server_ip"`
	ServerNumber             int            `json:"server_number"`
	Status                   string         `json:"status"`
	FilterIPv6               bool           `json:"filter_ipv6"`
	WhitelistHetznerServices bool           `json:"whitelist_hos"`
	Port                     string         `json:"port"`
	Rules                    []FirewallRule `json:"rules"`
}

type FirewallTemplate struct {
	ID                       string         `json:"id"`
	Name                     string         `json:"name"`
	FilterIPv6               bool           `json:"filter_ipv6"`
	WhitelistHetznerServices bool           `json:"whitelist_hos"`
	Default                  bool           `json:"is_default"`
	Rules                    []FirewallRule `json:"rules"`
}

func GetFirewallRules(client robot.ClientInterface, serverNumber int) (*FirewallSet, *robot.HTTPError) {
	path := fmt.Sprintf("firewall/%d", serverNumber)
	body, err := client.Get(path)
	if err != nil {
		return nil, err
	}

	var firewall FirewallSet
	if err := json.Unmarshal(body, &firewall); err != nil {
		return nil, &robot.HTTPError{StatusCode: 0, Message: "failed to unmarshal response", Err: err}
	}

	return &firewall, nil
}

func GetFirewallTemplates(client robot.ClientInterface) ([]FirewallTemplate, *robot.HTTPError) {
	path := "firewall/template"

	body, err := client.Get(path)
	if err != nil {
		return nil, err
	}

	var templates []FirewallTemplate
	if err := json.Unmarshal(body, &templates); err != nil {
		return nil, &robot.HTTPError{StatusCode: 0, Message: "failed to unmarshal response", Err: err}
	}

	return templates, nil
}

func CreateFirewallRule(client robot.ClientInterface, serverNumber int, cfg FirewallSet) *robot.HTTPError {
	path := fmt.Sprintf("firewall/%d", serverNumber)

	data := url.Values{}
	data.Set("status", cfg.Status)
	// TODO encoding of booleans and array of rules
	//data.Set("filter_ipv6", cfg.FilterIPv6)
	//data.Set("whitelist_hos", cfg.WhitelistHetznerServices)
	//data.Set("rules", rules)

	_, err := client.Post(path, data)
	if err != nil {
		return err
	}

	fmt.Println("Firewall rule created successfully.")
	return nil
}
