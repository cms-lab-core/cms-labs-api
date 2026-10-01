package tasks

import "errors"

type StartupFiberUC struct {
	UserDefaultCreateUC *UserDefaultCreateUC
	ProxmoxSyncUC       *ProxmoxSyncUC
	LTISyncResultUC     *LTISyncResultUC
	DemoSeedUC          *DemoSeedUC
}

func (u *StartupFiberUC) Startup(taskName *string, taskArgs ...string) (bool, error) {
	// Запускаем всегда UserDefaultCreateUC
	_ = u.UserDefaultCreateUC.Execute()
	if taskName == nil || *taskName == "" {
		return false, nil
	}
	if *taskName == "proxmox_sync" {
		return true, u.ProxmoxSyncUC.Execute()
	}
	if *taskName == "grade_sync" {
		return true, u.LTISyncResultUC.Execute()
	}
	if *taskName == "demo_seed" {
		catalogPath := ""
		if len(taskArgs) > 0 {
			catalogPath = taskArgs[0]
		}
		return true, u.DemoSeedUC.Execute(catalogPath)
	}
	return false, errors.New("task is undefined")
}
