package interfaces

import (
	"github.com/actiontech/dtle/nomad/structs"
)

// RunnnerHook is a lifecycle hook into the life cycle of an allocation runner.
type RunnerHook interface {
	Name() string
}

// RunnerPrerunHooks are executed before calling TaskRunner.Run for
// non-terminal allocations. Terminal allocations do *not* call prerun.
type RunnerPrerunHook interface {
	RunnerHook
	Prerun() error
}

// RunnerPostrunHooks are executed after calling TaskRunner.Run, even for
// terminal allocations. Therefore Postrun hooks must be safe to call without
// first calling Prerun hooks.
type RunnerPostrunHook interface {
	RunnerHook
	Postrun() error
}

// RunnerDestroyHooks are executed after AllocRunner.Run has exited and must
// make a best effort cleanup allocation resources. Destroy hooks must be safe
// to call without first calling Prerun.
type RunnerDestroyHook interface {
	RunnerHook
	Destroy() error
}

// RunnerUpdateHooks are executed when an allocation update is received from
// the server. Update is called concurrently with AllocRunner execution and
// therefore must be safe for concurrent access with other hook methods. Calls
// to Update are serialized so allocaiton updates will always be processed in
// order.
type RunnerUpdateHook interface {
	RunnerHook
	Update(*RunnerUpdateRequest) error
}

type RunnerUpdateRequest struct {
	Alloc *structs.Allocation
}

// ShutdownHook may be implemented by AllocRunner or TaskRunner hooks and will
// be called when the agent process is being shutdown gracefully.
type ShutdownHook interface {
	RunnerHook

	Shutdown()
}
