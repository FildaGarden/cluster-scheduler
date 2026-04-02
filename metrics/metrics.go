package metrics

import (
	"cluster-scheduler/proto"
	"net"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Zjistí síťovou adresu uzlu pomocí UDP
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80") // Google DNS - nejspis bude vhodny pozdeji zmenit na IP Mastera
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// Sber aktualnich metrik Agenta
func CollectStats(nodeID string) (*proto.Heartbeat, error) {
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	totalCores, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	const mbFactor = 1024 * 1024

	// Provizorni vypocet dostupnych CPU jader - might change later
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
		RunningJobs:    []string{}, // Vyplni Agent podle JobsMap
	}

	return hb, nil
}
