package metrics

import (
	"cluster-scheduler/proto"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func CollectStats(nodeID string) (*proto.Heartbeat, error) {
	// cpu.Percent gives us the usage percentage
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	// cpu.Counts(true) gives us the number of logical cores
	totalCores, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	const mbFactor = 1024 * 1024

	// Simplified available cores calculation (can be improved by tracking actual jobs)
	availableCores := int(float64(totalCores) * (1.0 - (cpuPercents[0] / 100.0)))
	if availableCores < 0 {
		availableCores = 0
	}

	hb := &proto.Heartbeat{
		NodeID:         nodeID,
		CPUPercent:     cpuPercents[0],
		MemoryPercent:  vMem.UsedPercent,
		TotalCores:     totalCores,
		AvailableCores: availableCores,
		FreeMemoryMB:   vMem.Available / mbFactor,
		TotalMemoryMB:  vMem.Total / mbFactor,
		RunningJobs:    []string{}, // Agent fills this from its JobMap
	}

	return hb, nil
}
