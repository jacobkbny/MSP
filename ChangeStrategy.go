package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func ChangeStrategy() {
	if Strategy == "NORMAL" {
		Strategy = "ABNORMAL"
		ScaleUp()
	} else {
		Strategy = "NORMAL"
		ChangeToNormal()
	}
	Data, err := json.Marshal(Strategy)
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ParseError")
		defer logFile.Close()
	}
	for _, V := range CurrentAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ConnectionError")
			defer logFile.Close()
		}
	}
	for _, V := range ReadyAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ConnectionError")
			defer logFile.Close()
		}
	}
	for _, V := range ZombieAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ConnectionError")
			defer logFile.Close()
		}
	}
}

func ChangeToNormal() {
	for K, V := range CurrentAddressTable {
		if len(ReadyAddressTable) <= 7 {
			ReadyAddressTable[K] = V
			delete(CurrentAddressTable, K)
		} else {
			Data, err := json.Marshal(CurrentAddressTable)
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ParseError")
				defer logFile.Close()
			}
			_, err = http.Post("http://"+GetMyIP()+":7000/UpdateHost", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ParseError")
				defer logFile.Close()
			}
			WriteHosts()
		}
	}
}
