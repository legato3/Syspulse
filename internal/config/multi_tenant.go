package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rcourtman/pulse-go-rewrite/internal/models"
	"github.com/rcourtman/pulse-go-rewrite/internal/pathutil"
	"github.com/rs/zerolog/log"
)

// MultiTenantPersistence manages a collection of TenantPersistence instances,
// one for each organization.
type MultiTenantPersistence struct {
	baseDataDir string
	mu          sync.RWMutex
	tenants     map[string]*ConfigPersistence
}

// NewMultiTenantPersistence creates a new multi-tenant persistence manager.
func NewMultiTenantPersistence(baseDataDir string) *MultiTenantPersistence {
	if normalized, err := pathutil.NormalizeDir(baseDataDir); err == nil {
		baseDataDir = normalized
	}
	return &MultiTenantPersistence{
		baseDataDir: baseDataDir,
		tenants:     make(map[string]*ConfigPersistence),
	}
}

// GetPersistence returns the persistence instance for a specific organization.
// It initializes the persistence if it hasn't been loaded yet.
func (mtp *MultiTenantPersistence) GetPersistence(orgID string) (*ConfigPersistence, error) {
	mtp.mu.RLock()
	persistence, exists := mtp.tenants[orgID]
	mtp.mu.RUnlock()

	if exists {
		return persistence, nil
	}

	mtp.mu.Lock()
	defer mtp.mu.Unlock()

	// Double-check locking pattern
	if persistence, exists = mtp.tenants[orgID]; exists {
		return persistence, nil
	}

	// Validate OrgID (prevent directory traversal)
	if filepath.Base(orgID) != orgID || orgID == "" || orgID == "." || orgID == ".." {
		return nil, fmt.Errorf("invalid organization ID: %s", orgID)
	}

	orgDir, err := mtp.orgDir(orgID)
	if err != nil {
		return nil, err
	}

	log.Info().Str("org_id", orgID).Str("dir", orgDir).Msg("Initializing tenant persistence")

	cp, err := newConfigPersistence(orgDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize persistence for org %s: %w", orgID, err)
	}

	// Ensure the directory exists
	if err := cp.EnsureConfigDir(); err != nil {
		return nil, err
	}

	mtp.tenants[orgID] = cp
	return cp, nil
}

// OrgExists checks if an organization exists (directory exists) without creating it.
func (mtp *MultiTenantPersistence) OrgExists(orgID string) bool {
	if orgID == "default" {
		return true
	}

	// Validate to prevent traversal
	if filepath.Base(orgID) != orgID || orgID == "" || orgID == "." || orgID == ".." {
		return false
	}

	orgDir, err := mtp.orgDir(orgID)
	if err != nil {
		return false
	}
	stat, err := os.Stat(orgDir)
	return err == nil && stat.IsDir()
}

func (mtp *MultiTenantPersistence) orgDir(orgID string) (string, error) {
	if orgID == "default" {
		return mtp.baseDataDir, nil
	}
	orgsRoot, err := pathutil.JoinBaseFile(mtp.baseDataDir, "orgs")
	if err != nil {
		return "", fmt.Errorf("failed to resolve org root: %w", err)
	}
	return pathutil.JoinBaseFile(orgsRoot, orgID)
}

// LoadOrganization loads the organization metadata including members.
// Org metadata is stored in <orgDir>/org.json.
func (mtp *MultiTenantPersistence) LoadOrganization(orgID string) (*models.Organization, error) {
	persistence, err := mtp.GetPersistence(orgID)
	if err != nil {
		return nil, err
	}

	org, err := persistence.LoadOrganization()
	if err != nil {
		// If org.json doesn't exist, return a default org
		return &models.Organization{
			ID:          orgID,
			DisplayName: orgID,
		}, nil
	}

	return org, nil
}

// SaveOrganization saves the organization metadata.
func (mtp *MultiTenantPersistence) SaveOrganization(org *models.Organization) error {
	persistence, err := mtp.GetPersistence(org.ID)
	if err != nil {
		return err
	}

	return persistence.SaveOrganization(org)
}
