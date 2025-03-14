package hetznerapi

import (
	"bytes"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SSHClientInterface interface {
	Auth(user, password string) error
	EstablishSSHSession() error
	ExecuteCommand(command string) (string, error)
	ExecuteLSCommand() (string, error)
	DownloadImage(url string) (string, error)
	InstallImage(disk string) (string, error)
	ListDisks() (string, error)
	WaitForReboot() bool
	SetTargetHost(host, port string)
}

type SSHClient struct {
	Host, Port string
	Session    *ssh.Session
	Config     *ssh.ClientConfig
}

func (client *SSHClient) Auth(user, password string) error {
	client.Config = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	return nil
}

func (client *SSHClient) EstablishSSHSession() error {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", client.Host, client.Port), client.Config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	client.Session = session
	return nil
}

func (client *SSHClient) ExecuteCommand(command string) (string, error) {
	var b bytes.Buffer
	if client.Session == nil {
		return "", fmt.Errorf("session is not established")
	}
	client.Session.Stdout = &b
	client.EstablishSSHSession()
	defer client.Session.Close()

	if err := client.Session.Run(command); err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}
	return b.String(), nil
}

func (client *SSHClient) ExecuteLSCommand() (string, error) {
	return client.ExecuteCommand("ls")
}

func (client *SSHClient) DownloadImage(url string) (string, error) {
	download := fmt.Sprintf("wget -O /tmp/talos.raw.xz %s", url)
	return client.ExecuteCommand(download)
}

func (client *SSHClient) ListDisks() (string, error) {
	return client.ExecuteCommand("lsblk")
}

func (client *SSHClient) VerifyDiskExists(disk string) (string, error) {
	return client.ExecuteCommand(fmt.Sprintf("lsblk | grep %s", disk))
}

func (client *SSHClient) InstallImage(disk string) (string, error) {
	unpack := fmt.Sprintf("zstdcat -dv /tmp/talos.raw.xz >/dev/%s", disk)
	return client.ExecuteCommand(unpack)
}

func (client *SSHClient) WaitForReboot() bool {
	maxRetries := 10
	retryInterval := 10 * time.Second

	for i := 0; i < maxRetries; i++ {
		logrus.WithFields(logrus.Fields{
			"attempt": i + 1,
			"host":    client.Host,
			"port":    client.Port,
		}).Info("Establishing SSH session")

		err := client.EstablishSSHSession()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"attempt": i + 1,
				"host":    client.Host,
				"port":    client.Port,
			}).Errorf("Error establishing SSH session: %v", err)
			if i < maxRetries-1 {
				logrus.Infof("Retrying in %s...", retryInterval)
				time.Sleep(retryInterval)
				continue
			}
			return false
		}
		return true
	}
	return false
}

func (client *SSHClient) SetTargetHost(host, port string) {
	client.Host = host
	client.Port = port
}
