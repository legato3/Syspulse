package models

import "strings"

// VendorManagedSystemRAIDDevices returns vendor-managed md arrays that should
// stay out of user-facing host state because they represent internal system
// volumes rather than customer-managed storage pools.
func VendorManagedSystemRAIDDevices(host Host) []string {
	switch {
	case hostMatchesVendorHint(host, "synology", "dsm"):
		return []string{"md0", "md1"}
	case hostMatchesVendorHint(host, "qnap", "qts", "quts"):
		return []string{"md9", "md13"}
	default:
		return nil
	}
}

// IsVendorManagedSystemRAIDArray reports whether an md array is one of the
// vendor-managed internal arrays that should be suppressed for this host.
func IsVendorManagedSystemRAIDArray(host Host, array HostRAIDArray) bool {
	device := normalizeHostRAIDDevice(array.Device)
	if device == "" {
		return false
	}

	for _, suppressed := range VendorManagedSystemRAIDDevices(host) {
		if device == suppressed {
			return true
		}
	}

	return false
}

// FilterVendorManagedSystemRAIDArrays removes vendor-managed system arrays from
// host state so every downstream consumer sees the same canonical storage view.
func FilterVendorManagedSystemRAIDArrays(host Host, arrays []HostRAIDArray) []HostRAIDArray {
	if len(arrays) == 0 {
		return arrays
	}

	filtered := make([]HostRAIDArray, 0, len(arrays))
	for _, array := range arrays {
		if IsVendorManagedSystemRAIDArray(host, array) {
			continue
		}
		filtered = append(filtered, array)
	}

	return filtered
}

func hostMatchesVendorHint(host Host, hints ...string) bool {
	fields := []string{
		host.Platform,
		host.OSName,
		host.OSVersion,
		host.DisplayName,
		host.Hostname,
	}

	for _, field := range fields {
		value := strings.ToLower(strings.TrimSpace(field))
		if value == "" {
			continue
		}
		for _, hint := range hints {
			if strings.Contains(value, hint) {
				return true
			}
		}
	}

	return false
}

func normalizeHostRAIDDevice(device string) string {
	device = strings.ToLower(strings.TrimSpace(device))
	return strings.TrimPrefix(device, "/dev/")
}
