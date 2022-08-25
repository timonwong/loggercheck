package logrlint

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoredLoggerFlag(t *testing.T) {
	f := ignoredLoggersFlag{}

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
	assert.ErrorContains(t, err, "-ignoredloggers: unknown logger: unknownlogger")
}
