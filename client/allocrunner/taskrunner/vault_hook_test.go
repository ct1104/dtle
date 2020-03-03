package taskrunner

import "github.com/actiontech/dtle/client/allocrunner/interfaces"

// Statically assert the stats hook implements the expected interfaces
var _ interfaces.TaskPrestartHook = (*vaultHook)(nil)
var _ interfaces.TaskStopHook = (*vaultHook)(nil)
var _ interfaces.ShutdownHook = (*vaultHook)(nil)
