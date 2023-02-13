package main

import (
	"net"
)

var Client map[string]net.Conn

// put address and port for Connection
func TCPConnection(address string, port string) {
	client, err := net.Dial("tcp", address+":"+port)
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "error,"+err.Error())
		defer logFile.Close()
	}
	Client[address+":"+port] = client
}

// 4000, 4001 for differenciation I have to make Client is like Client map[string]net.Conn
// I made them as a map so I have to test it
func CloseConnection(Client map[string]net.Conn) {
	for _, V := range Client {
		V.Close()
	}
}
