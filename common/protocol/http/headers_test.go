package http_test

import (
	"bufio"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/net"
	. "github.com/xtls/xray-core/common/protocol/http"
)

func TestParseXForwardedFor(t *testing.T) {
	header := http.Header{}
	header.Add("X-Forwarded-For", "129.78.138.66, 129.78.64.103")
	addrs := ParseXForwardedFor(header)
	if r := cmp.Diff(addrs, []net.Address{net.ParseAddress("129.78.138.66"), net.ParseAddress("129.78.64.103")}); r != "" {
		t.Error(r)
	}
}

func TestHopByHopHeadersRemoving(t *testing.T) {
	rawRequest := `GET /pkg/net/http/ HTTP/1.1
Host: golang.org
Connection: keep-alive,Foo, Bar
Foo: foo
Bar: bar
Proxy-Connection: keep-alive
Proxy-Authenticate: abc
Accept-Encoding: gzip
Accept-Charset: ISO-8859-1,UTF-8;q=0.7,*;q=0.7
Cache-Control: no-cache
Accept-Language: de,en;q=0.7,en-us;q=0.3

`
	b := bufio.NewReader(strings.NewReader(rawRequest))
	req, err := http.ReadRequest(b)
	common.Must(err)
	headers := []struct {
		Key   string
		Value string
	}{
		{
			Key:   "Foo",
			Value: "foo",
		},
		{
			Key:   "Bar",
			Value: "bar",
		},
		{
			Key:   "Connection",
			Value: "keep-alive,Foo, Bar",
		},
		{
			Key:   "Proxy-Connection",
			Value: "keep-alive",
		},
		{
			Key:   "Proxy-Authenticate",
			Value: "abc",
		},
	}
	for _, header := range headers {
		if v := req.Header.Get(header.Key); v != header.Value {
			t.Error("header ", header.Key, " = ", v, " want ", header.Value)
		}
	}

	RemoveHopByHopHeaders(req.Header)

	for _, header := range []string{"Connection", "Foo", "Bar", "Proxy-Connection", "Proxy-Authenticate"} {
		if v := req.Header.Get(header); v != "" {
			t.Error("header ", header, " = ", v)
		}
	}
}

func TestParseHost(t *testing.T) {
	testCases := []struct {
		RawHost     string
		DefaultPort net.Port
		Destination net.Destination
		Error       bool
	}{
		{
			RawHost:     "example.com:80",
			DefaultPort: 443,
			Destination: net.TCPDestination(net.DomainAddress("example.com"), 80),
		},
		{
			RawHost:     "tls.example.com",
			DefaultPort: 443,
			Destination: net.TCPDestination(net.DomainAddress("tls.example.com"), 443),
		},
		{
			RawHost:     "[2401:1bc0:51f0:ec08::1]:80",
			DefaultPort: 443,
			Destination: net.TCPDestination(net.ParseAddress("[2401:1bc0:51f0:ec08::1]"), 80),
		},
	}

	for _, testCase := range testCases {
		dest, err := ParseHost(testCase.RawHost, testCase.DefaultPort)
		if testCase.Error {
			if err == nil {
				t.Error("for test case: ", testCase.RawHost, " expected error, but actually nil")
			}
		} else {
			if dest != testCase.Destination {
				t.Error("for test case: ", testCase.RawHost, " expected host: ", testCase.Destination.String(), " but got ", dest.String())
			}
		}
	}
}

func TestAppendFastlyClientIP(t *testing.T) {
	testCases := []struct {
		name           string
		remoteAddr     string
		fastlyClientIP string
		existingXFF    string
		expectedXFF    string
	}{
		{
			name:           "Valid Fastly IP",
			remoteAddr:     "151.101.1.1",
			fastlyClientIP: "192.168.1.1",
			existingXFF:    "",
			expectedXFF:    "192.168.1.1",
		},
		{
			name:           "Valid Fastly IP with existing XFF",
			remoteAddr:     "151.101.1.1",
			fastlyClientIP: "192.168.1.1",
			existingXFF:    "10.0.0.1",
			expectedXFF:    "10.0.0.1, 192.168.1.1",
		},
		{
			name:           "Valid Fastly IP with duplicate IP",
			remoteAddr:     "151.101.1.1",
			fastlyClientIP: "192.168.1.1",
			existingXFF:    "10.0.0.1, 192.168.1.1",
			expectedXFF:    "10.0.0.1, 192.168.1.1",
		},
		{
			name:           "Non-Fastly IP",
			remoteAddr:     "192.168.1.1",
			fastlyClientIP: "192.168.1.1",
			existingXFF:    "10.0.0.1",
			expectedXFF:    "10.0.0.1",
		},
		{
			name:           "Valid Fastly IPv6",
			remoteAddr:     "2a04:4e40::1",
			fastlyClientIP: "192.168.1.1",
			existingXFF:    "",
			expectedXFF:    "192.168.1.1",
		},
		{
			name:           "No Fastly-Client-IP header",
			remoteAddr:     "151.101.1.1",
			fastlyClientIP: "",
			existingXFF:    "10.0.0.1",
			expectedXFF:    "10.0.0.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			header := http.Header{}
			if tc.existingXFF != "" {
				header.Set("X-Forwarded-For", tc.existingXFF)
			}
			if tc.fastlyClientIP != "" {
				header.Set("Fastly-Client-IP", tc.fastlyClientIP)
			}

			AppendFastlyClientIP(header, tc.remoteAddr)

			gotXFF := header.Get("X-Forwarded-For")
			if gotXFF != tc.expectedXFF {
				t.Errorf("X-Forwarded-For = %v, want %v", gotXFF, tc.expectedXFF)
			}
		})
	}
}
