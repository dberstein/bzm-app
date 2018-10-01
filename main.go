package main

import (
	"crypto/tls"
	"flag"
	"log"
	"os"
	"runtime"
)

var (
	listenAddress   string
	backendURL      string
	certificatePath string
	keyPath         string
	logger          *log.Logger
)

func init() {
	logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	flag.StringVar(&listenAddress, "l", ":443", "listen address")
	flag.StringVar(&certificatePath, "c", "cert.pem", "SSL certificate path")
	flag.StringVar(&keyPath, "k", "key.pem", "SSL key path")
	flag.StringVar(&backendURL, "b", "http://127.0.0.1", "Backend URL")
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	_, err := tls.LoadX509KeyPair(certificatePath, keyPath)
	if err != nil {
		log.Fatalf("error in tls.LoadX509KeyPair: %s", err)
	}

	os.Stderr.WriteString("For \"" + backendURL + "\" listening at [" + listenAddress + "] with cert [" + certificatePath + "], key [" + keyPath + "]\n")

	app := newService(listenAddress, backendURL)
	err = app.Server().ListenAndServeTLS(certificatePath, keyPath)
	if err != nil {
		logger.Fatal(err)
	}
}
