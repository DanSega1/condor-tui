// Package sysinfo collects lightweight system metrics for display in the header.
// All errors are silently swallowed — if a metric is unavailable the
// corresponding field is set to "–" so the UI never crashes.
package sysinfo

import (
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Stats holds a single snapshot of system metrics.
type Stats struct {
	CPU    string // e.g. "12%"  or "–"
	Memory string // e.g. "34%"  or "–"
	NetUp  string // e.g. "↑ 2k" or "–"
	NetDn  string // e.g. "↓ 8k" or "–"
}

var prevNetIO []net.IOCountersStat

// Collect gathers a fresh snapshot. Safe to call on every tick.
func Collect(wantCPU, wantMem, wantNet bool) Stats {
	s := Stats{CPU: "–", Memory: "–", NetUp: "–", NetDn: "–"}

	if wantCPU {
		if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
			s.CPU = fmt.Sprintf("%.0f%%", math.Round(pcts[0]))
		}
	}

	if wantMem {
		if vm, err := mem.VirtualMemory(); err == nil {
			s.Memory = fmt.Sprintf("%.0f%%", math.Round(vm.UsedPercent))
		}
	}

	if wantNet {
		curr, err := net.IOCounters(false)
		if err == nil && len(curr) > 0 {
			if prevNetIO != nil && len(prevNetIO) > 0 {
				up := curr[0].BytesSent - prevNetIO[0].BytesSent
				dn := curr[0].BytesRecv - prevNetIO[0].BytesRecv
				s.NetUp = "↑ " + formatBytes(up)
				s.NetDn = "↓ " + formatBytes(dn)
			}
			prevNetIO = curr
		}
	}

	return s
}

func formatBytes(b uint64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.0fk", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
