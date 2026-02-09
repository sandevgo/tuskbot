package installer

type InstallState struct {
	EnvVars map[string]string
}

func NewInstallState() *InstallState {
	return &InstallState{
		EnvVars: make(map[string]string),
	}
}
