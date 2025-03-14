package thdctl

import (
	"errors"
	"net/url"
	"testing"

	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/stretchr/testify/assert"
)

type mockRobotClient struct {
	shouldFail bool
}

func (m *mockRobotClient) Get(path string) ([]byte, *robot.HTTPError) {
	response := `[
		{
			"server": {
				"server_number": 123456,
				"server_name": "test-server",
				"product": "EX42",
				"dc": "FSN1-DC8",
				"server_ip": "192.168.1.1",
				"server_ipv6_net": "2001:db8::/32"
			}
		}
	]`
	return []byte(response), nil
}

func (m *mockRobotClient) Post(path string, values url.Values) ([]byte, *robot.HTTPError) {
	if m.shouldFail {
		return nil, &robot.HTTPError{StatusCode: 0, Message: "", Err: errors.New("failed to reboot server")}
	}
	response := `{"status": "success"}`
	return []byte(response), nil
}

func TestListServers(t *testing.T) {
	var client robot.ClientInterface = &mockRobotClient{}
	servers, err := hetznerapi.ListServers(client)
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, 123456, servers[0].Server.ServerNumber)
	assert.Equal(t, "test-server", servers[0].Server.ServerName)
	assert.Equal(t, "EX42", servers[0].Server.Product)
	assert.Equal(t, "FSN1-DC8", servers[0].Server.Datacenter)
	assert.Equal(t, "192.168.1.1", servers[0].Server.ServerIP)
	assert.Equal(t, "2001:db8::/32", servers[0].Server.ServerIPv6Net)
}

func TestRebootServer(t *testing.T) {
	var client robot.ClientInterface = &mockRobotClient{}
	err := hetznerapi.RebootServer(client, 123456)
	assert.NoError(t, err)
}

func TestRebootServerError(t *testing.T) {
	client := &mockRobotClient{shouldFail: true}
	err := hetznerapi.RebootServer(client, 123456)
	assert.Error(t, err)
	assert.Equal(t, "failed to reboot server", err.Error())
}
