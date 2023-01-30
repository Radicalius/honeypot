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

	"logging"
)

var logs *cloudwatchlogs.CloudWatchLogs

func handleLog(w http.ResponseWriter, r *http.Request) {
	cont, err := ioutil.ReadAll(r.Body)
	if handleError(w, err) {
		return
	}

	logMessage := new(logging.LogRequest)
	er := json.Unmarshal(cont, logMessage)
	if handleError(w, er) {
		return
	}

	now := time.Now().UnixNano() / 1000000

	_, er = logs.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{
				Message:   &logMessage.Message,
				Timestamp: &now,
			},
		},
		LogGroupName:  &logMessage.Service,
		LogStreamName: &logMessage.Session,
	})

	if handleError(w, er) {
		return
	}
}

func handleLogStream(w http.ResponseWriter, r *http.Request) {
	cont, err := ioutil.ReadAll(r.Body)
	if handleError(w, err) {
		return
	}

	logStream := new(logging.LogStreamRequest)
	er := json.Unmarshal(cont, logStream)
	if handleError(w, er) {
		return
	}

	resp, e := logs.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &logStream.Service,
	})
	if handleError(w, e) {
		return
	}

	if len(resp.LogGroups) == 0 {
		_, e = logs.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
			LogGroupName: &logStream.Service,
		})
		if handleError(w, e) {
			return
		}
	}

	start := time.Now().UnixNano() / 1000000
	startStr := fmt.Sprintf("%d", start)

	resp2, err2 := logs.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        &logStream.Service,
		LogStreamNamePrefix: &startStr,
	})
	if handleError(w, err2) {
		return
	}

	if len(resp2.LogStreams) == 0 {
		_, err := logs.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
			LogGroupName:  &logStream.Service,
			LogStreamName: &startStr,
		})
		if handleError(w, err) {
			return
		}
	}

	response := &logging.LogStreamResponse{
		Session: startStr,
	}
	body, err := json.Marshal(response)
	if handleError(w, err) {
		return
	}

	w.Write(body)
}

func handleError(w http.ResponseWriter, err error) bool {
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		response := &logging.ErrorResponse{
			Error: err.Error(),
		}
		body, err := json.Marshal(response)
		if err != nil {
			return true
		}
		w.Write(body)
		return true
	}

	return false
}

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	))

	logs = cloudwatchlogs.New(sess)

	http.HandleFunc("/v1/log", handleLog)
	http.HandleFunc("/v1/log-stream", handleLogStream)
	http.ListenAndServe(":25302", nil)
}
