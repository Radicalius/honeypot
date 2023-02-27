package main

import (
	"fmt"
	"log"
	"logging"
	"net"
	"reporting"
	"strings"
)

type DnsLog struct {
	Ip   string
	Data string
}

func (log *DnsLog) IpAddress() string {
	return strings.Split(log.Ip, ":")[0]
}

var serviceName string = "dns"
var logger logging.Logger

func main() {

	logger = logging.Logger{
		Service: &serviceName,
	}

	udpServer, err := net.ListenPacket("udp", ":53")
	if err != nil {
		log.Fatal(err)
	}
	defer udpServer.Close()

	for {
		buf := make([]byte, 10240)
		n, addr, err := udpServer.ReadFrom(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if n == 0 {
			continue
		}

		dnsLog := &DnsLog{
			Ip:   addr.String(),
			Data: string(buf[:n]),
		}

		logger.Log(dnsLog)

		reporting.ReportIp(reporting.IpReport{
			Service: serviceName,
			Ip:      dnsLog.IpAddress(),
			Data:    dnsLog,
		})
	}
}
