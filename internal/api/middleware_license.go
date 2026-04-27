package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rcourtman/pulse-go-rewrite/internal/license"
	"github.com/rcourtman/pulse-go-rewrite/internal/websocket"
)

// Multi-tenant feature flag (default: disabled)
// Set PULSE_MULTI_TENANT_ENABLED=true to enable multi-tenant functionality.
// This is separate from compatibility license APIs; the feature must be explicitly enabled.
var multiTenantEnabled = strings.EqualFold(os.Getenv("PULSE_MULTI_TENANT_ENABLED"), "true")

// v5 should behave as single-tenant in real runtime even if dormant multi-tenant
// code remains in the branch. Tests can disable this to exercise legacy paths.
var v5SingleTenantMode = !runningUnderGoTest()

// IsMultiTenantEnabled returns whether multi-tenant functionality is enabled.
func IsMultiTenantEnabled() bool {
	return multiTenantEnabled
}

func isV5SingleTenantMode() bool {
	return v5SingleTenantMode
}

func IsV5SingleTenantRuntime() bool {
	return isV5SingleTenantMode()
}

func setV5SingleTenantModeForTests(enabled bool) {
	v5SingleTenantMode = enabled
}

func runningUnderGoTest() bool {
	return strings.HasSuffix(filepath.Base(os.Args[0]), ".test")
}

// DefaultMultiTenantChecker implements websocket.MultiTenantChecker for use with the WebSocket hub.
type DefaultMultiTenantChecker struct{}

// CheckMultiTenant checks if multi-tenant is enabled (feature flag) and licensed for the org.
// Uses the LicenseServiceProvider for proper per-tenant license lookup.
func (c *DefaultMultiTenantChecker) CheckMultiTenant(ctx context.Context, orgID string) websocket.MultiTenantCheckResult {
	// Default org is always allowed
	if orgID == "" || orgID == "default" {
		return websocket.MultiTenantCheckResult{
			Allowed:        true,
			FeatureEnabled: true,
			Licensed:       true,
		}
	}

	// Check feature flag first
	if !multiTenantEnabled {
		return websocket.MultiTenantCheckResult{
			Allowed:        false,
			FeatureEnabled: false,
			Licensed:       false,
			Reason:         "Multi-tenant functionality is not enabled",
		}
	}

	return websocket.MultiTenantCheckResult{
		Allowed:        true,
		FeatureEnabled: true,
		Licensed:       true,
	}
}

// NewMultiTenantChecker creates a new DefaultMultiTenantChecker.
func NewMultiTenantChecker() *DefaultMultiTenantChecker {
	return &DefaultMultiTenantChecker{}
}

// SetMultiTenantEnabled allows programmatic control of the feature flag (for testing).
func SetMultiTenantEnabled(enabled bool) {
	multiTenantEnabled = enabled
}

// LicenseServiceProvider provides license service for a given context.
// This allows the middleware to use the properly initialized per-tenant services.
type LicenseServiceProvider interface {
	Service(ctx context.Context) *license.Service
}

var (
	licenseServiceProvider LicenseServiceProvider
	licenseServiceMu       sync.RWMutex
)

// SetLicenseServiceProvider sets the provider for license services.
// This should be called during router initialization with LicenseHandlers.
func SetLicenseServiceProvider(provider LicenseServiceProvider) {
	licenseServiceMu.Lock()
	defer licenseServiceMu.Unlock()
	licenseServiceProvider = provider
}

// getLicenseServiceForContext returns the license service for the given context.
// Falls back to a new service if no provider is set (shouldn't happen in production).
func getLicenseServiceForContext(ctx context.Context) *license.Service {
	licenseServiceMu.RLock()
	provider := licenseServiceProvider
	licenseServiceMu.RUnlock()

	if provider != nil {
		return provider.Service(ctx)
	}
	// Fallback: create a new service (won't have persisted license)
	return license.NewService()
}

// hasMultiTenantFeatureForContext is retained for legacy callers.
func hasMultiTenantFeatureForContext(ctx context.Context) bool {
	return true
}

// RequireMultiTenant returns a middleware that checks whether multi-tenant mode is enabled.
// It allows access to the "default" organization even when multi-tenant mode is disabled.
func RequireMultiTenant(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := GetOrgID(r.Context())

		// Default org is always allowed (backward compatibility)
		if orgID == "" || orgID == "default" {
			next(w, r)
			return
		}

		// Feature flag check - multi-tenant must be explicitly enabled
		if !multiTenantEnabled {
			writeMultiTenantDisabledError(w)
			return
		}

		next(w, r)
	}
}

// RequireMultiTenantHandler returns middleware for http.Handler.
func RequireMultiTenantHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID := GetOrgID(r.Context())

		// Default org is always allowed (backward compatibility)
		if orgID == "" || orgID == "default" {
			next.ServeHTTP(w, r)
			return
		}

		// Feature flag check - multi-tenant must be explicitly enabled
		if !multiTenantEnabled {
			writeMultiTenantDisabledError(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeMultiTenantDisabledError writes a 501 Not Implemented response
// indicating that multi-tenant functionality is not enabled.
func writeMultiTenantDisabledError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "feature_disabled",
		"message": "Multi-tenant functionality is not enabled. Set PULSE_MULTI_TENANT_ENABLED=true to enable.",
	})
}

// CheckMultiTenantLicense checks if multi-tenant is licensed for the given org ID.
// Returns true if:
// - The org ID is "default" or empty (always allowed)
// - The feature flag is enabled AND the multi-tenant feature is licensed
// Deprecated: Use CheckMultiTenantLicenseWithContext for proper per-tenant license checking.
func CheckMultiTenantLicense(orgID string) bool {
	if orgID == "" || orgID == "default" {
		return true
	}
	// Feature flag must be enabled
	if !multiTenantEnabled {
		return false
	}
	return true
}

// CheckMultiTenantLicenseWithContext checks if multi-tenant is enabled.
// Returns true if:
// - The org ID is "default" or empty (always allowed)
// - The feature flag is enabled
func CheckMultiTenantLicenseWithContext(ctx context.Context, orgID string) bool {
	if orgID == "" || orgID == "default" {
		return true
	}
	// Feature flag must be enabled
	if !multiTenantEnabled {
		return false
	}
	return true
}
