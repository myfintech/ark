package kube

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeNamespace(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"should normalize namespace": {
			input:    "SRE-2057/ark-ephemeral-environments",
			expected: "sre-2057-ark-ephemeral-environments",
		},
		"should truncate namespace to less than the max limit for k8s": {
			input:    "ff37cfba504b1562402b780fae00ea2697c1b27d36b4eb02b386db9e8d0c6d44",
			expected: "ff37cfba504b1562402b780fae00ea2697c1b27d36b4eb02b386db9e8d0c6d4",
		},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, NormalizeNamespace(test.input))
	}

}
