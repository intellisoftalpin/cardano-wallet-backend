package helpers

import "net"

func StringToIP(ips []string) (ipList []net.IP) {
	for _, ip := range ips {
		ipList = append(ipList, net.ParseIP(ip))
	}

	return ipList
}
