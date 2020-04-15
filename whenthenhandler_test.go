package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func NewMockWriter(nested io.Writer) mockResponseWriter {
	return mockResponseWriter{
		header: make(http.Header),
		status: new(int),
		Body:   nested,
	}
}

type mockResponseWriter struct {
	header http.Header
	status *int
	Body   io.Writer
}

func (m mockResponseWriter) Header() http.Header {
	return m.header
}

func (m mockResponseWriter) Write(a []byte) (int, error) {
	return m.Body.Write(a)
}

func (m mockResponseWriter) WriteHeader(statusCode int) {
	*m.status = statusCode
}

func TestWriteThen(t *testing.T) {
	actual := bytes.Buffer{}
	writer := NewMockWriter(&actual)

	writeThen(writer, &Then{
		Status: 100,
		Delay:  200,
		Headers: Header{
			"some-data": []string{"some", "values"},
		},
		Body: `{"content":"a"}`,
	})

	assert.Equal(t, 100, *writer.status)
	assert.Equal(t, "some", writer.header.Get("some-data"))
	assert.ElementsMatch(t, []string{"some", "values"}, writer.header["Some-Data"])
	assert.ElementsMatch(t, actual.Bytes(), []byte(`{"content":"a"}`))
}

func TestGetAddingFunc(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/path",
		},
		Method: http.MethodGet,
		Body: ioutil.NopCloser(strings.NewReader(`[{"When":{"method":"get", "uri": "/path"}, "Then": {"status": 200}}]`)),
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getAddingFunc(MockSuccessStorage{})(writer, req)

	assert.Equal(t, *writer.status, 201)
}

func TestGetAddingFunc_JsonMalformed(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/path",
		},
		Method: http.MethodGet,
		Body: ioutil.NopCloser(strings.NewReader(`hen":{"method":"get", "uri": "/path"}, "Then": {"status": 200}}]`)),
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getAddingFunc(MockSuccessStorage{})(writer, req)

	assert.Equal(t, *writer.status, 500)
}


func TestGetAddingFunc_StorageError(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/path",
		},
		Method: http.MethodGet,
		Body: ioutil.NopCloser(strings.NewReader(`[{"When":{"method":"get", "uri": "/path"}, "Then": {"status": 200}}]`)),
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getAddingFunc(MockFailStorage{})(writer, req)

	assert.Equal(t, *writer.status, 500)
}

func TestGetHandleFunc(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/path",
		},
		Method: http.MethodGet,
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getHandleFunc(MockSuccessStorage{})(writer, req)

	assert.Equal(t, *writer.status, 201)
}

func TestGetHandleFunc_Body(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/any",
		},
		Method: http.MethodGet,
		Body: ioutil.NopCloser(strings.NewReader(`[{"When":{"method":"any", "uri": "/any"}, "Then": {"status": 200}}]`)),
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getHandleFunc(MockSuccessStorage{})(writer, req)

	assert.Equal(t, *writer.status, 201)
}

func TestGetHandleFunc_NoMatch(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/any",
		},
		Method: http.MethodGet,
		Body: ioutil.NopCloser(strings.NewReader(`any`)),
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getHandleFunc(MockFailStorage{Err: NOT_FOUND})(writer, req)

	assert.Equal(t, *writer.status, 404)
}


func TestGetHandleFunc_WrappedErrorNoMatch(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/any",
		},
		Method: http.MethodGet,
		Body: ioutil.NopCloser(strings.NewReader(`any`)),
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getHandleFunc(MockFailStorage{Err: fmt.Errorf("metadata %w", NOT_FOUND)})(writer, req)

	assert.Equal(t, *writer.status, 404)
}

func TestGetHandleFunc_UnknownError(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Path:       "/any",
		},
		Method: http.MethodGet,
		Body: ioutil.NopCloser(strings.NewReader(`any`)),
	}

	buffer := &bytes.Buffer{}
	writer := NewMockWriter(buffer)

	getHandleFunc(MockFailStorage{})(writer, req)

	assert.Equal(t, *writer.status, 500)
}
