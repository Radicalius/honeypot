package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type LogMessage struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
	Ip      string
}

var logs *cloudwatchlogs.CloudWatchLogs
var start int64 = time.Now().UnixNano() / 1000000
var start_str string = fmt.Sprintf("%d", start)
var service string = "http"

func log(content *LogMessage) {
	js, err := json.Marshal(content)
	if err != nil {
		fmt.Println(err)
		return
	}

	now := time.Now().UnixNano() / 1000000
	js_ := string(js)

	_, er := logs.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			&cloudwatchlogs.InputLogEvent{
				Message:   &js_,
				Timestamp: &now,
			},
		},
		LogGroupName:  &service,
		LogStreamName: &start_str,
	})

	if er != nil {
		fmt.Println(er)
	}
}

func headersToMap(headers http.Header) map[string]string {
	ret := make(map[string]string)
	for k := range headers {
		ret[k] = headers.Get(k)
	}
	return ret
}

func catchAll(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil && err.Error() != "EOF" {
		fmt.Println(err)
		return
	}

	log(&LogMessage{
		req.Method,
		req.RequestURI,
		headersToMap(req.Header),
		string(body),
		req.RemoteAddr,
	})
	w.WriteHeader(404)
	w.Write([]byte("404: File not found."))
}

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	))
	logs = cloudwatchlogs.New(sess)

	_, err := logs.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  &service,
		LogStreamName: &start_str,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", catchAll)
	http.ListenAndServe("0.0.0.0:8000", nil)
}
