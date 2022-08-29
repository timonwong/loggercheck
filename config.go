package loggercheck

import (
	"github.com/timonwong/loggercheck/rules"
	"github.com/timonwong/loggercheck/sets"
)

type Config struct {
	Disable []string
	Rules   []string
}

func (c *Config) init(l *loggercheck) {
	if c == nil {
		return
	}
	// Init configs from external API call (golangci-lint for example).
	l.disable.StringSet = sets.NewStringSet(c.Disable...)
	l.ruleFile.filename = "<internal>"
	ruleset, err := rules.ParseRules(c.Rules)
	if err != nil {
		return
	}
	l.ruleFile.rulsetList = ruleset
}
