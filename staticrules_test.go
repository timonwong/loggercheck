package loggercheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mustNewStaticRuleSet_failCase(t *testing.T) {
	testCases := []struct {
		name      string
		rules     []string
		wantError string
	}{
		{
			name:      "nil",
			wantError: "no rules provided",
		},
		{
			name:      "empty",
			rules:     []string{},
			wantError: "no rules provided",
		},
		{
			name: "bad-rules",
			rules: []string{
				"# Comment",
				" ",
				"(*a/customonly.Logger).Debugw",
				"xxx",
			},
			wantError: "error parse rule at line 2: invalid rule format",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.PanicsWithError(t, tc.wantError, func() {
				mustNewStaticRuleSet("custom", tc.rules)
			})
		})
	}
}
