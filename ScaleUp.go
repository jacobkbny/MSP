package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func ScaleUp() {
	if len(ReadyAddressTable) > 0 {
		start := time.Now()
		for K, V := range ReadyAddressTable {
			CurrentAddressTable[K] = V
		}
		data, err := json.Marshal(CurrentAddressTable)
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ConnectionError")
			defer logFile.Close()
		}
		http.Post("http://"+GetMyIP()+":7000/UpdateHost", "application/json", bytes.NewBuffer(data))
		SUT = time.Since(start).String()
		for k := range ReadyAddressTable {
			delete(ReadyAddressTable, k)
		}
		WriteHosts()
	}
}
