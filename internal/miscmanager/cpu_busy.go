package miscmanager

import (
	"fmt"
)

// CpuBusyContext def
type CpuBusyContext struct {
	config   *CpuBusyTestConfig
	counter  int
	doneChan chan bool
}

func (m *Manager) cpuBusyTest() {
	logger := m.logger
	config := m.config.CpuBusyTest

	logger.Info("Enter cpuBusyTest")
	defer logger.Info("Leave cpuBusyTest")

	cpuBusyCtx := &CpuBusyContext{
		config: config,
	}

	cpuBusyCtx.doneChan = make(chan bool)

	for i := 0; i < config.MaxWorker; i++ {
		workerID := fmt.Sprintf("worker_%d", i)
		go m.busyWork(workerID, cpuBusyCtx)
	}

	for i := 0; i < config.MaxWorker; i++ {
		<-m.doneChan
	}

	close(cpuBusyCtx.doneChan)
}

func (m Manager) busyWork(workID string, cpuBusyCtx *CpuBusyContext) {
	config := cpuBusyCtx.config
	logger := m.logger.Named(workID)
	logger.Info("Enter " + workID)
	defer logger.Info("Leave " + workID)

	for i := 0; i < config.Number; i++ {
		if cpuBusyCtx.counter < 10000 {
			cpuBusyCtx.counter += 1
		} else {
			cpuBusyCtx.counter -= 1
		}
	}

	cpuBusyCtx.doneChan <- true
}
