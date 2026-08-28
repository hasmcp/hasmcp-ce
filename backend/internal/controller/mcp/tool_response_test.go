package mcp

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadToolResponseSuccess(t *testing.T) {
	response := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"answer":"ok"}`)),
	}

	body, isError := readToolResponse(response)
	if isError {
		t.Fatal("successful response was marked as an error")
	}
	if body != `{"answer":"ok"}` {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestReadToolResponsePreservesUpstreamError(t *testing.T) {
	response := &http.Response{
		Status:     "401 Unauthorized",
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"invalid API key"}}`)),
	}

	body, isError := readToolResponse(response)
	if !isError {
		t.Fatal("upstream authentication failure was not marked as an error")
	}
	if body != `{"error":{"message":"invalid API key"}}` {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestReadToolResponseUsesStatusForEmptyError(t *testing.T) {
	response := &http.Response{
		Status:     "403 Forbidden",
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	body, isError := readToolResponse(response)
	if !isError {
		t.Fatal("upstream authorization failure was not marked as an error")
	}
	if body != "upstream request failed: 403 Forbidden" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestReadToolResponseCapsErrorBody(t *testing.T) {
	response := &http.Response{
		Status:     "500 Internal Server Error",
		StatusCode: http.StatusInternalServerError,
		Body: io.NopCloser(strings.NewReader(
			strings.Repeat("x", _maxToolErrorResponseBodyBytes+100),
		)),
	}

	body, isError := readToolResponse(response)
	if !isError {
		t.Fatal("upstream failure was not marked as an error")
	}
	if !strings.HasSuffix(body, "\n[upstream error response truncated]") {
		t.Fatalf("truncation marker missing: %q", body[len(body)-64:])
	}
	if len(body) > _maxToolErrorResponseBodyBytes+64 {
		t.Fatalf("error body was not bounded: %d bytes", len(body))
	}
}

func TestReadToolResponseHandlesReadFailure(t *testing.T) {
	response := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(errorReader{}),
	}

	body, isError := readToolResponse(response)
	if !isError {
		t.Fatal("response read failure was not marked as an error")
	}
	if body != "failed to read the upstream response" {
		t.Fatalf("unexpected body: %q", body)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}
