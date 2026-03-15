package migration

type Migration func(runtimePath string) error

var Migrations []Migration = []Migration{
	0: func(p string) error {
		// create subdirectories if not exists [prompt, config, data]
		return nil
	},
	1: func(p string) error {
		return nil
	},
}

func Run(runtimePath string) error {
	return nil
}
