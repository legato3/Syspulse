package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/rcourtman/pulse-go-rewrite/pkg/proxmox"
)

type slowGuestAgentClusterClient struct {
	stubPVEClient
	resources []proxmox.ClusterResource
	fsDelay   time.Duration
}

func (c *slowGuestAgentClusterClient) GetClusterResources(ctx context.Context, resourceType string) ([]proxmox.ClusterResource, error) {
	return c.resources, nil
}

func (c *slowGuestAgentClusterClient) GetVMStatus(ctx context.Context, node string, vmid int) (*proxmox.VMStatus, error) {
	return &proxmox.VMStatus{
		MaxMem: 8 * 1024,
		Mem:    4 * 1024,
		Agent:  proxmox.VMAgentField{Value: 1},
	}, nil
}

func (c *slowGuestAgentClusterClient) GetVMFSInfo(ctx context.Context, node string, vmid int) ([]proxmox.VMFileSystem, error) {
	select {
	case <-time.After(c.fsDelay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return []proxmox.VMFileSystem{{
		Mountpoint: "/",
		Type:       "ext4",
		TotalBytes: 100 * 1024 * 1024 * 1024,
		UsedBytes:  40 * 1024 * 1024 * 1024,
		Disk:       "/dev/vda",
	}}, nil
}

func TestGuestAgentFSInfoBudgetHonorsConfiguredTimeouts(t *testing.T) {
	t.Parallel()

	m := &Monitor{
		guestAgentFSInfoTimeout: 15 * time.Second,
		guestAgentRetries:       1,
	}

	budget := m.guestAgentFSInfoBudget()
	if budget < 30*time.Second {
		t.Fatalf("guestAgentFSInfoBudget() = %s, want at least 30s", budget)
	}
}

func TestPollVMsAndContainersEfficientCompletesDiskQueriesWithinPollBudget(t *testing.T) {
	t.Setenv("PULSE_DATA_DIR", t.TempDir())

	client := &slowGuestAgentClusterClient{
		fsDelay: 60 * time.Millisecond,
		resources: []proxmox.ClusterResource{
			{Type: "qemu", Node: "node1", VMID: 100, Name: "vm100", Status: "running", MaxMem: 8 * 1024, Mem: 4 * 1024, MaxDisk: 100 * 1024 * 1024 * 1024, MaxCPU: 4},
			{Type: "qemu", Node: "node1", VMID: 101, Name: "vm101", Status: "running", MaxMem: 8 * 1024, Mem: 4 * 1024, MaxDisk: 100 * 1024 * 1024 * 1024, MaxCPU: 4},
			{Type: "qemu", Node: "node1", VMID: 102, Name: "vm102", Status: "running", MaxMem: 8 * 1024, Mem: 4 * 1024, MaxDisk: 100 * 1024 * 1024 * 1024, MaxCPU: 4},
			{Type: "qemu", Node: "node1", VMID: 103, Name: "vm103", Status: "running", MaxMem: 8 * 1024, Mem: 4 * 1024, MaxDisk: 100 * 1024 * 1024 * 1024, MaxCPU: 4},
		},
	}

	mon := newTestPVEMonitor("pve1")
	defer mon.alertManager.Stop()
	defer mon.notificationMgr.Stop()

	mon.rateTracker = NewRateTracker()
	mon.guestMetadataCache = make(map[string]guestMetadataCacheEntry)
	mon.guestMetadataLimiter = make(map[string]time.Time)
	mon.vmRRDMemCache = make(map[string]rrdMemCacheEntry)
	mon.vmAgentMemCache = make(map[string]agentMemCacheEntry)
	mon.guestAgentFSInfoTimeout = 250 * time.Millisecond
	mon.guestAgentNetworkTimeout = 250 * time.Millisecond
	mon.guestAgentOSInfoTimeout = 250 * time.Millisecond
	mon.guestAgentVersionTimeout = 250 * time.Millisecond
	mon.guestAgentRetries = 0
	mon.guestAgentWorkSlots = make(chan struct{}, 4)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()

	if ok := mon.pollVMsAndContainersEfficient(ctx, "pve1", "", false, client, map[string]string{"node1": "online"}); !ok {
		t.Fatal("pollVMsAndContainersEfficient() returned false")
	}

	state := mon.state.GetSnapshot()
	if len(state.VMs) != 4 {
		t.Fatalf("expected 4 VMs, got %d", len(state.VMs))
	}
	for _, vm := range state.VMs {
		if vm.Disk.Total <= 0 || vm.Disk.Usage <= 0 {
			t.Fatalf("expected guest-agent disk data for %s, got total=%d usage=%.2f", vm.Name, vm.Disk.Total, vm.Disk.Usage)
		}
	}
}
