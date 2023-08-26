package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	method  string
	url     string
	headers map[string]string
	body    any
	length  int64
	form    *url.Values
}

func Get(url string) *Request {
	return &Request{"GET", url, map[string]string{}, nil, 0, nil}
}

func Post(url string) *Request {
	return &Request{"POST", url, map[string]string{}, nil, 0, nil}
}

func (r *Request) WithHeader(key, value string) *Request {
	r.headers[NormalizeHeader(key)] = value
	return r
}

func (r *Request) WithHeaders(headers map[string]string) *Request {
	for k, v := range headers {
		r.WithHeader(k, v)
	}
	return r
}

func (r *Request) WithBody(body any) *Request {
	if r.method == "GET" {
		return r
	}
	r.body = body
	return r
}

func (r *Request) WithFormValue(key string, value string) *Request {
	if r.form == nil {
		r.form = &url.Values{}
	}
	r.form.Add(key, value)
	return r
}

func (r *Request) getBody() (string, io.ReadCloser, error) {
	if r.body == nil && r.form == nil {
		return "", nil, nil
	}
	if r.body != nil && r.form != nil {
		return "", nil, fmt.Errorf("can't specify body and form for the same request")
	}
	if r.form != nil {
		return "application/x-www-form-urlencoded", io.NopCloser(strings.NewReader(r.form.Encode())), nil
	}
	var contentType string
	var reader io.ReadCloser
	switch r.body.(type) {
	case string:
		contentType = "text/plain"
	case *string:
		contentType = "text/plain"
	default:
		contentType = "application/json"
	}
	switch v := r.body.(type) {
	case string:
		reader = io.NopCloser(strings.NewReader(v))
	case *string:
		reader = io.NopCloser(strings.NewReader(*v))
	case io.Reader:
		reader = io.NopCloser(v)
	default:
		buffer := new(bytes.Buffer)
		enc := json.NewEncoder(buffer)
		if err := enc.Encode(v); err != nil {
			return "", nil, err
		}
		reader = io.NopCloser(buffer)
	}
	return contentType, reader, nil
}

func (r *Request) Do() (*http.Response, error) {
	if r.body != nil && r.form != nil {
		return nil, fmt.Errorf("can't specify body and form for the same request")
	}
	var body io.ReadCloser
	var err error
	var contentType string
	contentType, body, err = r.getBody()
	if err != nil {
		return nil, err
	}
	if _, ok := r.headers[ContentType]; !ok {
		r.headers[ContentType] = contentType
	}
	req, err := http.NewRequest(r.method, r.url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range r.headers {
		req.Header.Add(k, v)
	}
	return http.DefaultClient.Do(req)
}
