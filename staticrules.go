package loggercheck

import (
	"fmt"

	"github.com/timonwong/loggercheck/rules"
)

var staticRuleList = rules.RulesetList{
	mustNewStaticRuleSet("logr", []string{
		"(github.com/go-logr/logr.Logger).Error",
		"(github.com/go-logr/logr.Logger).Info",
		"(github.com/go-logr/logr.Logger).WithValues",
	}),
	mustNewStaticRuleSet("klog", []string{
		"k8s.io/klog/v2.InfoS",
		"k8s.io/klog/v2.InfoSDepth",
		"k8s.io/klog/v2.ErrorS",
		"(k8s.io/klog/v2.Verbose).InfoS",
		"(k8s.io/klog/v2.Verbose).InfoSDepth",
		"(k8s.io/klog/v2.Verbose).ErrorS",
	}),
	mustNewStaticRuleSet("zap", []string{
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

// mustNewStaticRuleSet only called at init, catch errors during development.
// In production it will not panic.
func mustNewStaticRuleSet(name string, lines []string) rules.Ruleset {
	if len(lines) == 0 {
		panic("empty rule lines")
	}

	rulesetList, err := rules.ParseRules(lines)
	if err != nil {
		panic(err)
	}

	if len(rulesetList) != 1 {
		panic(fmt.Errorf("expected 1 ruleset, got %d", len(rulesetList)))
	}

	ruleset := rulesetList[0]
	ruleset.Name = name
	return ruleset
}
