package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFuncRule(t *testing.T) {
	testCases := []struct {
		name              string
		rule              string
		wantError         error
		wantPackageImport string
		wantRule          MethodRule
	}{
		{
			name:      "invalid-rule-missing-paren",
			rule:      "(*go.uber.org/zap/SugaredLogger.Debugw",
			wantError: ErrInvalidRule,
		},
		{
			name:      "invalid-rule-receiver-no-type",
			rule:      "(*go.uber.org/zap/SugaredLogger).Debugw",
			wantError: ErrInvalidRule,
		},
		{
			name:      "invalid-rule-just-import",
			rule:      "go.uber.org/zap",
			wantError: ErrInvalidRule,
		},
		{
			name:              "zap",
			rule:              "(*go.uber.org/zap.SugaredLogger).Debugw",
			wantPackageImport: "go.uber.org/zap",
			wantRule: MethodRule{
				IsReceiver:   true,
				ReceiverType: "*SugaredLogger",
				MethodName:   "Debugw",
			},
		},
		{
			name:              "klog-no-receiver",
			rule:              "k8s.io/klog/v2.InfoS",
			wantPackageImport: "k8s.io/klog/v2",
			wantRule: MethodRule{
				MethodName: "InfoS",
			},
		},
		{
			name:              "logr",
			rule:              "(github.com/go-logr/logr.Logger).Error",
			wantPackageImport: "github.com/go-logr/logr",
			wantRule: MethodRule{
				IsReceiver:   true,
				ReceiverType: "Logger",
				MethodName:   "Error",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotPackageImport, gotRule, err := ParseMethodRule(tc.rule)
			if tc.wantError != nil {
				assert.EqualError(t, err, tc.wantError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantPackageImport, gotPackageImport)
				assert.Equal(t, tc.wantRule, gotRule)
			}
		})
	}
}
