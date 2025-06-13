package node

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"humpback-agent/types"
)

func (n *Node) CpuInfo() (*types.CpuInfo, error) {
	physicsCount, err := cpu.Counts(false)
	if err != nil {
		return nil, err
	}
	logicalCount, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}
	cpuInfo := &types.CpuInfo{
		PhysicsCount: physicsCount,
		LogicalCount: logicalCount,
	}
	percents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	if len(percents) > 0 {
		cpuInfo.Percent = percents[0]
	}
	cpus, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	for _, cpu := range cpus {
		cpuInfo.Names = append(cpuInfo.Names, cpu.ModelName)
	}

	return cpuInfo, nil
}
