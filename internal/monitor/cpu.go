package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

// CPUMonitor 用于监控CPU使用率
type CPUMonitor struct {
	// 更新间隔
	Interval time.Duration
	// CPU使用率变化时的回调函数
	OnUpdate func(usage float64)
	// 停止监控的通道
	stopCh chan struct{}
	// 是否正在运行
	running bool
}

// NewCPUMonitor 创建一个新的CPU监控器
func NewCPUMonitor(interval time.Duration) *CPUMonitor {
	if interval < time.Second {
		interval = time.Second
	}
	return &CPUMonitor{
		Interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 开始监控CPU使用率
func (m *CPUMonitor) Start() {
	if m.running || m.OnUpdate == nil {
		return
	}

	m.running = true
	go func() {
		ticker := time.NewTicker(m.Interval)
		defer ticker.Stop()

		// 第一次读取CPU使用率（丢弃，因为第一次读取通常不准确）
		_, _ = cpu.Percent(0, false)

		for {
			select {
			case <-ticker.C:
				percentages, err := cpu.Percent(0, false)
				if err != nil || len(percentages) == 0 {
					continue
				}
				// 获取总体CPU使用率
				usage := percentages[0]

				// 确保使用率在0-100之间
				if usage < 0 {
					usage = 0
				} else if usage > 100 {
					usage = 100
				}

				// 调用回调函数
				m.OnUpdate(usage)

			case <-m.stopCh:
				m.running = false
				return
			}
		}
	}()
}

// Stop 停止监控CPU使用率
func (m *CPUMonitor) Stop() {
	if !m.running {
		return
	}
	m.stopCh <- struct{}{}
}

// IsRunning 检查监控器是否正在运行
func (m *CPUMonitor) IsRunning() bool {
	return m.running
}
