package resources

import (
	"testing"
	"time"

	"github.com/rcourtman/pulse-go-rewrite/internal/models"
)

func TestStorePopulateFromSnapshot(t *testing.T) {
	store := NewStore()
	now := time.Now()

	// Create a minimal snapshot with test data
	snapshot := models.StateSnapshot{
		Nodes: []models.Node{
			{
				ID:       "node-1",
				Name:     "pve-node-1",
				Instance: "https://pve1.local:8006",
				Status:   "online",
				CPU:      0.455, // Proxmox API returns 0-1 ratio
				Memory: models.Memory{
					Total: 16000000000,
					Used:  8000000000,
					Free:  8000000000,
					Usage: 50.0,
				},
				Uptime:   86400,
				LastSeen: now,
			},
		},
		VMs: []models.VM{
			{
				ID:       "vm-100",
				VMID:     100,
				Name:     "test-vm",
				Node:     "pve-node-1",
				Instance: "https://pve1.local:8006",
				Status:   "running",
				CPU:      0.25, // Proxmox API returns 0-1 ratio
				Memory: models.Memory{
					Total: 4000000000,
					Used:  2000000000,
					Free:  2000000000,
					Usage: 50.0,
				},
			},
		},
		Containers: []models.Container{
			{
				ID:       "ct-200",
				VMID:     200,
				Name:     "test-container",
				Node:     "pve-node-1",
				Instance: "https://pve1.local:8006",
				Status:   "running",
			},
		},
		Hosts: []models.Host{
			{
				ID:       "host-1",
				Hostname: "my-host",
				Status:   "online",
			},
		},
		DockerHosts: []models.DockerHost{
			{
				ID:       "docker-host-1",
				Hostname: "docker-host-1",
				Status:   "online",
				Containers: []models.DockerContainer{
					{
						ID:    "container-1",
						Name:  "web",
						State: "running",
					},
				},
			},
		},
		ActiveAlerts: []models.Alert{
			{
				ID:         "vm-100-cpu",
				ResourceID: "vm-100",
				Type:       "cpu",
				Level:      "warning",
				Message:    "VM CPU high",
				Value:      92,
				Threshold:  80,
				StartTime:  now,
			},
			{
				ID:         "host-1-cpu",
				ResourceID: "host:host-1",
				Type:       "cpu",
				Level:      "warning",
				Message:    "Host CPU high",
				Value:      95,
				Threshold:  80,
				StartTime:  now,
			},
			{
				ID:         "host-1-disk",
				ResourceID: "host:host-1/disk:root",
				Type:       "disk",
				Level:      "critical",
				Message:    "Host disk high",
				Value:      98,
				Threshold:  90,
				StartTime:  now,
			},
			{
				ID:         "docker-container-state",
				ResourceID: "docker:docker-host-1/container-1",
				Type:       "state",
				Level:      "critical",
				Message:    "Container stopped",
				Value:      0,
				Threshold:  1,
				StartTime:  now,
			},
			{
				ID:         "docker-host-offline",
				ResourceID: "docker:docker-host-1",
				Type:       "offline",
				Level:      "critical",
				Message:    "Docker host offline",
				Value:      0,
				Threshold:  1,
				StartTime:  now,
			},
		},
	}

	// Populate the store
	store.PopulateFromSnapshot(snapshot)

	// Verify resources were created
	all := store.GetAll()
	if len(all) == 0 {
		t.Fatal("Expected resources to be populated, got 0")
	}

	t.Logf("Populated %d resources", len(all))

	// Check for each type
	nodes := store.Query().OfType(ResourceTypeNode).Execute()
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	} else {
		t.Logf("Node: id=%s, name=%s, cpu=%.1f%%", nodes[0].ID, nodes[0].Name, nodes[0].CPU.Current)
	}

	vms := store.Query().OfType(ResourceTypeVM).Execute()
	if len(vms) != 1 {
		t.Errorf("Expected 1 VM, got %d", len(vms))
	} else {
		t.Logf("VM: id=%s, name=%s, status=%s", vms[0].ID, vms[0].Name, vms[0].Status)
	}

	containers := store.Query().OfType(ResourceTypeContainer).Execute()
	if len(containers) != 1 {
		t.Errorf("Expected 1 container, got %d", len(containers))
	} else {
		t.Logf("Container: id=%s, name=%s", containers[0].ID, containers[0].Name)
	}

	hosts := store.Query().OfType(ResourceTypeHost).Execute()
	if len(hosts) != 1 {
		t.Errorf("Expected 1 host, got %d", len(hosts))
	} else {
		t.Logf("Host: id=%s, hostname=%s", hosts[0].ID, hosts[0].Identity.Hostname)
		if len(hosts[0].Alerts) != 2 {
			t.Errorf("expected host alerts to include host and disk alerts, got %d", len(hosts[0].Alerts))
		}
	}

	dockerHosts := store.Query().OfType(ResourceTypeDockerHost).Execute()
	if len(dockerHosts) != 1 {
		t.Errorf("Expected 1 docker host, got %d", len(dockerHosts))
	} else if len(dockerHosts[0].Alerts) != 1 {
		t.Errorf("expected docker host alert to attach, got %d", len(dockerHosts[0].Alerts))
	}

	dockerContainers := store.Query().OfType(ResourceTypeDockerContainer).Execute()
	if len(dockerContainers) != 1 {
		t.Errorf("Expected 1 docker container, got %d", len(dockerContainers))
	} else if len(dockerContainers[0].Alerts) != 1 {
		t.Errorf("expected docker container alert to attach, got %d", len(dockerContainers[0].Alerts))
	}

	stats := store.GetStats()
	if stats.WithAlerts != 4 {
		t.Errorf("expected 4 resources with alerts, got %d", stats.WithAlerts)
	}

	// Test summary
	t.Logf("SUCCESS: PopulateFromSnapshot works correctly!")
	t.Logf("Total resources: %d", len(all))
}

