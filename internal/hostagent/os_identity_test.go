package hostagent

import (
	"context"
	"os"
	"testing"

	gohost "github.com/shirou/gopsutil/v4/host"
)

func TestResolveHostOSIdentity(t *testing.T) {
	t.Run("detects synology dsm from version file when gopsutil is generic", func(t *testing.T) {
		mc := &mockCollector{
			goos: "linux",
			readFileFn: func(name string) ([]byte, error) {
				switch name {
				case "/etc.defaults/VERSION":
					return []byte(`majorversion="7"
minorversion="2"
productversion="7.2.2"
buildnumber="72806"
smallfixnumber="3"
`), nil
				default:
					return nil, os.ErrNotExist
				}
			},
		}

		name, version := resolveHostOSIdentity(mc, "linux", "")

		if name != "Synology DSM" {
			t.Fatalf("name = %q, want %q", name, "Synology DSM")
		}
		if version != "7.2.2-72806 Update 3" {
			t.Fatalf("version = %q, want %q", version, "7.2.2-72806 Update 3")
		}
	})

	t.Run("detects qnap from config file when gopsutil is generic", func(t *testing.T) {
		mc := &mockCollector{
			goos: "linux",
			readFileFn: func(name string) ([]byte, error) {
				switch name {
				case "/etc/config/uLinux.conf":
					return []byte(`Version = 5.2.0
Platform = QTS
`), nil
				default:
					return nil, os.ErrNotExist
				}
			},
		}

		name, version := resolveHostOSIdentity(mc, "linux", "")

		if name != "QNAP QTS" {
			t.Fatalf("name = %q, want %q", name, "QNAP QTS")
		}
		if version != "5.2.0" {
			t.Fatalf("version = %q, want %q", version, "5.2.0")
		}
	})

	t.Run("keeps existing identity when no vendor hint is present", func(t *testing.T) {
		mc := &mockCollector{goos: "linux"}

		name, version := resolveHostOSIdentity(mc, "ubuntu", "24.04")

		if name != "ubuntu" {
			t.Fatalf("name = %q, want %q", name, "ubuntu")
		}
		if version != "24.04" {
			t.Fatalf("version = %q, want %q", version, "24.04")
		}
	})

	t.Run("does not misclassify generic version files as synology", func(t *testing.T) {
		mc := &mockCollector{
			goos: "linux",
			readFileFn: func(name string) ([]byte, error) {
				switch name {
				case "/etc/VERSION":
					return []byte("VERSION=1\n"), nil
				default:
					return nil, os.ErrNotExist
				}
			},
		}

		name, version := resolveHostOSIdentity(mc, "linux", "")

		if name != "linux" {
			t.Fatalf("name = %q, want %q", name, "linux")
		}
		if version != "" {
			t.Fatalf("version = %q, want empty string", version)
		}
	})
}

func TestBuildReportDetectsSynologyDSMFromVersionFile(t *testing.T) {
	fixedReadFile := func(name string) ([]byte, error) {
		switch name {
		case "/etc/machine-id":
			return []byte("0123456789abcdef0123456789abcdef\n"), nil
		case "/etc.defaults/VERSION":
			return []byte(`majorversion="7"
minorversion="2"
productversion="7.2.2"
buildnumber="72806"
smallfixnumber="3"
`), nil
		default:
			return nil, os.ErrNotExist
		}
	}

	mc := &mockCollector{
		goos: "linux",
		hostInfoFn: func(context.Context) (*gohost.InfoStat, error) {
			return &gohost.InfoStat{
				Hostname:        "nas",
				HostID:          "",
				Platform:        "linux",
				PlatformFamily:  "linux",
				PlatformVersion: "",
				KernelVersion:   "4.4.302+",
				KernelArch:      "x86_64",
			}, nil
		},
		readFileFn: fixedReadFile,
	}

	agent, err := New(Config{
		APIToken:  "token",
		LogLevel:  -1,
		Collector: mc,
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	report, err := agent.buildReport(context.Background())
	if err != nil {
		t.Fatalf("buildReport() failed: %v", err)
	}

	if report.Host.OSName != "Synology DSM" {
		t.Fatalf("Host.OSName = %q, want %q", report.Host.OSName, "Synology DSM")
	}
	if report.Host.OSVersion != "7.2.2-72806 Update 3" {
		t.Fatalf("Host.OSVersion = %q, want %q", report.Host.OSVersion, "7.2.2-72806 Update 3")
	}
}
