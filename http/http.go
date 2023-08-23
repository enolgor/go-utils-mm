package http

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
	debug   io.Writer
}

func Get(url string) *Request {
	return &Request{"GET", url, map[string]string{}, nil, 0, nil, nil}
}

func Post(url string) *Request {
	return &Request{"POST", url, map[string]string{}, nil, 0, nil, nil}
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

func (r *Request) AddFormValue(key string, value string) *Request {
	if r.form == nil {
		r.form = &url.Values{}
	}
	r.form.Add(key, value)
	return r
}

func (r *Request) getBody() (string, io.ReadCloser, error) {
	if r.body == nil {
		return "", nil, nil
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

func (r *Request) Debug(debug io.Writer) *Request {
	r.debug = debug
	return r
}

func (r *Request) Do() (*http.Response, error) {
	if r.body != nil && r.form != nil {
		return nil, fmt.Errorf("can't specify body and form for the same request")
	}
	var body io.ReadCloser
	var err error
	var contentType string
	if r.body != nil {
		contentType, body, err = r.getBody()
		if err != nil {
			return nil, err
		}
		if _, ok := r.headers[ContentType]; !ok {
			r.headers[ContentType] = contentType
		}
	}
	req, err := http.NewRequest(r.method, r.url, body)
	if err != nil {
		return nil, err
	}
	if r.form != nil {
		req.PostForm = *r.form
		r.headers[ContentType] = "application/x-www-form-urlencoded"
	}
	for k, v := range r.headers {
		req.Header.Add(k, v)
	}
	if r.debug != nil {
		r.printDebug(req)
	}
	return http.DefaultClient.Do(req)
}

func (r *Request) printDebug(req *http.Request) {
	fmt.Fprintln(r.debug, "----PERFORMING REQUEST----")
	fmt.Fprintf(r.debug, "Host: %s\n", req.Host)
	fmt.Fprintf(r.debug, "Url: %s\n", req.URL.String())
	fmt.Fprintln(r.debug, "Headers:")
	for k, v := range req.Header {
		for i := range v {
			fmt.Fprintf(r.debug, "  %s: %s\n", k, v[i])
		}
	}
	fmt.Fprintln(r.debug, "Body:")
	req.Body = io.NopCloser(io.TeeReader(req.Body, r.debug))
}
