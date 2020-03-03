// +build !linux

package executor

import (
	"github.com/actiontech/dtle/plugins/drivers"
	hclog "github.com/hashicorp/go-hclog"
)

func NewExecutorWithIsolation(logger hclog.Logger) Executor {
	logger = logger.Named("executor")
	logger.Error("isolation executor is not supported on this platform, using default")
	return NewExecutor(logger)
}

func (e *UniversalExecutor) configureResourceContainer(_ int) error { return nil }

func (e *UniversalExecutor) runAs(_ string) error { return nil }

func (e *UniversalExecutor) getAllPids() (map[int]*nomadPid, error) {
	return getAllPidsByScanning()
}

func (e *UniversalExecutor) start(command *ExecCommand) error {
	return e.childCmd.Start()
}

func withNetworkIsolation(f func() error, _ *drivers.NetworkIsolationSpec) error {
	return f()
}
