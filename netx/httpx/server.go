package httpx

import (
	"net"
	"net/http"
)

type Server struct {
	http.Server // canonical go http server
	listener    net.Listener
}

func (srv *Server) Listener() net.Listener {
	return srv.listener
}

func (srv *Server) Serve(l net.Listener) error {
	srv.listener = l
	return srv.Server.Serve(l)
}

func (srv *Server) ServeTLS(l net.Listener, certFile string, keyFile string) error {
	srv.listener = l
	return srv.Server.ServeTLS(l, certFile, keyFile)
}

func (srv *Server) ListenAndServe() (err error) {
	addr := srv.Server.Addr
	if addr == "" {
		addr = ":http"
	}

	srv.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer srv.listener.Close()

	return srv.Server.Serve(srv.listener)
}

func (srv *Server) ListenAndServeTLS(certFile string, keyFile string) (err error) {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	srv.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer srv.listener.Close()

	return srv.Server.ServeTLS(srv.listener, certFile, keyFile)
}
