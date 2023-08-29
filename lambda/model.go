package lambda

type LambdaRequest struct {
	Body                  string            `json:"body"`
	IsBase64Encoded       bool              `json:"isBase64Encoded"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	RawPath               string            `json:"rawPath"`
	RawQueryString        string            `json:"rawQueryString"`
	Headers               map[string]string `json:"headers"`
	Cookies               []string          `json:"cookies"`
	RequestContext        struct {
		DomainName string `json:"domainName"`
		Http       struct {
			Method string `json:"method"`
		} `json:"http"`
	} `json:"requestContext"`
}

type LambdaResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	Cookies         []string          `json:"cookies"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}
