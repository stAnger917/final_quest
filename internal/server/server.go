package server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	HTTPServer *http.Server
}

func InitNewServer(address string, handler http.Handler) *Server {
	return &Server{
		HTTPServer: &http.Server{
			Addr:         address,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	return s.HTTPServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.HTTPServer.Shutdown(ctx)
}
