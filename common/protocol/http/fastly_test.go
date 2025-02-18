package http_test

import (
	"testing"

	. "github.com/xtls/xray-core/common/protocol/http"
)

func TestIsIPInFastlyRange(t *testing.T) {
	testCases := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "Valid Fastly IPv4",
			ip:       "151.101.1.1",
			expected: true,
		},
		{
			name:     "Valid Fastly IPv4 2",
			ip:       "199.232.0.1",
			expected: true,
		},
		{
			name:     "Valid Fastly IPv6",
			ip:       "2a04:4e40::1",
			expected: true,
		},
		{
			name:     "Non-Fastly IPv4",
			ip:       "192.168.1.1",
			expected: false,
		},
		{
			name:     "Non-Fastly IPv6",
			ip:       "2001:db8::1",
			expected: false,
		},
		{
			name:     "Invalid IP",
			ip:       "not-an-ip",
			expected: false,
		},
		{
			name:     "Empty IP",
			ip:       "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsIPInFastlyRange(tc.ip)
			if result != tc.expected {
				t.Errorf("IsIPInFastlyRange(%v) = %v, want %v", tc.ip, result, tc.expected)
			}
		})
	}
}
