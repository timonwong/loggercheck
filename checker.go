package loggercheck

import (
	"fmt"

	"github.com/timonwong/loggercheck/pattern"
)

var staticPatternGroups = pattern.GroupList{
	mustNewPatternGroup("logr", []string{
		"(github.com/go-logr/logr.Logger).Error",
		"(github.com/go-logr/logr.Logger).Info",
		"(github.com/go-logr/logr.Logger).WithValues",
	}),
	mustNewPatternGroup("klog", []string{
		"k8s.io/klog/v2.InfoS",
		"k8s.io/klog/v2.InfoSDepth",
		"k8s.io/klog/v2.ErrorS",
		"(k8s.io/klog/v2.Verbose).InfoS",
		"(k8s.io/klog/v2.Verbose).InfoSDepth",
		"(k8s.io/klog/v2.Verbose).ErrorS",
	}),
	mustNewPatternGroup("zap", []string{
		"(*go.uber.org/zap.SugaredLogger).With",
		"(*go.uber.org/zap.SugaredLogger).Debugw",
		"(*go.uber.org/zap.SugaredLogger).Infow",
		"(*go.uber.org/zap.SugaredLogger).Warnw",
		"(*go.uber.org/zap.SugaredLogger).Errorw",
		"(*go.uber.org/zap.SugaredLogger).DPanicw",
		"(*go.uber.org/zap.SugaredLogger).Panicw",
		"(*go.uber.org/zap.SugaredLogger).Fatalw",
	}),
}

// mustNewPatternGroup only called at init, catch errors during development.
// In production it will not panic.
func mustNewPatternGroup(name string, patternLines []string) pattern.Group {
	if len(patternLines) == 0 {
		panic("empty pattern lines")
	}

	var packageImport string
	patterns := make([]pattern.Pattern, 0, len(patternLines))
	for _, s := range patternLines {
		pat, err := pattern.ParseRule(s)
		if err != nil {
			panic(err)
		}
		patterns = append(patterns, pat)
		if packageImport == "" {
			packageImport = pat.PackageImport
		} else if packageImport != pat.PackageImport {
			panic(fmt.Sprintf("package import mismatch: %s != %s", packageImport, pat.PackageImport))
		}
	}

	return pattern.Group{
		Name:          name,
		PackageImport: packageImport,
		Patterns:      patterns,
	}
}
