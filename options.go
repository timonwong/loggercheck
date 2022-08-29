package loggercheck

import (
	"github.com/timonwong/loggercheck/rules"
	"github.com/timonwong/loggercheck/sets"
)

type Option func(*loggercheck) error

func WithDisable(disable []string) Option {
	return func(l *loggercheck) error {
		l.disable.StringSet = sets.NewStringSet(disable...)
		return nil
	}
}

func WithRules(customRules []string) Option {
	return func(l *loggercheck) error {
		l.ruleFile.filename = "<internal>"
		ruleset, err := rules.ParseRules(customRules)
		if err != nil {
			return err
		}
		l.ruleFile.rulsetList = ruleset
		return nil
	}
}
