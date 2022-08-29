package loggercheck

import (
	"github.com/timonwong/loggercheck/rules"
	"github.com/timonwong/loggercheck/sets"
)

type Config struct {
	Disable     sets.StringSet
	RulesetList rules.RulesetList
}

func (c *Config) init(l *loggercheck) {
	if c == nil {
		return
	}

	// Init configs from external API call (golangci-lint for example).
	l.disable.StringSet = c.Disable
	l.ruleFile.filename = "<internal>"
	l.ruleFile.rulsetList = c.RulesetList
}
