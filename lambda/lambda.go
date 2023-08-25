package lambda

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var DefaultHandler Handler = Echo

func Run() {
	lambda.Start(func(event lambdaRequest) (lambdaResponse, error) {
		return getResponse(DefaultHandler(getRequest(event)))
	})
}

func Err(err error, status int) (response Response) {
	response.Body = []byte(err.Error())
	response.StatusCode = status
	return response
}

func WithAuthentication(handler Handler, expiry time.Duration, password string) Handler {
	return func(request Request) Response {
		if request.Path[0] == "authenticate" {
			body := string(request.Body)
			if strings.Index(body, "pass=") != 0 {
				return authenticateResponse()
			}
			pass := body[5:]
			if password != pass {
				return authenticateResponse()
			}
			encr, err := encrypt(pass)
			if err != nil {
				return authenticateResponse()
			}
			return Response{
				StatusCode: 302,
				Headers: map[string]string{
					"Location": "/",
				},
				Cookies: []*http.Cookie{
					{Name: "auth", Value: encr, Path: "/", Expires: time.Now().Add(expiry)},
				},
			}
		}
		var auth string
		var ok bool
		if auth, ok = request.Cookies["auth"]; !ok {
			if auth, ok = request.Headers["x-auth"]; !ok {
				return authenticateResponse()
			}
		}
		pass, err := decrypt(auth)
		if err != nil {
			return authenticateResponse()
		}
		if pass != password {
			return authenticateResponse()
		}
		return handler(request)
	}
}

func getRequest(event lambdaRequest) (request Request) {
	if event.IsBase64Encoded {
		request.Body, _ = base64.StdEncoding.DecodeString(event.Body)
	} else {
		request.Body = []byte(event.Body)
	}
	request.Method = event.RequestContext.Http.Method
	request.Path = strings.Split(event.RawPath, "/")[1:]
	request.QueryStringParameters = event.QueryStringParameters
	request.Headers = event.Headers
	cookies := parseCookies(event.Cookies)
	request.Cookies = make(map[string]string)
	for i := range cookies {
		request.Cookies[cookies[i].Name] = cookies[i].Value
	}
	return
}

func GetRequest(event lambdaRequest) *http.Request {
	var body []byte
	if event.IsBase64Encoded {
		body, _ = base64.StdEncoding.DecodeString(event.Body)
	} else {
		body = []byte(event.Body)
	}
	req, _ := http.NewRequest(event.RequestContext.Http.Method, event.RawPath, bytes.NewBuffer(body))
	req.Header = make(http.Header)
	for k, v := range event.Headers {
		req.Header.Add(k, v)
	}
}

func getResponse(response Response) (lambdaResponse lambdaResponse, err error) {
	lambdaResponse.IsBase64Encoded = true
	lambdaResponse.Body = base64.StdEncoding.EncodeToString(response.Body)
	lambdaResponse.StatusCode = response.StatusCode
	lambdaResponse.Headers = response.Headers
	lambdaResponse.Cookies = stringifyCookies(response.Cookies)
	return
}

func parseCookies(cookies []string) []*http.Cookie {
	rawCookies := strings.Join(cookies, ";")
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	request := http.Request{Header: header}
	return request.Cookies()
}

func stringifyCookies(cookies []*http.Cookie) []string {
	cooks := make([]string, len(cookies))
	for i := range cookies {
		cooks[i] = cookies[i].String()
	}
	return cooks
}

func Echo(request Request) (response Response) {
	response.Body = request.Body
	response.StatusCode = 200
	return
}
