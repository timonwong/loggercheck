package loggercheck

import (
	"sync"
	"testing"
)

// The analyzer runs once per package, possibly concurrently. Custom rules must
// be appended to rulesetList only once, otherwise rulesetList keeps growing and
// a concurrent reader can index a stale slice with a newer index, which panics
// with "index out of range".
func TestLoadConfigOnce(t *testing.T) {
	l := newLoggerCheck(WithRules([]string{
		"(example.com/logger.Logger).Infow",
		"example.com/logger.Infow",
	}))

	const workers = 32
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.loadConfig(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	want := len(staticRuleList) + 1 // both rules share one package import
	if got := len(l.rulesetList); got != want {
		t.Fatalf("len(rulesetList) = %d, want %d", got, want)
	}
	if got := l.rulesetIndicesByImport["example.com/logger"]; len(got) != 1 {
		t.Fatalf("indices for example.com/logger = %v, want exactly one", got)
	}
}
