package loggercheck

import (
	"github.com/timonwong/loggercheck/pattern"
	"github.com/timonwong/loggercheck/sets"
)

type Config struct {
	Disable       sets.StringSet
	PatternGroups pattern.GroupList
}

func (c *Config) init(l *loggercheck) {
	if c == nil {
		return
	}

	// Init configs from external API call (golangci-lint for example).
	l.disable.StringSet = c.Disable
	l.patternFile.filename = "<internal>"
	l.patternFile.patternGroups = c.PatternGroups
}
