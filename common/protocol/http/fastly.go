package http

import (
	"net"
	"sync"
)

var (
	fastlyIPRanges     []*net.IPNet
	fastlyIPv6Ranges   []*net.IPNet
	fastlyRangesOnce   sync.Once
)

// Initialize the Fastly IP ranges
func initFastlyIPRanges() {
	fastlyIPv4Cidrs := []string{
		"23.235.32.0/20",
		"43.249.72.0/22",
		"103.244.50.0/24",
		"103.245.222.0/23",
		"103.245.224.0/24",
		"104.156.80.0/20",
		"140.248.64.0/18",
		"140.248.128.0/17",
		"146.75.0.0/17",
		"151.101.0.0/16",
		"157.52.64.0/18",
		"167.82.0.0/17",
		"167.82.128.0/20",
		"167.82.160.0/20",
		"167.82.224.0/20",
		"172.111.64.0/18",
		"185.31.16.0/22",
		"199.27.72.0/21",
		"199.232.0.0/16",
	}

	fastlyIPv6Cidrs := []string{
		"2a04:4e40::/32",
		"2a04:4e42::/32",
	}

	for _, cidr := range fastlyIPv4Cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		fastlyIPRanges = append(fastlyIPRanges, network)
	}

	for _, cidr := range fastlyIPv6Cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		fastlyIPv6Ranges = append(fastlyIPv6Ranges, network)
	}
}

// IsIPInFastlyRange checks if an IP address is in Fastly's IP ranges
func IsIPInFastlyRange(ipStr string) bool {
	fastlyRangesOnce.Do(initFastlyIPRanges)

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	if ip.To4() != nil {
		// IPv4
		for _, network := range fastlyIPRanges {
			if network.Contains(ip) {
				return true
			}
		}
	} else {
		// IPv6
		for _, network := range fastlyIPv6Ranges {
			if network.Contains(ip) {
				return true
			}
		}
	}

	return false
}
