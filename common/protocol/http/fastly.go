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
	// Fastly IPv4 ranges
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
		"146.75.128.0/17",
		"151.101.0.0/16",
		"157.52.64.0/18",
		"157.52.128.0/17",
		"167.82.0.0/17",
		"167.82.128.0/20",
		"167.82.160.0/20",
		"167.82.224.0/20",
		"172.111.64.0/18",
		"172.111.128.0/17",
		"185.31.16.0/22",
		"199.27.72.0/21",
		"199.232.0.0/16",
	}

	// Fastly IPv6 ranges
	fastlyIPv6Cidrs := []string{
		"2a04:4e40::/32",
		"2a04:4e42::/32",
		"2a04:4e46::/32",
	}

	// Parse and store Fastly IPv4 ranges
	for _, cidr := range fastlyIPv4Cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		fastlyIPRanges = append(fastlyIPRanges, network)
	}

	// Parse and store Fastly IPv6 ranges
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
	// Initialize Fastly IP ranges if not already done
	fastlyRangesOnce.Do(initFastlyIPRanges)

	// Parse the IP address
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check if the IP address is in Fastly's IPv4 ranges
	if ip.To4() != nil {
		for _, network := range fastlyIPRanges {
			if network.Contains(ip) {
				return true
			}
		}
	} else {
		// Check if the IP address is in Fastly's IPv6 ranges
		for _, network := range fastlyIPv6Ranges {
			if network.Contains(ip) {
				return true
			}
		}
	}

	return false
}
