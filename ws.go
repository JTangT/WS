package main

import (
	"WebSocket_TCP/zlog"
	"net/http"
	"net"
	"golang.org/x/net/websocket"
	"io"
	proxyprotocol "github.com/pires/go-proxyproto"
)

type Addr struct {
	NetworkType string
	NetworkString string
}

func (this *Addr) Network()string{
	return this.NetworkType
}

func (this *Addr) String()string{
	return this.NetworkString
}

func LoadWSRules(i string){
	Setting.mu.RLock()
	tcpaddress, _ := net.ResolveTCPAddr("tcp", ":"+Setting.Rules[i].Port)
	ln, err := net.ListenTCP("tcp", tcpaddress)
	if err == nil {
		zlog.Info("Loaded [",i,"] (WebSocket)", Setting.Rules[i].Port, " => ", Setting.Rules[i].Address)
	}else{
		zlog.Error("Load failed [",i,"] (Websocket) Error: ",err)
		Setting.mu.RUnlock()
		return
	}
	Setting.mu.RUnlock()

	http.HandleFunc("/",func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		io.WriteString(w,Page404)
		return
	})
	http.Handle("/ws/",websocket.Handler(func(ws *websocket.Conn){
		WS_Handle(i,ws)
	}))
	http.Serve(ln,nil)
}

func WS_Handle(i string , ws *websocket.Conn){
	ws.PayloadType = websocket.BinaryFrame
	Setting.mu.RLock()
	rule := Setting.Rules[i]
	Setting.mu.RUnlock()

   conn,err := net.Dial("tcp" , rule.Address)
   if err != nil {
	   ws.Close()
	   return
   }

   if rule.ProxyProtocolVersion != 0 {
	header := proxyprotocol.HeaderProxyFromAddrs(byte(rule.ProxyProtocolVersion),&Addr{
		NetworkType: ws.Request().Header.Get("X-Forward-Protocol"),
		NetworkString: ws.Request().Header.Get("X-Forward-Address"),
	},conn.LocalAddr())
	header.WriteTo(conn)
}
   go net_copyIO(ws,conn)
   net_copyIO(conn,ws)
}

