package node

import (
	"github.com/shirou/gopsutil/v4/mem"
	"humpback-agent/types"
)

func (n *Node) MemoryInfo() (*types.MemoryInfo, error) {
	virtualMemoryStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &types.MemoryInfo{
		Total:   virtualMemoryStat.Total,
		Used:    virtualMemoryStat.Used,
		Percent: virtualMemoryStat.UsedPercent,
	}, nil
}
