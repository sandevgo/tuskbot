package service

type Config struct {
	Name             string
	DisplayName      string
	Description      string
	Executable       string
	Arguments        []string
	WorkingDirectory string
	UserService      bool
}
