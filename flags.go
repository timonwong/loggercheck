package loggercheck

import (
	"fmt"
	"os"
	"strings"

	"github.com/timonwong/loggercheck/pattern"
	"github.com/timonwong/loggercheck/sets"
)

type loggerCheckersFlag struct {
	sets.StringSet
}

// Set implements flag.Value interface.
func (f *loggerCheckersFlag) Set(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		f.StringSet = nil
		return nil
	}

	parts := strings.Split(s, ",")
	set := sets.NewStringSet(parts...)
	err := validateIgnoredLoggerFlag(set)
	if err != nil {
		return err
	}
	f.StringSet = set
	return nil
}

// String implements flag.Value interface
func (f *loggerCheckersFlag) String() string {
	return strings.Join(f.List(), ",")
}

func validateIgnoredLoggerFlag(set sets.StringSet) error {
	for key := range set {
		if !staticPatternGroups.HasName(key) {
			return fmt.Errorf("unknown logger: %q", key)
		}
	}

	return nil
}

type patternFileFlag struct {
	filename      string
	patternGroups pattern.GroupList
}

// Set implements flag.Value interface.
func (f *patternFileFlag) Set(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	pgList, err := pattern.ParseRuleFile(r)
	if err != nil {
		return err
	}

	f.filename = filename
	f.patternGroups = pgList
	return nil
}

// String implements flag.Value interface
func (f *patternFileFlag) String() string {
	return f.filename
}
