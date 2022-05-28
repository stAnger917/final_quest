package server

import (
	"context"
	"net/http"
)

type Server struct {
	HttpServer *http.Server
}

func InitNewServer(address string, handler http.Handler) *Server {
	return &Server{
		HttpServer: &http.Server{
			Addr:    address,
			Handler: handler,
		},
	}
}

func (s *Server) Run() error {
	return s.HttpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.HttpServer.Shutdown(ctx)
}
