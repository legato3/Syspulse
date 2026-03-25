package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/rcourtman/pulse-go-rewrite/internal/models"
	"github.com/rcourtman/pulse-go-rewrite/internal/resources"
)

// StateProvider allows ResourceHandlers to fetch current state on demand.
type StateProvider interface {
	GetState() models.StateSnapshot
}

// TenantStateProvider allows ResourceHandlers to fetch current state for a specific tenant.
type TenantStateProvider interface {
	GetStateForTenant(orgID string) models.StateSnapshot
}

// ResourceHandlers provides HTTP handlers for the unified resource API.
type ResourceHandlers struct {
	store               *resources.Store            // Default store (legacy/default tenant)
	storesByTenant      map[string]*resources.Store // Per-tenant stores
	storeMu             sync.RWMutex                // Protects storesByTenant
	stateProvider       StateProvider               // Legacy state provider
	tenantStateProvider TenantStateProvider         // Tenant-aware state provider
}

// NewResourceHandlers creates resource handlers with a new store.
func NewResourceHandlers() *ResourceHandlers {
	return &ResourceHandlers{
		store:          resources.NewStore(),
		storesByTenant: make(map[string]*resources.Store),
	}
}

// SetStateProvider sets the state provider for on-demand resource population.
func (h *ResourceHandlers) SetStateProvider(provider StateProvider) {
	h.stateProvider = provider
}

// SetTenantStateProvider sets the tenant-aware state provider.
func (h *ResourceHandlers) SetTenantStateProvider(provider TenantStateProvider) {
	h.tenantStateProvider = provider
}

// Store returns the underlying resource store for populating from the monitor.
func (h *ResourceHandlers) Store() *resources.Store {
	return h.store
}

// getStoreForTenant returns the resource store for a specific tenant.
// Creates a new store lazily if one doesn't exist.
func (h *ResourceHandlers) getStoreForTenant(orgID string) *resources.Store {
	if orgID == "" || orgID == "default" {
		return h.store
	}

	h.storeMu.RLock()
	store, exists := h.storesByTenant[orgID]
	h.storeMu.RUnlock()

	if exists {
		return store
	}

	// Create new store for tenant
	h.storeMu.Lock()
	defer h.storeMu.Unlock()

	// Double-check after acquiring write lock
	if store, exists = h.storesByTenant[orgID]; exists {
		return store
	}

	store = resources.NewStore()
	h.storesByTenant[orgID] = store
	return store
}

// GetStoreStats returns stats for all tenant stores.
func (h *ResourceHandlers) GetStoreStats() map[string]resources.StoreStats {
	h.storeMu.RLock()
	defer h.storeMu.RUnlock()

	stats := make(map[string]resources.StoreStats)
	stats["default"] = h.store.GetStats()

	for orgID, store := range h.storesByTenant {
		stats[orgID] = store.GetStats()
	}

	return stats
}

// HandleGetResources returns all resources, optionally filtered by query params.
// GET /api/resources
// Query params:
//   - type: filter by resource type (comma-separated)
//   - platform: filter by platform type (comma-separated)
//   - status: filter by status (comma-separated)
//   - parent: filter by parent ID
//   - infrastructure: if "true", only return infrastructure resources
//   - workloads: if "true", only return workload resources
func (h *ResourceHandlers) HandleGetResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := GetOrgID(r.Context())
	h.populateStoreFromCurrentState(orgID)
	store := h.getStoreForTenant(orgID)

	query := store.Query()

	// Parse type filter
	if typeParam := r.URL.Query().Get("type"); typeParam != "" {
		types := parseResourceTypes(typeParam)
		if len(types) > 0 {
			query = query.OfType(types...)
		}
	}

	// Parse platform filter
	if platformParam := r.URL.Query().Get("platform"); platformParam != "" {
		platforms := parsePlatformTypes(platformParam)
		if len(platforms) > 0 {
			query = query.FromPlatform(platforms...)
		}
	}

	// Parse status filter
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		statuses := parseStatuses(statusParam)
		if len(statuses) > 0 {
			query = query.WithStatus(statuses...)
		}
	}

	// Parse parent filter
	if parentID := r.URL.Query().Get("parent"); parentID != "" {
		query = query.WithParent(parentID)
	}

	// Parse alerts filter
	if r.URL.Query().Get("alerts") == "true" {
		query = query.WithAlerts()
	}

	// Execute query
	var result []resources.Resource

	// Handle infrastructure/workloads shortcut
	if r.URL.Query().Get("infrastructure") == "true" {
		result = store.GetInfrastructure()
	} else if r.URL.Query().Get("workloads") == "true" {
		result = store.GetWorkloads()
	} else {
		result = query.Execute()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResourcesResponse{
		Resources: result,
		Count:     len(result),
		Stats:     store.GetStats(),
	})
}

// HandleGetResource returns a single resource by ID.
// GET /api/resources/{id}
func (h *ResourceHandlers) HandleGetResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := GetOrgID(r.Context())
	h.populateStoreFromCurrentState(orgID)
	store := h.getStoreForTenant(orgID)

	// Extract ID from path: /api/resources/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/resources/")
	if path == "" || path == "/" {
		http.Error(w, "Resource ID required", http.StatusBadRequest)
		return
	}

	resource, ok := store.Get(path)
	if !ok {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

// HandleGetResourceStats returns statistics about the resource store.
// GET /api/resources/stats
func (h *ResourceHandlers) HandleGetResourceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := GetOrgID(r.Context())
	h.populateStoreFromCurrentState(orgID)
	store := h.getStoreForTenant(orgID)

	stats := store.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *ResourceHandlers) populateStoreFromCurrentState(orgID string) {
	if h.tenantStateProvider != nil && orgID != "" && orgID != "default" {
		h.PopulateFromSnapshotForTenant(orgID, h.tenantStateProvider.GetStateForTenant(orgID))
		return
	}
	if h.stateProvider != nil {
		h.PopulateFromSnapshot(h.stateProvider.GetState())
	}
}

// ResourcesResponse is the response for /api/resources.
type ResourcesResponse struct {
	Resources []resources.Resource `json:"resources"`
	Count     int                  `json:"count"`
	Stats     resources.StoreStats `json:"stats"`
}

// PopulateFromState converts all resources from a StateSnapshot to the unified store.
// This should be called whenever the state is updated.
func (h *ResourceHandlers) PopulateFromSnapshot(snapshot models.StateSnapshot) {
	h.store.PopulateFromSnapshot(snapshot)
}

// PopulateFromSnapshotForTenant converts all resources from a StateSnapshot to a tenant-specific store.
func (h *ResourceHandlers) PopulateFromSnapshotForTenant(orgID string, snapshot models.StateSnapshot) {
	h.getStoreForTenant(orgID).PopulateFromSnapshot(snapshot)
}

// Helper functions for parsing query parameters

func parseResourceTypes(s string) []resources.ResourceType {
	parts := strings.Split(s, ",")
	var result []resources.ResourceType
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, resources.ResourceType(p))
	}
	return result
}

func parsePlatformTypes(s string) []resources.PlatformType {
	parts := strings.Split(s, ",")
	var result []resources.PlatformType
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, resources.PlatformType(p))
	}
	return result
}

func parseStatuses(s string) []resources.ResourceStatus {
	parts := strings.Split(s, ",")
	var result []resources.ResourceStatus
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, resources.ResourceStatus(p))
	}
	return result
}
