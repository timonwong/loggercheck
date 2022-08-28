package loggercheck

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoredLoggerFlag(t *testing.T) {
	f := loggerCheckersFlag{}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&f, "ignoredloggers", "")

	var err error

	err = fs.Parse([]string{"-ignoredloggers=klog"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"klog"}, f.List())

	err = fs.Parse([]string{"-ignoredloggers=logr,klog"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"klog", "logr"}, f.List())

	err = fs.Parse([]string{"-ignoredloggers=logr,klog,unknownlogger"})
	assert.ErrorContains(t, err, "-ignoredloggers: unknown logger: \"unknownlogger\"")
}

func TestWrongRuleFile(t *testing.T) {
	f := ruleFileFlag{}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&f, "rulefile", "")

	err := fs.Parse([]string{"-rulefile=testdata/xxx-not-exists-xxx.yaml"})
	assert.ErrorContains(t, err, "open testdata/xxx-not-exists-xxx.yaml: no such file or directory")
}
