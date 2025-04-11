package hetznerapi

import (
	"bufio"
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
)

type DiskInfo struct {
	Name       string
	Size       string
	Type       string
	Mountpoint string
}

// ParseLSBLKOutput parses the output of the `lsblk` command and extracts disk information.
func ParseLSBLKOutput(output string) ([]DiskInfo, error) {
	var disks []DiskInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip the header line
	if scanner.Scan() {
		// Header line is ignored
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip lines with special characters like "├─" or "└─"
		if strings.Contains(line, "├─") || strings.Contains(line, "└─") {
			continue
		}

		fields := strings.Fields(line)

		// Ensure there are enough fields to parse
		if len(fields) >= 6 {
			mountpoint := ""
			if len(fields) > 6 {
				mountpoint = strings.Join(fields[6:], " ")
			}
			disks = append(disks, DiskInfo{
				Name:       fields[0],
				Size:       fields[3],
				Type:       fields[5],
				Mountpoint: mountpoint,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return disks, nil
}

// LogAsJSON logs the DiskInfo slice as a JSON structure using logrus.
func LogAsJSON(disks []DiskInfo) {
	jsonData, err := json.Marshal(disks)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal DiskInfo to JSON")
		return
	}
	logrus.WithField("disks", string(jsonData)).Info("Disk information")
}
