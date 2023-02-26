package main

import (
	"fmt"
	"log"
	"logging"
	"net"
	"os"
	"reporting"
	"strings"
)

type TelnetLog struct {
	Ip       string
	Username string
	Password string
	Action   string
	Command  *string
}

func (log *TelnetLog) IpAddress() string {
	return strings.Split(log.Ip, ":")[0]
}

var serviceName string = "telnet"
var logger logging.Logger

func handleRequest(conn net.Conn) {
	state := "username"

	var username string
	var password string

	for {
		var data []byte = make([]byte, 10240)
		var n int
		var err error

		if state == "username" {
			conn.Write([]byte("Username: "))
		} else if state == "password" {
			conn.Write([]byte("Password: "))
		} else {
			conn.Write([]byte("/ # "))
		}

		n, err = conn.Read(data)
		if n == 0 {
			continue
		}

		if err != nil {
			fmt.Errorf("Error parsing message %s", err.Error())
			continue
		}

		dataStr := string(data[:n-1])

		if state == "username" {
			username = dataStr
			state = "password"
		} else if state == "password" {
			password = dataStr
			state = "command"
		} else {
			data := &TelnetLog{
				Ip:       conn.RemoteAddr().String(),
				Username: username,
				Password: password,
				Action:   state,
				Command:  &dataStr,
			}

			logger.Log(data)
			reporting.ReportIp(reporting.IpReport{
				Service: "telnet",
				Ip:      data.IpAddress(),
				Data:    data,
			})
		}
	}
}

func main() {
	logger = logging.Logger{
		Service: &serviceName,
	}

	l, err := net.Listen("tcp", ":21")
	if err != nil {
		log.Fatal(fmt.Sprintf("Error opening socket: %s", err.Error()))
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleRequest(conn)
	}
}
