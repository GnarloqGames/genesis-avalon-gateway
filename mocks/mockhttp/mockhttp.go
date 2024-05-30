package mockhttp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	url        string

	expectations map[string]*Expectation
}

func New(expectations ...*Expectation) (*Server, error) {
	server := &Server{
		expectations: make(map[string]*Expectation),
	}

	mux := http.NewServeMux()

	for i := range expectations {
		exp := expectations[i]
		server.expectations[exp.Route] = exp

		mux.HandleFunc(exp.Route, func(w http.ResponseWriter, r *http.Request) {
			if exp.ExpectedMethod != "" && exp.ExpectedMethod != r.Method {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			bodyBuf := bytes.NewBuffer(body)
			newBodyReadCloser := io.NopCloser(bodyBuf)
			r.Body = newBodyReadCloser

			if exp.ExpectedBody != nil && !bytes.Equal(exp.ExpectedBody, body) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			exp.Handler(w, r)

			exp.ActualVisits += 1
		})
	}

	var (
		s = &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		}
	)

	server.httpServer = s

	listener, err := server.createListener(61200)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := server.httpServer.Serve(listener); err != nil {
			slog.Debug("mock server produced an error", "error", err.Error())
		}
	}()

	return server, nil
}

func (s *Server) createListener(startingPort uint16) (net.Listener, error) {
	var (
		port     = startingPort
		maxPort  = ^uint16(0) - 1
		listener net.Listener
	)

	for {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		slog.Debug("creating listener", "address", addr)

		if l, err := net.Listen("tcp", addr); err == nil {
			tmpServer := &http.Server{
				Addr:              addr,
				ReadHeaderTimeout: 1 * time.Second,
			}

			var serverErr error

			go func() {
				if err := tmpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Println(serverErr)
					serverErr = err
				}
			}()
			time.Sleep(100 * time.Millisecond)

			if serverErr != nil {
				tmpServer.Close() //nolint

				listener = l
				s.url = fmt.Sprintf("http://%s", addr)
				s.httpServer.Addr = addr

				break
			}
		}

		slog.Debug("failed to create listener", "address", addr)

		port += 1
		if port > maxPort {
			slog.Debug("exhausted ports, this should not happen")
			return nil, fmt.Errorf("exhausted ports")
		}
	}

	return listener, nil
}

func (s *Server) Close() error {
	return s.httpServer.Close()
}

func (s *Server) URL() string {
	return s.url
}
