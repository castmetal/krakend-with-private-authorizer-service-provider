package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func SendRequest(urlRequest string, method string, params map[string]interface{}, headers map[string]string) ([]byte, *http.Response, error) {
	var jsonData []byte
	var sendError error
	var body []byte

	if (method == "POST" || method == "PATCH" || method == "PUT") && params != nil && len(params) > 0 {
		jsonString, err := json.Marshal(params)
		if err != nil {
			log.Fatal("Can not convert body to json")
			return nil, nil, err
		}
		jsonData = []byte(jsonString)
	}

	req, err := http.NewRequest(method, urlRequest, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	d1 := make(chan bool)
	d2 := make(chan bool)
	d3 := make(chan bool)
	go setParamsWithGet(d1, req, method, params)
	go setHeadersToRequest(d2, req, headers)
	go setBasicAuth(d3, req)
	<-d1
	<-d2
	<-d3

	client := &http.Client{}
	res, e := client.Do(req)

	if res != nil && res.Body != nil {
		body, _ = ioutil.ReadAll(res.Body)
	}

	if e == nil {
		defer res.Body.Close()
	}

	if res == nil {
		sendError = errors.New(strconv.Itoa(500))
	} else if res != nil && res.StatusCode > 299 {
		sendError = errors.New(strconv.Itoa(res.StatusCode))
	}

	return body, res, sendError
}

func setParamsWithGet(done chan bool, req *http.Request, method string, params map[string]interface{}) {
	if method == "GET" && params != nil && len(params) > 0 {
		setParamsRequestUrl(req, params)
	}
	done <- true
}

func setParamsRequestUrl(req *http.Request, params map[string]interface{}) {
	q := req.URL.Query()
	for p, v := range params {
		valStr := fmt.Sprint(v)
		q.Add(p, valStr)
	}
}

func setHeadersToRequest(done chan bool, req *http.Request, headers map[string]string) {
	for p, v := range headers {
		valStr := fmt.Sprint(v)
		req.Header.Add(p, valStr)
	}

	done <- true
}

func setBasicAuth(done chan bool, req *http.Request) {
	env := GetEnvVariable("ENV")
	if env == "staging" {
		basicUser := GetEnvVariable("BASIC_AUTH_USERNAME")
		basicPassword := GetEnvVariable("BASIC_AUTH_PASSWORD")
		req.SetBasicAuth(basicUser, basicPassword)
	}

	done <- true
}
