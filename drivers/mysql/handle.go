package mysql

import (
	"context"
	"fmt"
	"github.com/actiontech/dtle/drivers/mysql/common"
	"github.com/actiontech/dtle/drivers/mysql/config"
	"github.com/actiontech/dtle/drivers/mysql/kafka"
	"github.com/actiontech/dtle/drivers/mysql/mysql"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

type taskHandle struct {
	logger hclog.Logger

	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	taskConfig  *drivers.TaskConfig
	procState   drivers.TaskState
	startedAt   time.Time
	completedAt time.Time
	exitResult  *drivers.ExitResult

	runner DriverHandle

	ctx context.Context
	cancelFunc context.CancelFunc
	waitCh chan *drivers.ExitResult
}

func newDtleTaskHandle(logger hclog.Logger, cfg *drivers.TaskConfig, state drivers.TaskState, started time.Time) *taskHandle {
	h := &taskHandle{
		logger:      logger,
		stateLock:   sync.RWMutex{},
		taskConfig:  cfg,
		procState:   state,
		startedAt:   started,
		completedAt: time.Time{},
		exitResult:  nil,
		waitCh:      make(chan *drivers.ExitResult, 1),
	}
	h.ctx, h.cancelFunc = context.WithCancel(context.TODO())
	return h
}

func (h *taskHandle) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:          h.taskConfig.ID,
		Name:        h.taskConfig.Name,
		State:       h.procState,
		StartedAt:   h.startedAt,
		CompletedAt: h.completedAt,
		ExitResult:  h.exitResult,
	}
}

func (h *taskHandle) IsRunning() bool {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()
	return h.procState == drivers.TaskStateRunning
}

func (h *taskHandle) run(taskConfig *config.DtleTaskConfig, d *Driver) {
	var err error
	h.stateLock.Lock()
	if h.exitResult == nil {
		h.exitResult = &drivers.ExitResult{}
	}
	h.procState = drivers.TaskStateRunning
	h.stateLock.Unlock()

	// TODO: detect if the taskConfig OOMed

	cfg := h.taskConfig
	ctx := &common.ExecContext{
		Subject:    cfg.JobName,
		StateDir:   d.config.DataDir,
	}

	taskConfig.SetDefaultForEmpty()
	driverConfig := &config.MySQLDriverConfig{DtleTaskConfig: *taskConfig}


	switch strings.ToLower(cfg.TaskGroupName) {
	case "src", "source":
		h.runner, err = mysql.NewExtractor(ctx, driverConfig, d.logger, d.storeManager, h.waitCh)
		if err != nil {
			h.exitResult.Err = errors.Wrap(err, "NewExtractor")
			return
		}
		go h.runner.Run()
	case "dst", "dest", "destination":
		if taskConfig.KafkaConfig != nil {
			d.logger.Debug("found kafka", "KafkaConfig", taskConfig.KafkaConfig)
			h.runner = kafka.NewKafkaRunner(ctx, taskConfig.KafkaConfig, d.logger,
				d.storeManager, d.config.NatsAdvertise, h.waitCh)
			go h.runner.Run()
		} else {
			h.runner, err = mysql.NewApplier(ctx, driverConfig, d.logger, d.storeManager,
				d.config.NatsAdvertise, h.waitCh)
			if err != nil {
				h.exitResult.Err = errors.Wrap(err, "NewApplier")
				return
			}
			go h.runner.Run()
		}
	default:
		h.exitResult.Err = fmt.Errorf("unknown processor type: %+v", cfg.TaskGroupName)
		return
	}

}
func (h *taskHandle) Destroy() bool {
	h.stateLock.RLock()
	//driver.des
	h.cancelFunc()
	if h.runner != nil {
		err := h.runner.Shutdown()
		if err != nil {
			h.logger.Error("error in h.runner.Shutdown", "err", err)
		}
	}
	return h.procState == drivers.TaskStateExited
}

type DriverHandle interface {
	Run()

	// Shutdown is used to stop the task
	Shutdown() error

	// Stats returns aggregated stats of the driver
	Stats() (*common.TaskStatistics, error)
}
