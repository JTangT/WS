package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"github.com/CoiaPrant/zlog"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Rule struct {
	TLS                  bool
	Port                 string
	Address              string
	ProxyProtocolVersion int
}

type Config struct {
	mu    sync.RWMutex
	Mode  string
	Rules map[string]Rule
}

var ConfigFile string
var Setting Config
var certFile string
var keyFile string

func main() {
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
	} else if Setting.Mode == "Client" {
		go LoadClient()
	} else {
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

func LoadServer() {
	for index, _ := range Setting.Rules {
		go LoadWSRules(index)
	}
}

func LoadClient() {
	for index, _ := range Setting.Rules {
		go LoadWSCRules(index)
	}
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}

func sendRequest(url string, body io.Reader, addHeaders map[string]string, method string) (statuscode int, resp []byte, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")

	if len(addHeaders) > 0 {
		for k, v := range addHeaders {
			req.Header.Add(k, v)
		}
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	statuscode = response.StatusCode
	resp, err = ioutil.ReadAll(response.Body)
	return
}

func CreateTLSFile(certFile, keyFile string) {
	var ip string
	os.Remove(certFile)
	os.Remove(keyFile)
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)
	subject := pkix.Name{
		Country:            []string{"US"},
		Province:           []string{"WDC"},
		Organization:       []string{"Microsoft Corporation"},
		OrganizationalUnit: []string{"Microsoft Corporation"},
		CommonName:         "www.microstft.com",
	}

	_, resp, err := sendRequest("https://api.ip.sb/ip", nil, nil, "GET")
	if err == nil {
		ip = string(resp)
	} else {
		ip = "127.0.0.1"
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP(ip)},
	}

	pk, _ := rsa.GenerateKey(rand.Reader, 2048)

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	certOut, _ := os.Create(certFile)
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, _ := os.Create(keyFile)
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	keyOut.Close()
	zlog.Success("Created the ssl certfile,location: ", certFile)
	zlog.Success("Created the ssl keyfile,location: ", keyFile)
}
