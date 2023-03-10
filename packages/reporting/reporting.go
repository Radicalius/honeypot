package reporting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type IpReport struct {
	Service string
	Ip      string
	Data    interface{}
}

var client = &http.Client{}
var deduplicate = make(map[string]int64)

func ReportIp(report IpReport) {

	if !ValidateRequest(report.Ip) {
		return
	}

	go AbuseIpDbReport(report)
}

func AbuseIpDbReport(report IpReport) {

	var abuseIpDbCategories string
	if report.Service == "http" {
		abuseIpDbCategories = "21"
	} else if report.Service == "telnet" {
		abuseIpDbCategories = "18,23"
	} else if report.Service == "ftp" {
		abuseIpDbCategories = "5,18"
	} else if report.Service == "dns" {
		abuseIpDbCategories = "2"
	} else if report.Service == "smtp" {
		abuseIpDbCategories = "11"
	}

	comment, err := json.Marshal(report.Data)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := url.Values{}
	data.Set("ip", report.Ip)
	data.Set("categories", abuseIpDbCategories)
	data.Set("comment", string(comment))

	requestBuffer := strings.NewReader(data.Encode())
	httpRequest, err := http.NewRequest("POST", "https://api.abuseipdb.com/api/v2/report", requestBuffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Key", os.Getenv("AbuseIPDBKey"))

	response, err := client.Do(httpRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	if response.StatusCode > 299 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("ERROR: Error reporting to AbuseIPDB: " + string(body))
	}
}

func ValidateRequest(reportIp string) bool {
	env, exists := os.LookupEnv("Environment")
	if !exists || env != "Production" {
		fmt.Println("WARNING: Non-production Environment.  Skipping Report.")
		return false
	}

	exclude, exists := os.LookupEnv("ExcludedIPs")
	if exists && exclude != "" {
		for _, ip := range strings.Split(exclude, ",") {
			if strings.Contains(reportIp, ip) {
				fmt.Println("WARNING: skipped report because of ip filter: " + ip)
				return false
			}
		}
	}

	if deduplicate[reportIp] > (time.Now().Unix() - 3600) {
		return false
	}

	deduplicate[reportIp] = time.Now().Unix()
	return true
}
