package mockhttp

import (
	"net/http"
)

type Expectation struct {
	Route   string
	Handler http.HandlerFunc

	ExpectedMethod string
	ExpectedBody   []byte

	ExpectedVisits int
	ActualVisits   int
}

type MockOption func(*Expectation)

func WithExpectedVisits(visits int) MockOption {
	return func(e *Expectation) {
		e.ExpectedVisits = visits
	}
}

func WithMethod(method string) MockOption {
	return func(e *Expectation) {
		e.ExpectedMethod = method
	}
}

func WithBody(body []byte) MockOption {
	return func(e *Expectation) {
		e.ExpectedBody = body
	}
}

func WithHandler(handler http.HandlerFunc) MockOption {
	return func(e *Expectation) {
		e.Handler = handler
	}
}

func Expect(route string, opts ...MockOption) *Expectation {
	exp := &Expectation{
		Route: route,
	}

	for _, opt := range opts {
		opt(exp)
	}

	return exp
}
