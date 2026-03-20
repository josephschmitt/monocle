package adapters

// InstallResult reports what happened during install/uninstall.
type InstallResult struct {
	Agent       string
	Installed   bool
	AlreadyDone bool
	Err         error
	Details     []string // additional installation info lines
}
