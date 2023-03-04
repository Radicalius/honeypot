package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"logging"
	"reporting"
)

type LogMessage struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
	Ip      string
}

func (message *LogMessage) IpAddress() string {
	return strings.Split(message.Ip, ":")[0]
}

var log *logging.Logger
var service string = "http"

func headersToMap(headers http.Header) map[string]string {
	ret := make(map[string]string)
	for k := range headers {
		ret[k] = headers.Get(k)
	}
	return ret
}

func catchAll(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := &LogMessage{
		req.Method,
		req.RequestURI,
		headersToMap(req.Header),
		string(body),
		req.RemoteAddr,
	}

	log.Log(data)

	if (data.Path != "/" && data.Path != "/favicon.ico") || data.Body != "" {
		reporting.ReportIp(reporting.IpReport{
			Service: "http",
			Ip:      data.IpAddress(),
			Data:    data,
		})
	}

	w.WriteHeader(404)
	w.Write([]byte("404: File not found."))
}

func main() {
	log = &logging.Logger{
		Service: &service,
	}

	http.HandleFunc("/", catchAll)
	http.ListenAndServe("0.0.0.0:80", nil)
}
