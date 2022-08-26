package logrlint

import "sort"

type loggerChecker struct {
	packageImport string
	funcNames     stringSet
}

var loggerCheckersByName = loggerCheckerMap{
	"logr": {
		packageImport: "github.com/go-logr/logr",
		funcNames:     newStringSet("Error", "Info", "WithValues"),
	},
	"klog": {
		packageImport: "k8s.io/klog/v2",
		funcNames:     newStringSet("InfoS", "InfoSDepth", "ErrorS"),
	},
}

type loggerCheckerMap map[string]loggerChecker

func (m loggerCheckerMap) Names() []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
