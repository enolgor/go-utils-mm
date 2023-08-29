package lambda

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/enolgor/go-utils/server"
)

// func WithAuthentication(handler Handler, expiry time.Duration, password string) Handler {
// 	return func(request Request) Response {
// 		if request.Path[0] == "authenticate" {
// 			body := string(request.Body)
// 			if strings.Index(body, "pass=") != 0 {
// 				return authenticateResponse()
// 			}
// 			pass := body[5:]
// 			if password != pass {
// 				return authenticateResponse()
// 			}
// 			encr, err := encrypt(pass)
// 			if err != nil {
// 				return authenticateResponse()
// 			}
// 			return Response{
// 				StatusCode: 302,
// 				Headers: map[string]string{
// 					"Location": "/",
// 				},
// 				Cookies: []*http.Cookie{
// 					{Name: "auth", Value: encr, Path: "/", Expires: time.Now().Add(expiry)},
// 				},
// 			}
// 		}
// 		var auth string
// 		var ok bool
// 		if auth, ok = request.Cookies["auth"]; !ok {
// 			if auth, ok = request.Headers["x-auth"]; !ok {
// 				return authenticateResponse()
// 			}
// 		}
// 		pass, err := decrypt(auth)
// 		if err != nil {
// 			return authenticateResponse()
// 		}
// 		if pass != password {
// 			return authenticateResponse()
// 		}
// 		return handler(request)
// 	}
// }

// var handlers map[string]func(w http.ResponseWriter, req *http.Request) = make(map[string]func(w http.ResponseWriter, req *http.Request))

// func HandleFunc(path string, f func(w http.ResponseWriter, req *http.Request)) {
// 	handlers[path] = f
// }

func Handler(router *server.Router) func(event *LambdaRequest) (*LambdaResponse, error) {
	return func(event *LambdaRequest) (*LambdaResponse, error) {
		rw := NewLambdaResponseWriter()
		req := GetRequest(event)
		router.ServeHTTP(rw, req)
		return rw.GetLambdaResponse(), nil
	}
}

func GetRequest(event *LambdaRequest) *http.Request {
	var body []byte
	if event.IsBase64Encoded {
		body, _ = base64.StdEncoding.DecodeString(event.Body)
	} else {
		body = []byte(event.Body)
	}
	url := fmt.Sprintf("https://%s%s?%s", event.RequestContext.DomainName, event.RawPath, event.RawQueryString)
	req, _ := http.NewRequest(event.RequestContext.Http.Method, url, bytes.NewBuffer(body))
	req.Header = make(http.Header)
	for k, v := range event.Headers {
		req.Header.Add(k, v)
	}
	if event.Cookies != nil && len(event.Cookies) > 0 {
		req.Header.Add("Cookie", strings.Join(event.Cookies, ";"))
	}
	return req
}

type LambdaResponseWriter struct {
	body   *bytes.Buffer
	header map[string][]string
	status int
}

func NewLambdaResponseWriter() *LambdaResponseWriter {
	return &LambdaResponseWriter{
		body:   &bytes.Buffer{},
		header: make(map[string][]string),
		status: 200,
	}
}

func (lrw *LambdaResponseWriter) Header() http.Header {
	return http.Header(lrw.header)
}

func (lrw *LambdaResponseWriter) Write(part []byte) (int, error) {
	return lrw.body.Write(part)
}

func (lrw *LambdaResponseWriter) WriteHeader(statusCode int) {
	lrw.status = statusCode
}

func (lrw LambdaResponseWriter) GetLambdaResponse() *LambdaResponse {
	lr := &LambdaResponse{
		IsBase64Encoded: false,
		StatusCode:      lrw.status,
	}
	lr.Headers = make(map[string]string)
	for k, v := range lrw.header {
		lr.Headers[k] = v[0]
	}
	//cookies
	if lrw.body.Len() != 0 {
		lr.IsBase64Encoded = true
		lr.Body = base64.StdEncoding.EncodeToString(lrw.body.Bytes())
	}
	return lr
}
