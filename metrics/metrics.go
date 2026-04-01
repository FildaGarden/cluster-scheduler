package metrics

import (
	"cluster-scheduler/proto"
	"net"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GetLocalIP zjistí síťovou adresu uzlu pomocí UDP spojení
func GetLocalIP() string {
	// Připoj se na 8.8.8.8 — Google DNS
	// Neposílá žádná data, jen zjistí jaké rozhraní by použil
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	// Zjisti lokální adresu tohoto spojení
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

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
