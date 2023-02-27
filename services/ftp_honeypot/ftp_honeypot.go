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

var serviceName string = "ftp"
var logger logging.Logger

type FtpLog struct {
	Ip       string
	Username string
	Password string
	Command  string
}

func (log *FtpLog) IpAddress() string {
	return strings.Split(log.Ip, ":")[0]
}

func handleRequest(conn net.Conn) {
	var username string
	var password string

	conn.Write([]byte("220 Welcome to the DLP Test FTP Server\r\n"))

	for {
		var data []byte = make([]byte, 10240)
		var n int
		var err error

		n, err = conn.Read(data)
		if n == 0 {
			continue
		}

		if err != nil {
			fmt.Errorf("Error parsing message %s", err.Error())
			continue
		}

		dataStr := string(data[:n])
		dataStr = strings.Replace(dataStr, "\r\n", "", 1)
		parts := strings.Split(dataStr, " ")
		command := parts[0]

		fmt.Println(command)

		if command == "USER" {
			username = parts[1]
			conn.Write([]byte("331 Please specify the password.\r\n"))
		} else if command == "PASS" {
			password = parts[1]
			conn.Write([]byte("230 Login successful.\r\n"))
		} else if command == "SYST" {
			conn.Write([]byte("215 UNIX Type: L8\r\n"))
		} else if command == "CLOSE" || command == "EXIT" || command == "QUIT" {
			conn.Write([]byte("221 Goodbye.\r\n"))
			conn.Close()
		} else {
			conn.Write([]byte("500 Invalid command.\r\n"))
		}

		logData := &FtpLog{
			Ip:       conn.RemoteAddr().String(),
			Username: username,
			Password: password,
			Command:  dataStr,
		}

		logger.Log(logData)

		reporting.ReportIp(reporting.IpReport{
			Service: serviceName,
			Ip:      logData.IpAddress(),
			Data:    logData,
		})
	}
}

func main() {
	logger = logging.Logger{
		Service: &serviceName,
	}

	l, err := net.Listen("tcp", ":23")
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
