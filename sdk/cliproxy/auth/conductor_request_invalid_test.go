package auth

import (
	"net/http"
	"testing"
)

type testStatusError struct {
	status int
	msg    string
}

func (e testStatusError) Error() string {
	return e.msg
}

func (e testStatusError) StatusCode() int {
	return e.status
}

func TestIsRequestInvalidError_BadRequestUsageLimitIsRetriable(t *testing.T) {
	err := testStatusError{
		status: http.StatusBadRequest,
		msg:    `{"type":"invalid_request_error","message":"You've hit your usage limit."}`,
	}
	if isRequestInvalidError(err) {
		t.Fatalf("expected usage-limit error to be retriable (isRequestInvalidError=false), got true")
	}
}

func TestIsRequestInvalidError_BadRequestInvalidRequestIsNonRetriable(t *testing.T) {
	err := testStatusError{
		status: http.StatusBadRequest,
		msg:    `{"type":"invalid_request_error","message":"Missing required parameter: input"}`,
	}
	if !isRequestInvalidError(err) {
		t.Fatalf("expected invalid_request_error to be non-retriable (isRequestInvalidError=true), got false")
	}
}
