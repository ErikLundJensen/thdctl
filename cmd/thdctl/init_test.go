package thdctl

import (
	"net/url"
	"testing"

	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/stretchr/testify/mock"
)

// Mocking the robot.Client
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Get(path string) ([]byte, *robot.HTTPError) {
	args := m.Called(path)
	if args.Get(1) == nil {
		return args.Get(0).([]byte), nil
	}
	return args.Get(0).([]byte), &robot.HTTPError{StatusCode: 0, Message: "", Err: args.Get(1).(error)}
}

func (m *MockClient) Post(path string, data url.Values) ([]byte, *robot.HTTPError) {
	args := m.Called(path, data)
	if args.Get(1) == nil {
		return args.Get(0).([]byte), nil
	}
	return args.Get(0).([]byte), &robot.HTTPError{StatusCode: 0, Message: "", Err: args.Get(1).(error)}
}

// Mocking the SSHClientInterface
type MockSSHClient struct {
	mock.Mock
}

func (m *MockSSHClient) Auth(user, password string) error {
	args := m.Called(user, password)
	return args.Error(0)
}

func (m *MockSSHClient) WaitForReboot() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSSHClient) DownloadImage(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockSSHClient) InstallImage(disk string) (string, error) {
	args := m.Called(disk)
	return args.String(0), args.Error(1)
}

func (m *MockSSHClient) ListDisks() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockSSHClient) EstablishSSHSession() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSSHClient) SetTargetHost(host, port string) {
	m.Called(host, port)
}

func (m *MockSSHClient) ExecuteCommand(cmd string) (string, error) {
	args := m.Called(cmd)
	return args.String(0), args.Error(1)
}

func (m *MockSSHClient) ExecuteLSCommand() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestInitializeServer(t *testing.T) {
	mockClient := new(MockClient)
	serverNumber := 12345
	flags := cmdFlags{
		skipReboot:         false,
		enableRescueSystem: true,
		disk:               "nvme0n1",
		version:            "v1.9.2",
		image:              "",
	}

	// Mocking GetRescueSystemDetails
	mockClient.On("Get", mock.Anything).Return([]byte(`{"rescue": {"active": false}}`), nil)

	// Mocking EnableRescueSystem
	mockClient.On("Post", mock.Anything, mock.Anything).Return([]byte(`{"rescue": {"active": true, "password": "testpassword"}}`), nil)

	// Mocking RebootServer
	mockClient.On("Post", mock.Anything, mock.Anything).Return(nil, nil)

	// Mocking SSHClientInterface
	mockSSHClient := new(MockSSHClient)
	mockSSHClient.On("Auth", mock.Anything, mock.Anything).Return(nil)
	mockSSHClient.On("WaitForReboot").Return(true)
	mockSSHClient.On("DownloadImage", mock.Anything).Return("Downloaded", nil)
	mockSSHClient.On("InstallImage", mock.Anything).Return("Installed", nil)
	mockSSHClient.On("ListDisks", mock.Anything).Return("Disks", nil)
	mockSSHClient.On("SetTargetHost", mock.Anything, mock.Anything).Return(nil)

	// Call the function
	initializeServer(mockClient, mockSSHClient, serverNumber, flags)

	// Assertions
	mockClient.AssertExpectations(t)
	mockSSHClient.AssertExpectations(t)
}
