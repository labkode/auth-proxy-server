package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Configuration stores server configuration parameters
type Configuration struct {
	Port        int    `json:"port"`         // server port number
	Base        string `json:"base"`         // base URL
	Verbose     int    `json:"verbose"`      // verbose output
	ServerCrt   string `json:"server_cert"`  // path to server crt file
	ServerKey   string `json:"server_key"`   // path to server key file
	LogFile     string `json:"log_file"`     // log file
	HttpServer  bool   `json:"http_server"`  // run http service or not
	GRPCAddress string `json:"grpc_address"` // address of gRPC backend server
}

// Config variable represents configuration object
var Config Configuration

// helper function to parse configuration
func parseConfig(configFile string) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to read", err)
		return err
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		log.Println("Unable to parse", err)
		return err
	}
	return nil
}

func checkFile(fname string) string {
	_, err := os.Stat(fname)
	if err == nil {
		return fname
	}
	log.Fatalf("unable to read %s, error %v\n", fname, err)
	return ""
}

// our backend gRpc service
var backendGrpc GrpcService

// http server implementation
func grpcHttpServer() {
	// check if provided crt/key files exists
	serverCrt := checkFile(Config.ServerCrt)
	serverKey := checkFile(Config.ServerKey)

	// initialize gRPC remote backend
	var err error
	backendGrpc, err = NewGRPCService(Config.GRPCAddress)
	if err != nil {
		log.Fatal(err)
	}

	// the request handler
	http.HandleFunc(fmt.Sprintf("%s/", Config.Base), RequestHandler)

	// start HTTP or HTTPs server based on provided configuration
	addr := fmt.Sprintf(":%d", Config.Port)
	if serverCrt != "" && serverKey != "" {
		//start HTTPS server which require user certificates
		server := &http.Server{Addr: addr}
		log.Printf("Starting HTTPs server on %s", addr)
		log.Fatal(server.ListenAndServeTLS(serverCrt, serverKey))
	} else {
		log.Fatal("No server certificate files is provided")
	}
}

// grpc proxy server implementation
func grpcServer() {
	log.Fatal("Not implemented yet")
}