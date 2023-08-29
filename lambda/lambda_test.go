package lambda

import (
	"io"
	"testing"
)

var exampleEvent = LambdaRequest{
	Body:                  "some-body",
	IsBase64Encoded:       false,
	QueryStringParameters: map[string]string{"a": "b", "c": "d"},
	RawPath:               "/the/path",
	RawQueryString:        "?a=b&c=d",
	Headers:               map[string]string{"Accept": "application/json"},
	Cookies:               []string{"cook1=val1", "cook2=val2"},
	RequestContext: struct {
		DomainName string `json:"domainName"`
		Http       struct {
			Method string `json:"method"`
		} `json:"http"`
	}{
		DomainName: "some.domain.com",
		Http: struct {
			Method string `json:"method"`
		}{"GET"},
	},
}

func TestLambda(t *testing.T) {
	req := GetRequest(&exampleEvent)

	expectedUrl := "https://some.domain.com/the/path?a=b&c=d"
	url := req.URL.String()
	if expectedUrl != url {
		t.Errorf("expected: %s, got: %s", expectedUrl, url)
	}

	expectedBody := "some-body"
	body, _ := io.ReadAll(req.Body)
	if expectedBody != string(body) {
		t.Errorf("expected: %s, got: %s", expectedBody, body)
	}
}
