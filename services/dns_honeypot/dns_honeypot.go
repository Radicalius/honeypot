package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"logging"
	"net"
	"reporting"
	"strconv"
	"strings"
)

func parseName(data []byte) ([]string, int) {
	p := 0
	name := make([]string, 0)
	for {
		if data[p] == 0 {
			break
		}

		length := int(data[p])
		p++
		domainPart := string(data[p : p+length])
		name = append(name, domainPart)
		p += length
	}
	return name, p
}

func writeName(buf *bytes.Buffer, name []string) {
	for _, n := range name {
		length := len([]rune(n))
		lenByte := make([]byte, 1)
		lenByte[0] = byte(length)
		buf.Write(lenByte)
		buf.Write([]byte(n))
	}

	buf.Write([]byte{byte(0)})
}

func writeBinaryIpAddess(buf *bytes.Buffer, ip string) {
	parts := strings.Split(ip, ".")
	for _, p := range parts {
		octet := make([]byte, 1)
		i, err := strconv.Atoi(p)
		if err != nil {
			fmt.Println(err)
			i = 0
		}
		octet[0] = byte(i)
		buf.Write(octet)
	}
}

func bytesIpToString(data []byte) string {
	return strconv.Itoa(int(data[0])) + "." + strconv.Itoa(int(data[1])) + "." + strconv.Itoa(int(data[2])) + "." + strconv.Itoa(int(data[3]))
}

func boolToInt(inp bool) uint16 {
	if inp {
		return 1
	}

	return 0
}

type DnsPacket struct {
	Header    DnsHeader
	Questions []DnsQuestion
	Answers   []DnsAnswer
}

func (packet *DnsPacket) Parse(data []byte) {
	packet.Header = DnsHeader{}
	packet.Header.Parse(data)

	offset := 12
	for i := 0; i < int(packet.Header.Questions); i++ {
		question := DnsQuestion{}
		offset += question.Parse(data[offset:])
		packet.Questions = append(packet.Questions, question)
	}

	for i := 0; i < int(packet.Header.Answers); i++ {
		answer := DnsAnswer{}
		offset += answer.Parse(data[offset:])
		packet.Answers = append(packet.Answers, answer)
	}
}

func (packet *DnsPacket) Encode() []byte {
	buf := &bytes.Buffer{}

	packet.Header.Write(buf)

	for _, q := range packet.Questions {
		q.Write(buf)
	}

	for _, a := range packet.Answers {
		a.Write(buf)
	}

	return buf.Bytes()
}

type DnsHeader struct {
	Identification      uint16
	QuestionResponse    bool
	Opcode              uint16
	AuthoritativeAnswer bool
	RecursionDesired    bool
	RecursionAvailable  bool
	ResponseCode        uint16
	Questions           uint16
	Answers             uint16
	Authorities         uint16
	Additionals         uint16
}

func (header *DnsHeader) Parse(data []byte) {
	header.Identification = binary.BigEndian.Uint16(data[0:2])
	flags := binary.BigEndian.Uint16(data[2:4])
	header.QuestionResponse = (flags >> 15) == 1
	header.Opcode = (flags >> 11) & 15
	header.AuthoritativeAnswer = ((flags >> 10) & 1) == 1
	header.RecursionDesired = ((flags >> 9) & 1) == 1
	header.RecursionAvailable = ((flags >> 8) & 1) == 1
	header.ResponseCode = flags & 15
	header.Questions = binary.BigEndian.Uint16(data[4:6])
	header.Answers = binary.BigEndian.Uint16(data[6:8])
	header.Authorities = binary.BigEndian.Uint16(data[8:10])
	header.Additionals = binary.BigEndian.Uint16(data[10:12])
}

func (header *DnsHeader) Write(buf *bytes.Buffer) {
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, header.Identification)
	buf.Write(buffer)

	var flags uint16 = 0
	flags |= (boolToInt(header.QuestionResponse)) << 15
	flags |= (header.Opcode << 12)
	flags |= (boolToInt(header.AuthoritativeAnswer) << 11)
	flags |= (boolToInt(header.RecursionAvailable) << 10)
	flags |= (boolToInt(header.RecursionDesired) << 9)
	flags |= header.ResponseCode
	binary.BigEndian.PutUint16(buffer, flags)
	buf.Write(buffer)

	binary.BigEndian.PutUint16(buffer, header.Questions)
	buf.Write(buffer)
	binary.BigEndian.PutUint16(buffer, header.Answers)
	buf.Write(buffer)
	binary.BigEndian.PutUint16(buffer, header.Authorities)
	buf.Write(buffer)
	binary.BigEndian.PutUint16(buffer, header.Additionals)
	buf.Write(buffer)
}

type DnsQuestion struct {
	Name       []string
	QueryType  uint16
	QueryClass uint16
}

func (question *DnsQuestion) Parse(data []byte) int {
	var p int = 0
	question.Name, p = parseName(data)
	p++

	question.QueryType = binary.BigEndian.Uint16(data[p : p+2])
	question.QueryClass = binary.BigEndian.Uint16(data[p+2 : p+4])
	return p + 4
}

func (question *DnsQuestion) Write(buf *bytes.Buffer) {
	writeName(buf, question.Name)

	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, question.QueryType)
	buf.Write(buffer)
	binary.BigEndian.PutUint16(buffer, question.QueryClass)
	buf.Write(buffer)
}

type DnsAnswer struct {
	Name         []string
	Type         uint16
	Class        uint16
	TTL          uint16
	DataLength   uint16
	ResponseData []string
}

func (answer *DnsAnswer) Parse(data []byte) int {
	var p int = 0
	answer.Name, p = parseName(data)
	p++

	answer.Type = binary.BigEndian.Uint16(data[p : p+2])
	answer.Class = binary.BigEndian.Uint16(data[p+2 : p+4])
	answer.TTL = binary.BigEndian.Uint16(data[p+4 : p+6])
	answer.DataLength = binary.BigEndian.Uint16(data[p+6 : p+8])
	p += 8

	for i := 0; i < int(answer.DataLength); i++ {
		answer.ResponseData = append(answer.ResponseData, bytesIpToString(data[p:]))
		p += 4
	}

	return p
}

func (answer *DnsAnswer) Write(buf *bytes.Buffer) {
	writeName(buf, answer.Name)

	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, answer.Type)
	buf.Write(buffer)
	binary.BigEndian.PutUint16(buffer, answer.Class)
	buf.Write(buffer)
	binary.BigEndian.PutUint16(buffer, answer.TTL)
	buf.Write(buffer)
	binary.BigEndian.PutUint16(buffer, answer.DataLength)
	buf.Write(buffer)

	for _, ip := range answer.ResponseData {
		writeBinaryIpAddess(buf, ip)
	}
}

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

		packet := DnsPacket{}
		packet.Parse(buf)
		req, err := json.Marshal(packet)
		if err != nil {
			fmt.Println(err)
			continue
		}

		dnsLog := &DnsLog{
			Ip:   addr.String(),
			Data: string(req),
		}

		logger.Log(dnsLog)

		reporting.ReportIp(reporting.IpReport{
			Service: serviceName,
			Ip:      dnsLog.IpAddress(),
			Data:    dnsLog,
		})
	}
}
