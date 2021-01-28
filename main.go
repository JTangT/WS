package main

import (
	"WebSocket_TCP/zlog"
	"encoding/json"
	"flag"
	"net"
	"io/ioutil"
	"io"
	"syscall"
	"os"
	"os/signal"
	"sync"
)

type Rule struct {
	Port string
	Address string
	ProxyProtocolVersion int
}

type Config struct {
	mu sync.RWMutex
	Mode string
	Rules map[string]Rule
}

var ConfigFile string
var Setting Config

func main(){
	flag.StringVar(&ConfigFile, "config", "config.json", "The config file location.")
	help := flag.Bool("h", false, "Show help")
    flag.Parse()

		if *help {
			flag.PrintDefaults()
			os.Exit(0)
		}


	conf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		zlog.Fatal("Cannot read the config file. (io Error) " + err.Error())
	}

	err = json.Unmarshal(conf, &Setting)
	if err != nil {
		zlog.Error("Cannot read the config file. (Parse Error) " + err.Error())
		return
	}

	if Setting.Mode == "Server" {
        go LoadServer()
	}else if Setting.Mode == "Client" {
        go LoadClient()
	}else{
		zlog.Fatal("Json Error: Unknow Mode")
		os.Exit(0)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()
	<-done
	zlog.PrintText("Exiting\n")
}

func LoadServer(){
	for index, _ := range Setting.Rules {
	    go LoadWSRules(index) 
	}
}

func LoadClient(){
	for index, _ := range Setting.Rules {
	    go LoadWSCRules(index)
	}
}

func net_copyIO(src,dest net.Conn){
defer src.Close()
defer dest.Close()
io.Copy(src,dest)
}