package tcp

import (
	"fmt"
	"net"
)

func IP_finder() string {
	ip_adress := ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	for _, address := range addrs {
		// Check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("My local IP address is:", ipnet.IP.String())
				ip_adress = ipnet.IP.String()
				break // or continue if you want to list all non-loopback addresses
			}
		}
	}
	return ip_adress
}
