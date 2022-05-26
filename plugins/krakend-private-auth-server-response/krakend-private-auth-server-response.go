package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/satori/go.uuid"

	_utils "krakend-private-auth-server-response/utils"
)

// download code on github.com/castmetal/krakend-private-auth-server-response
const Namespace = "krakend-private-auth-server-response"

type RequestGatewayResponse struct {
	Body    map[string]interface{}
	Headers map[string][]string
	Status  int
}

type ParamsRequest struct {
	Params      map[string]interface{}
	Headers     map[string]string
	AuthUrl     string
	AuthMethod  string
	PrivateFlag string
	ErrorFlag   string
}

type RequestAuthResponse struct {
	Body   map[string]interface{}
	Status int
}

type statusRecorder struct {
	http.ResponseWriter
	status  int
	buf     *bytes.Buffer
	written bool
}

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(Namespace)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func generateUuid() string {
	enc, err := json.Marshal(uuid.NewV4())
	if err != nil {
		fmt.Println(err)
		return ""
	}

	uuidResponse := strings.Replace(string(enc), "\"", "", -1)

	return uuidResponse
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, handler http.Handler) (http.Handler, error) {
	fmt.Println("\n[PRIVATE-AUTH-SERVER-RESPONSE]: HANDLING REQUESTS\n")

	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var uid string

		requestArgs := getParamsToRequest(extra, req)

		uri := req.RequestURI
		private_flag := "/" + requestArgs.PrivateFlag

		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(uri, private_flag) {
			authResponse, err := sendRequestAuth(requestArgs)

			if err != nil {
				req.Header.Set("x-user", "")
			}

			if err != nil && authResponse.Status >= 500 {
				w.WriteHeader(authResponse.Status)
				json.NewEncoder(w).Encode(authResponse.Body)

				return
			} else if authResponse.Status >= 400 {
				w.WriteHeader(authResponse.Status)
				json.NewEncoder(w).Encode(authResponse.Body)

				return
			}

			uid = generateUuid()

			authResponse.Body["_valid-uuid-req-server"] = uid
			b, _ := json.Marshal(authResponse.Body)

			req.Header.Set("x-user", string(b))
		} else {
			uid = generateUuid()
		}

		setHeaders(req, uid)
		rec := statusRecorder{w, 401, &bytes.Buffer{}, false}

		handler.ServeHTTP(&rec, req.WithContext(ctx))

		m := make(map[string]interface{})
		err := json.Unmarshal(rec.buf.Bytes(), &m)
		if err != nil {
			fmt.Println(err)
		}

		if m[requestArgs.ErrorFlag] != nil {
			res := make(map[string]interface{})
			for k, v := range m[requestArgs.ErrorFlag].(map[string]interface{}) {
				res[k] = v
			}

			resBody := new(map[string]interface{})
			e := json.Unmarshal([]byte(res["http_body"].(string)), &resBody)
			_utils.ErrorHandling(e)

			status := res["http_status_code"].(float64)

			w.WriteHeader(int(status))
			json.NewEncoder(w).Encode(resBody)

			return
		}

		w.WriteHeader(rec.status)

		if rec.status == 500 {
			resBody := make(map[string]interface{})
			resBody["message"] = "Serviço indisponível, tente novamente em instantes."

			json.NewEncoder(w).Encode(resBody)

			return
		}

		json.NewEncoder(w).Encode(m)
	}), nil
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.written = true
	rec.status = code
}

func (rec *statusRecorder) Write(p []byte) (int, error) {
	return rec.buf.Write(p)
}

func setHeaders(req *http.Request, uid string) {
	req.Header.Set("x-origin", req.RequestURI)
	req.Header.Set("x-request-id", uid)
}

func getParamsToRequest(extra map[string]interface{}, req *http.Request) ParamsRequest {
	var paramsMap map[string]interface{}

	paramsMap = extra[Namespace].(map[string]interface{})

	tokenHeader := fmt.Sprint(paramsMap["token_header"])
	headers := make(map[string]string)
	headers["content-type"] = "application/json"
	headers[tokenHeader] = req.Header.Get(tokenHeader)

	p, err := json.Marshal(paramsMap["params"])
	if err != nil {
		fmt.Println(err)
	}

	params := make(map[string]interface{})
	err = json.Unmarshal(p, &params)
	if err != nil {
		fmt.Println(err)
		params = make(map[string]interface{})
	}

	return ParamsRequest{
		Params:      params,
		Headers:     headers,
		AuthUrl:     fmt.Sprint(paramsMap["auth_url"]),
		AuthMethod:  fmt.Sprint(paramsMap["auth_method"]),
		PrivateFlag: fmt.Sprint(paramsMap["private_flag"]),
		ErrorFlag:   "error_" + fmt.Sprint(paramsMap["error_flag"]),
	}
}

func sendRequestAuth(requestArgs ParamsRequest) (RequestAuthResponse, error) {
	var response RequestGatewayResponse

	body, responseRequest, err := _utils.SendRequest(requestArgs.AuthUrl, requestArgs.AuthMethod, requestArgs.Params, requestArgs.Headers)
	e := json.Unmarshal([]byte(body), &response.Body)
	_utils.ErrorHandling(e)

	if responseRequest == nil {
		response.Status = 500
	} else {
		response.Status = responseRequest.StatusCode
	}

	if err != nil && response.Status >= 500 {
		messageBody := make(map[string]interface{})

		messageBody["message"] = "Serviço indisponível, tente novamente em instantes."
		response.Body = messageBody

		data := response.Body

		return RequestAuthResponse{
			Status: http.StatusBadGateway,
			Body:   data,
		}, err
	} else if response.Status >= 400 {
		b := make(map[string]interface{})
		err = json.Unmarshal(body, &b)
		if err != nil {
			fmt.Println(err)
		}

		return RequestAuthResponse{
			Status: response.Status,
			Body:   b,
		}, err
	}

	return RequestAuthResponse{
		Status: http.StatusOK,
		Body:   response.Body,
	}, err
}

func init() {
	fmt.Println("\n[PRIVATE-AUTH-SERVER-RESPONSE] - " + Namespace + " handler plugin loaded! \n")
}

func main() {
}
