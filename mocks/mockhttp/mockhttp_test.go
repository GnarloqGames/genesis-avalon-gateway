package mockhttp

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

type TestRoute struct {
	route      string
	method     string
	body       []byte
	statusCode int
}

func TestMockHttp(t *testing.T) {
	tests := []struct {
		expectation *Expectation
		route       TestRoute
	}{
		{
			expectation: Expect("/",
				WithExpectedVisits(1),
				WithHandler(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("hi"))
				}),
			),
			route: TestRoute{
				route:      "/",
				statusCode: http.StatusOK,
			},
		},
		{
			expectation: Expect("/wrong-method",
				WithExpectedVisits(0),
				WithMethod(http.MethodPost),
				WithHandler(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("hi"))
				}),
			),
			route: TestRoute{
				route:      "/wrong-method",
				statusCode: http.StatusNotFound,
			},
		},
	}

	expectations := make([]*Expectation, 0)
	for _, tt := range tests {
		expectations = append(expectations, tt.expectation)
	}

	s, err := New(expectations...)
	require.NoError(t, err)
	defer s.Close()

	for _, tt := range tests {
		method := tt.route.method
		if method == "" {
			method = http.MethodGet
		}

		tf := func(t *testing.T) {
			req := resty.New().
				R().
				SetHeader("Content-Type", "application/json")

			if len(tt.route.body) > 0 {
				req = req.SetBody(tt.route.body)
			}

			resp, err := req.Execute(method, fmt.Sprintf("%s%s", s.URL(), tt.route.route))

			require.NoError(t, err)
			require.Equal(t, tt.route.statusCode, resp.StatusCode())
			require.Equal(t, tt.expectation.ExpectedVisits, tt.expectation.ActualVisits)
		}

		label := fmt.Sprintf("%s:%s", method, tt.route.route)
		t.Run(label, tf)
	}
}

func TestCreateListener(t *testing.T) {
	mux := http.NewServeMux()
	runningServer := &http.Server{
		Addr:    "127.0.0.1:61200",
		Handler: mux,
	}
	go runningServer.ListenAndServe()
	defer runningServer.Close()

	// needed to let time for the first listener to accept connections
	time.Sleep(500 * time.Millisecond)

	newServer := &Server{
		httpServer: &http.Server{},
	}
	newServer.createListener(61200)

	require.Equal(t, "http://127.0.0.1:61201", newServer.URL())
}
