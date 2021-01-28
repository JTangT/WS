package main

import (
	"WebSocket_TCP/zlog"
	"net"
	"golang.org/x/net/websocket"
	"strings"
)

func LoadWSCRules(i string){
	Setting.mu.RLock()
	tcpaddress, _ := net.ResolveTCPAddr("tcp", ":"+Setting.Rules[i].Port)
	
	ln,err := net.ListenTCP("tcp",tcpaddress)
	if err == nil {
		zlog.Info("Loaded [",i,"] (WebSocket Client) ", Setting.Rules[i].Port, " => ", Setting.Rules[i].Address)
	}else{
		zlog.Error("Load failed [",i,"] (WebSocket Client) Error:",err)
		Setting.mu.RUnlock()
		return
	}

	Setting.mu.RUnlock()

	for{
		conn,err := ln.Accept()

		if err != nil {
            if err, ok := err.(net.Error); ok && err.Temporary() {
                continue
            }
			break
		}

		Setting.mu.RLock()
		rule := Setting.Rules[i]
		Setting.mu.RUnlock()

        go wsc_handleRequest(conn,rule)
	}
}

func wsc_handleRequest(conn net.Conn, r Rule) {
	ws_config,err := websocket.NewConfig("ws://" + r.Address + "/ws/","http://" + r.Address + "/ws/")
	if err != nil {
		_ = conn.Close()
		return
	}
	ws_config.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	ws_config.Header.Set("X-Forward-For",ParseAddrToIP(conn.RemoteAddr().String()))
	ws_config.Header.Set("X-Forward-Protocol",conn.RemoteAddr().Network())
	ws_config.Header.Set("X-Forward-Address",conn.RemoteAddr().String())
	proxy, err := websocket.DialConfig(ws_config)
	if err != nil {
		_ = conn.Close()
		return
	}
	proxy.PayloadType = websocket.BinaryFrame

	go net_copyIO(conn,proxy)
	go net_copyIO(proxy,conn)
}

func ParseAddrToIP(addr string) string {
	var str string
	arr :=strings.Split(addr,":")
        for i :=0;i< (len(arr) - 1);i++{
			if i!=0{
			str = str + ":" + arr[i]
			}else{
			str = str + arr[i]
			}
        }
    return str
}