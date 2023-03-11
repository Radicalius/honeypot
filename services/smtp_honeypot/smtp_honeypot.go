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

type SmtpLog struct {
	Ip   string
	Data []string
}

func (log *SmtpLog) IpAddress() string {
	return strings.Split(log.Ip, ":")[0]
}

var logger logging.Logger
var serviceName string = "smtp"

func handleRequest(conn net.Conn) {
	sessionLog := &SmtpLog{
		Ip: conn.RemoteAddr().String(),
	}
	var data []byte = make([]byte, 10240)

	conn.Write([]byte("220\r\n"))

	for {
		n, err := conn.Read(data)
		if err != nil {
			fmt.Println(err)
			break
		}

		if n == 0 {
			break
		}

		strData := string(data[:n])
		strData = strings.Trim(strData, "\r\n")
		sessionLog.Data = append(sessionLog.Data, strData)

		parts := strings.Split(strData, " ")
		cmd := parts[0]

		if cmd == "HELO" {
			resp := fmt.Sprintf("250 Hello %s\r\n", parts[1])
			conn.Write([]byte(resp))
		} else if cmd == "DATA" {
			conn.Write([]byte("354 End data with <CR><LF>.<CR><LF>\r\n"))
		} else if strings.Contains(cmd, "QUIT") {
			conn.Write([]byte("221 Bye\r\n"))
			break
		} else {
			conn.Write([]byte("250 Ok\r\n"))
		}
	}

	logger.Log(sessionLog)
	reporting.ReportIp(reporting.IpReport{
		Service: serviceName,
		Ip:      sessionLog.IpAddress(),
		Data:    sessionLog,
	})

	conn.Close()
}

func main() {

	logger = logging.Logger{
		Service: &serviceName,
	}

	l, err := net.Listen("tcp", ":25")
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
