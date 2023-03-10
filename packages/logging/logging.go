package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type LogRequest struct {
	Service string
	Session string
	Message string
}

type LogStreamRequest struct {
	Service string
}

type LogStreamResponse struct {
	Session string
}

type ErrorResponse struct {
	Error string
}

type Logger struct {
	Service *string
	Session *string
}

type LogMessage interface {
	IpAddress() string
}

func (logger *Logger) Log(message LogMessage) {

	env, exists := os.LookupEnv("Environment")
	if !exists || env != "Production" {
		fmt.Println("WARNING: Non-production Environment.  Skipping Log Message.")
		return
	}

	exclude, exists := os.LookupEnv("ExcludedIPs")
	if exists && exclude != "" {
		for _, ip := range strings.Split(exclude, ",") {
			if strings.Contains(message.IpAddress(), ip) {
				fmt.Println("WARNING: skipped log message because of ip filter: " + ip)
				return
			}
		}
	}

	logServerBaseUrl, exists := os.LookupEnv("LogServerBaseURL")
	if !exists {
		log.Fatal("LogServerBaseURL is a required envvar")
	}

	if logger.Session == nil {
		if !logger.initLogSession(logServerBaseUrl) {
			return
		}
	}

	log := &LogRequest{
		Service: *logger.Service,
		Session: *logger.Session,
	}

	body, err4 := json.Marshal(message)
	if err4 != nil {
		fmt.Println(err4)
		return
	}
	log.Message = string(body)

	logData, err := json.Marshal(log)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err2 := http.Post(logServerBaseUrl+"/v1/log", "application/json", bytes.NewReader(logData))
	if err2 != nil {
		fmt.Println(err2)
		return
	}

	respBody, err3 := ioutil.ReadAll(resp.Body)
	if err3 != nil {
		fmt.Println(err3)
	}

	if resp.StatusCode != 200 {
		errResp := &ErrorResponse{}
		err3 := json.Unmarshal(respBody, errResp)
		if err3 != nil {
			fmt.Println(err3)
			return
		}

		fmt.Println(errResp.Error)
	}
}

func (logger *Logger) initLogSession(logServerBaseUrl string) bool {
	request := &LogStreamRequest{
		Service: *logger.Service,
	}
	body, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		return false
	}

	resp, err2 := http.Post(logServerBaseUrl+"/v1/log-stream", "application/json", bytes.NewReader(body))
	if err2 != nil {
		fmt.Println(err)
		return false
	}

	respBody, err3 := ioutil.ReadAll(resp.Body)
	if err3 != nil {
		fmt.Println(err3)
		return false
	}

	if resp.StatusCode == 200 {
		sessionResp := &LogStreamResponse{}
		err4 := json.Unmarshal(respBody, sessionResp)
		if err4 != nil {
			fmt.Println(err4)
			return false
		}

		logger.Session = &sessionResp.Session
		return true
	} else {
		sessionResp := &ErrorResponse{}
		err5 := json.Unmarshal(respBody, sessionResp)
		if err5 != nil {
			fmt.Println(err5)
			return false
		}

		fmt.Println(sessionResp.Error)
	}

	return false
}
