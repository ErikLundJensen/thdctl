package hetznerapi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestParseLSBLKOutput(t *testing.T) {

	lsblkOutput := `NAME   MAJ:MIN RM   SIZE RO TYPE MOUNTPOINTS
loop0    7:0    0   3.4G  1 loop 
sda      8:0    0 465.8G  0 disk 
├─sda1   8:1    0   100M  0 part 
├─sda2   8:2    0     1M  0 part 
├─sda3   8:3    0  1000M  0 part 
└─sda4   8:4    0     1M  0 part 
sdb      8:16   0 465.8G  0 disk 
├─sdb1   8:17   0   100M  0 part 
├─sdb2   8:18   0     1M  0 part 
├─sdb3   8:19   0  1000M  0 part 
└─sdb4   8:20   0     1M  0 part`

	expected := []DiskInfo{
		{Name: "loop0", Size: "3.4G", Type: "loop", Mountpoint: ""},
		{Name: "sda", Size: "465.8G", Type: "disk", Mountpoint: ""},
		{Name: "sdb", Size: "465.8G", Type: "disk", Mountpoint: ""},
	}

	disks, err := ParseLSBLKOutput(lsblkOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Compare only top-level disks (loop0, sda, sdb)
	if !reflect.DeepEqual(disks, expected) {
		t.Errorf("expected %v, got %v", expected, disks)
	}
}

func TestLogAsJSON(t *testing.T) {
	disks := []DiskInfo{
		{Name: "loop0", Size: "3.4G", Type: "loop", Mountpoint: ""},
		{Name: "sda", Size: "465.8G", Type: "disk", Mountpoint: ""},
		{Name: "sdb", Size: "465.8G", Type: "disk", Mountpoint: ""},
	}

	// Capture log output
	var logOutput strings.Builder
	logrus.SetOutput(&logOutput)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// Call LogAsJSON
	LogAsJSON(disks)

	// Verify log contains JSON representation of disks
	expectedSubstring := `\"Name\":\"loop0\"`
	if !strings.Contains(logOutput.String(), expectedSubstring) {
		t.Errorf("expected log to contain %s. Then content is %s", expectedSubstring, logOutput.String())
	}
}
