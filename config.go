package loggercheck

type Config struct {
	Disable        []string  `json:"disable"`
	CustomCheckers []Checker `yaml:"custom-checkers"`
}

type Checker struct {
	Name          string   `yaml:"name"`
	PackageImport string   `yaml:"package-import"`
	Funcs         []string `yaml:"funcs"`
}
