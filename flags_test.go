package loggercheck

import (
	"flag"
	"io"
	"os"
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

func TestConfigFlagDumpSampleConfig(t *testing.T) {
	f := configFlag{}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&f, "config", "")

	err := fs.Parse([]string{"-config=sample"})
	assert.NoError(t, err)
}

func TestConfigFlagLoadFail(t *testing.T) {
	f := configFlag{}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&f, "config", "")

	err := fs.Parse([]string{"-config=/tmp/absolute-not-exists-config.yaml"})
	assert.ErrorContains(t, err, "read cfg file /tmp/absolute-not-exists-config.yaml failed")
}

func TestConfigFlagLoadConfig(t *testing.T) {
	l := &loggercheck{}
	l.config = configFlag{l: l}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&l.config, "config", "")

	fs.Var(&l.disable, "disable", "")

	testConfig := []byte(`# loggercheck sample config
disable:
    - klog
    - logr
custom-checkers: []
`)
	configFile, err := os.CreateTemp("/tmp", "loggercheck-test-cfg-")
	if err != nil {
		t.Fatal(err)
	}

	if _, errWrite := configFile.Write(testConfig); errWrite != nil {
		t.Errorf("write test config file failed: %v", errWrite)
	}
	configFile.Close()

	testConfigFile := configFile.Name()
	t.Cleanup(func() {
		os.Remove(testConfigFile)
	})

	err = fs.Parse([]string{"-config=" + testConfigFile})
	assert.NoError(t, err)
	assert.Equal(t, []string{"klog", "logr"}, l.disable.List())

	// config file should not override `-disable` flag value
	err = fs.Parse([]string{"-config=" + testConfigFile, "-disable=klog"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"klog"}, l.disable.List())

	// config file should not override `-disable` flag value
	err = fs.Parse([]string{"-disable=logr", "-config=" + testConfigFile})
	assert.NoError(t, err)
	assert.Equal(t, []string{"logr"}, l.disable.List())
}