// TestPopulateFromSnapshotRemovesStaleResources verifies that resources not present
// in subsequent snapshots are removed from the store (fixing the "ghost LXCs" bug).
func TestPopulateFromSnapshotRemovesStaleResources(t *testing.T) {
	store := NewStore()

	// First snapshot with 2 containers
	snapshot1 := models.StateSnapshot{
		Containers: []models.Container{
			{
				ID:       "ct-100",
				VMID:     100,
				Name:     "container-to-keep",
				Node:     "pve-node-1",
				Instance: "pve1",
				Status:   "running",
			},
			{
				ID:       "ct-200",
				VMID:     200,
				Name:     "container-to-remove",
				Node:     "pve-node-1",
				Instance: "pve1",
				Status:   "running",
			},
		},
	}

	// Populate with first snapshot
	store.PopulateFromSnapshot(snapshot1)

	// Verify both containers exist
	containers := store.Query().OfType(ResourceTypeContainer).Execute()
	if len(containers) != 2 {
		t.Fatalf("Expected 2 containers after first snapshot, got %d", len(containers))
	}
	t.Logf("After first snapshot: %d containers", len(containers))

	// Second snapshot with only 1 container (the other was "deleted" from Proxmox)
	snapshot2 := models.StateSnapshot{
		Containers: []models.Container{
			{
				ID:       "ct-100",
				VMID:     100,
				Name:     "container-to-keep",
				Node:     "pve-node-1",
				Instance: "pve1",
				Status:   "running",
			},
			// container-to-remove is NOT included - simulating it was deleted
		},
	}

	// Populate with second snapshot
	store.PopulateFromSnapshot(snapshot2)

	// Verify only 1 container remains
	containers = store.Query().OfType(ResourceTypeContainer).Execute()
	if len(containers) != 1 {
		t.Fatalf("Expected 1 container after second snapshot (removed container should be gone), got %d", len(containers))
	}

	// Verify the correct container remains
	if containers[0].ID != "ct-100" {
		t.Errorf("Expected container-to-keep (ct-100) to remain, got %s", containers[0].ID)
	}

	// Verify the removed container is gone
	_, found := store.Get("ct-200")
	if found {
		t.Error("container-to-remove (ct-200) should have been removed from the store")
	}

	t.Logf("SUCCESS: Removed resources are correctly cleaned up!")
	t.Logf("After second snapshot: %d container(s) - 'container-to-remove' was properly removed", len(containers))
}
