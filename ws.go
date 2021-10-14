package main

import (
	"github.com/CoiaPrant/zlog"
	proxyprotocol "github.com/pires/go-proxyproto"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"net/http"
)

type Addr struct {
	NetworkType   string
	NetworkString string
}

func (this *Addr) Network() string {
	return this.NetworkType
}

func (this *Addr) String() string {
	return this.NetworkString
}

func LoadWSRules(i string, rule Rule) {
	tcpaddress, _ := net.ResolveTCPAddr("tcp", ":"+rule.Port)
	ln, err := net.ListenTCP("tcp", tcpaddress)
	if err == nil {
		zlog.Info("Loaded [", i, "] (WebSocket)", rule.Port, " => ", rule.Address)
	} else {
		zlog.Error("Load failed [", i, "] (WebSocket) Error: ", err)
		return
	}

	Router := http.NewServeMux()
	Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		io.WriteString(w, Page404)
		return
	})
	Router.Handle("/ws/", websocket.Handler(func(ws *websocket.Conn) {
		WS_Handle(rule, ws)
	}))
	if rule.TLS {
		CreateTLSFile(certFile, keyFile)
		http.ServeTLS(ln, Router, certFile, keyFile)
	} else {
		http.Serve(ln, Router)
	}
}

func WS_Handle(rule Rule, ws *websocket.Conn) {
	conn, err := net.Dial("tcp", rule.Address)
	if err != nil {
		ws.Close()
		return
	}

	if rule.ProxyProtocolVersion != 0 {
		header := proxyprotocol.HeaderProxyFromAddrs(byte(rule.ProxyProtocolVersion), &Addr{
			NetworkType:   ws.Request().Header.Get("X-Forward-Protocol"),
			NetworkString: ws.Request().Header.Get("X-Forward-Address"),
		}, conn.LocalAddr())
		header.WriteTo(conn)
	}
	go copyIO(ws, conn)
	copyIO(conn, ws)
}
