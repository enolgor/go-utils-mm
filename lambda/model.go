package lambda

import "net/http"

type lambdaRequest struct {
	Body                  string            `json:"body"`
	IsBase64Encoded       bool              `json:"isBase64Encoded"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	RawPath               string            `json:"rawPath"`
	Headers               map[string]string `json:"headers"`
	Cookies               []string          `json:"cookies"`
	RequestContext        struct {
		Http struct {
			Method string `json:"method"`
		} `json:"http"`
	} `json:"requestContext"`
}

type lambdaResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	Cookies         []string          `json:"cookies"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

type Request struct {
	Method                string
	Body                  []byte
	QueryStringParameters map[string]string
	Path                  []string
	Headers               map[string]string
	Cookies               map[string]string
}

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Cookies    []*http.Cookie
}

type Handler func(request Request) (response Response)
