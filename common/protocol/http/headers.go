package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/xtls/xray-core/common/net"
)

// ParseXForwardedFor parses X-Forwarded-For header in http headers, and return the IP list in it.
func ParseXForwardedFor(header http.Header) []net.Address {
	xff := header.Get("X-Forwarded-For")
	if xff == "" {
		return nil
	}
	list := strings.Split(xff, ",")
	addrs := make([]net.Address, 0, len(list))
	for _, proxy := range list {
		addrs = append(addrs, net.ParseAddress(proxy))
	}
	return addrs
}

// RemoveHopByHopHeaders removes hop by hop headers in http header list.
func RemoveHopByHopHeaders(header http.Header) {
	// Strip hop-by-hop header based on RFC:
	// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.5.1
	// https://www.mnot.net/blog/2011/07/11/what_proxies_must_do

	header.Del("Proxy-Connection")
	header.Del("Proxy-Authenticate")
	header.Del("Proxy-Authorization")
	header.Del("TE")
	header.Del("Trailers")
	header.Del("Transfer-Encoding")
	header.Del("Upgrade")

	connections := header.Get("Connection")
	header.Del("Connection")
	if connections == "" {
		return
	}
	for _, h := range strings.Split(connections, ",") {
		header.Del(strings.TrimSpace(h))
	}
}

// ParseHost splits host and port from a raw string. Default port is used when raw string doesn't contain port.
func ParseHost(rawHost string, defaultPort net.Port) (net.Destination, error) {
	port := defaultPort
	host, rawPort, err := net.SplitHostPort(rawHost)
	if err != nil {
		if addrError, ok := err.(*net.AddrError); ok && strings.Contains(addrError.Err, "missing port") {
			host = rawHost
		} else {
			return net.Destination{}, err
		}
	} else if len(rawPort) > 0 {
		intPort, err := strconv.Atoi(rawPort)
		if err != nil {
			return net.Destination{}, err
		}
		port = net.Port(intPort)
	}

	return net.TCPDestination(net.ParseAddress(host), port), nil
}

// AppendFastlyClientIP appends real client IP from Fastly-Client-IP header to X-Forwarded-For header if the request comes from Fastly CDN
func AppendFastlyClientIP(header http.Header, remoteAddr string) {
	// Extract IP from remoteAddr (format: "IP:port")
	remoteIP := remoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		remoteIP = host
	}

	// Check if the request is coming from Fastly's IP range
	if !IsIPInFastlyRange(remoteIP) {
		return
	}

	// Get real client IP from Fastly headers in order of preference
	var clientIP string

	// 1. Try Fastly-Client-IP (most reliable)
	clientIP = header.Get("Fastly-Client-IP")

	// 2. Try True-Client-IP as fallback
	if clientIP == "" {
		clientIP = header.Get("X-Forwarded-For")
	}

	// Get and clean the X-Forwarded-For header
	xff := header.Get("X-Forwarded-For")
	if xff == "" {
		header.Set("X-Forwarded-For", clientIP)
		return
	}

	// Split XFF into IPs and filter out Fastly IPs
	ips := strings.Split(xff, ",")
	filteredIPs := make([]string, 0, len(ips))

	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		// Only keep non-empty, valid, non-Fastly IPs
		if ip != "" && net.ParseIP(ip) != nil && !IsIPInFastlyRange(ip) {
			filteredIPs = append(filteredIPs, ip)
		}
	}

	if len(filteredIPs) > 0 {
		header.Set("X-Forwarded-For", strings.Join(filteredIPs, ", "))
	} else {
		header.Set("X-Forwarded-For", clientIP)
	}

	// Remove Fastly-specific headers to prevent spoofing in subsequent requests
	header.Del("Fastly-Client-IP")
	header.Del("True-Client-IP")
}
