package aliveMessages

import (
	"log"
	"net"
	"os"
	"project/variables"
	"sort"
	"strconv"
	"strings"
	"time"
)

const PORT = "9999"
const BROADCAST_ADDR = "255.255.255.255:" + PORT
const BROADCAST_PERIOD = 100 * time.Millisecond
const LISTEN_ADDR = "0.0.0.0:" + PORT
const LISTEN_TIMEOUT = 10 * time.Second
const NODE_LIFE = 3 * time.Second
const ALLOWED_CONSECUTIVE_ERRORS = 100

func BroadcastLife() {

	myIP := variables.MyIP

	conn, err := net.Dial("udp4", BROADCAST_ADDR)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(BROADCAST_PERIOD)
	defer ticker.Stop()

	var errCount int = 0

	for range ticker.C {

		message := "Please give us an A on the project:)"
		_, err := conn.Write([]byte(message))
		if err != nil {
			errCount++

			if errCount > ALLOWED_CONSECUTIVE_ERRORS {
				log.Println("Too many consecutive udp errors, Restarting UDP connection")
				conn.Close()
				for {
					conn, err = net.Dial("udp4", BROADCAST_ADDR)
					if err != nil {
						log.Println(err)
						time.Sleep(time.Second)
					} else {
						break
					}
				}

				errCount = 0
			}
		}
		if variables.MyIP != myIP {
			myIP = variables.MyIP
			log.Println("Changed IP, Restarting UDP connection")
			conn.Close()
			conn, err = net.Dial("udp4", BROADCAST_ADDR)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func LookForLife(livingIPsChan chan<- []string) {

	myIP := variables.MyIP

	IPLifetimes := make(map[string]time.Time)
	IPLifetimes[myIP] = time.Now().Add(time.Hour)

	pc, err := net.ListenPacket("udp", LISTEN_ADDR)
	if err != nil {
		log.Println(err)
		return
	}
	defer pc.Close()

	buffer := make([]byte, 32768)

	for {
		if variables.MyIP != myIP {
			IPLifetimes[myIP] = time.Now()
			myIP = variables.MyIP
			IPLifetimes[myIP] = time.Now().Add(time.Hour)
		}

		err := pc.SetReadDeadline(time.Now().Add(LISTEN_TIMEOUT))
		if err != nil {
			log.Println("Failed to set a deadline for the read operation:", err)
		}

		_, addr, err := pc.ReadFrom(buffer)
		var addrString string
		if addr != nil {
			addrString = strings.Split(addr.String(), ":")[0]
		} else {
			addrString = ""
		}

		if err != nil {
			if os.IsTimeout(err) {
				IPLifetimes = updateLivingIPs(IPLifetimes, "", myIP)
				livingIPsChan <- getLivingIPs(IPLifetimes)
				continue
			} else {
				log.Println("Read error:", err)
				continue
			}
		} else {

			IPLifetimes = updateLivingIPs(IPLifetimes, addrString, myIP)
			livingIPsChan <- getLivingIPs(IPLifetimes)
		}
	}
}

func updateLivingIPs(IPLifetimes map[string]time.Time, newAddr string, myIP string) map[string]time.Time {

	if newAddr == "" {
		for addrInList := range IPLifetimes {
			if addrInList != myIP {
				IPLifetimes[addrInList] = time.Now()
			}
		}
	} else {
		_, ok := IPLifetimes[newAddr]
		if !ok {
			if newAddr != myIP {
				log.Println("New node discovered: ", newAddr)
			} else {
				log.Println()
				log.Println("This is my IP: ", myIP)
			}
			IPLifetimes[newAddr] = time.Now().Add(NODE_LIFE)
		} else {
			if myIP != newAddr {
				IPLifetimes[newAddr] = time.Now().Add(NODE_LIFE)
			}
		}
	}
	return IPLifetimes
}

func getLivingIPs(m map[string]time.Time) []string {
	livingIPs := []string{}
	for address, death := range m {
		if death.After(time.Now()) {
			livingIPs = append(livingIPs, address)
		}
	}
	livingIPs = ipSorter(livingIPs)
	return livingIPs
}

func ipSorter(ipStrings []string) []string {
	ipMap := make(map[string]int)
	var ipStringsNew []string
	var ipIntsNew []int
	for _, ipStr := range ipStrings {
		ipInt, _ := strconv.Atoi(strings.Replace(ipStr, ".", "", -1))
		ipMap[ipStr] = ipInt
		ipIntsNew = append(ipIntsNew, ipInt)
	}
	sort.Ints(ipIntsNew)
	for _, ipInt := range ipIntsNew {
		for ipStr, ipInt2 := range ipMap {
			if ipInt == ipInt2 {
				ipStringsNew = append(ipStringsNew, ipStr)
			}
		}
	}
	return ipStringsNew
}

func GetPrimaryIP() string {
	var primaryIP string
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Println("Error at GetPrimaryIP(): ", err)
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	primaryIP = localAddr.IP.String()
	return primaryIP
}
