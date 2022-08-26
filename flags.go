package loggercheck

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type loggerCheckersFlag struct {
	stringSet
}

// Set implements flag.Value interface.
func (f *loggerCheckersFlag) Set(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		f.stringSet = nil
		return nil
	}

	parts := strings.Split(s, ",")
	set := newStringSet(parts...)
	err := validateIgnoredLoggerFlag(set)
	if err != nil {
		return err
	}
	f.stringSet = set
	return nil
}

// String implements flag.Value interface
func (f *loggerCheckersFlag) String() string {
	return strings.Join(f.List(), ",")
}

func validateIgnoredLoggerFlag(set stringSet) error {
	for key := range set {
		if !loggerCheckersByName.HasKey(key) {
			return fmt.Errorf("unknown logger: %q", key)
		}
	}

	return nil
}

type configFlag struct {
	cfg *Config
}

// Set implements flag.Value interface.
func (f *configFlag) Set(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	if s == "sample" {
		f.dumpSampleConfig()
		return nil
	}

	content, err := os.ReadFile(s)
	if err != nil {
		return fmt.Errorf("read cfg file %s failed, err=%w", s, err)
	}

	f.cfg = &Config{}
	if err := yaml.Unmarshal(content, f.cfg); err != nil {
		f.cfg = nil
		return fmt.Errorf("load cfg file %s failed, err=%w", s, err)
	}

	// add custom loggers
	for _, ck := range f.cfg.CustomCheckers {
		addLogger(ck.Name, ck.PackageImport, ck.Funcs)
	}

	return nil
}

func (f *configFlag) dumpSampleConfig() {
	cfg := &Config{
		Disable: []string{"klog", "logr", "zap"},
		CustomCheckers: []Checker{
			{
				Name:          "mylogger",
				PackageImport: "example.com/mylogger",
				Funcs: []string{
					"(*example.com/mylogger.Logger).Debugw",
					"(*example.com/mylogger.Logger).Infow",
					"(*example.com/mylogger.Logger).Warnw",
					"(*example.com/mylogger.Logger).Errorw",
					"(*example.com/mylogger.Logger).With",
				},
			},
		},
	}
	out, _ := yaml.Marshal(cfg)
	fmt.Println("# loggercheck sample config")
	fmt.Println(string(out))
}

// String implements flag.Value interface
func (f *configFlag) String() string {
	out, _ := yaml.Marshal(f.cfg)
	return string(out)
}
