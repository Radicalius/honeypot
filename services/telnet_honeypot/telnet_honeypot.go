package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"logging"
	"net"
	"os"
	"regexp"
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

var procMounts []byte
var elf []byte

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
			for _, command := range strings.Split(dataStr, ";") {
				command = strings.Trim(command, " \t\n\r")

				match, _ := regexp.MatchString(".*[A-Z]{5,6}.*", command)
				if match {
					exp, _ := regexp.Compile("[A-Z]{5,6}")
					token := exp.FindString(command)
					resp := fmt.Sprintf("%s: applet not found\r\n", token)
					conn.Write([]byte(resp))
				}

				if strings.Contains(command, "/proc/mounts") {
					conn.Write(procMounts)
				}

				match, _ = regexp.MatchString(".*echo -n?e '(\\\\x[0-9a-f]{2})+'", command)
				if match {
					exp, _ := regexp.Compile("\\\\x[0-9a-f]{2}")
					for _, digit := range exp.FindAllString(command, -1) {
						digit := strings.Replace(digit, "\\x", "", 1)
						data, _ := hex.DecodeString(digit)
						conn.Write(data)
					}
				}

				if command == "tftp" {
					conn.Write([]byte("tftp: applet not found\r\n"))
				}

				if command == "wget" {
					conn.Write([]byte("wget: missing URL\nUsage: wget [OPTION]... [URL]...\n\nTry `wget --help' for more options.\r\n"))
				}

				if strings.Contains(command, "dd") {
					conn.Write(elf)
					conn.Write([]byte("\r\n"))
				}

				if strings.Contains(command, "exit") {
					conn.Close()
				}
			}
		}

		logData := &TelnetLog{
			Ip:       conn.RemoteAddr().String(),
			Username: username,
			Password: password,
			Action:   state,
			Command:  &dataStr,
		}

		logger.Log(logData)

		if username != "" && password != "" {
			reporting.ReportIp(reporting.IpReport{
				Service: "telnet",
				Ip:      logData.IpAddress(),
				Data:    logData,
			})
		}
	}
}

func loadB64File(fname string) []byte {
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	dat, err = base64.StdEncoding.DecodeString(string(dat))
	if err != nil {
		log.Fatal(err)
	}

	return dat
}

func main() {

	procMounts = loadB64File("./proc_mounts.b64")
	elf = loadB64File("./elf.b64")

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
