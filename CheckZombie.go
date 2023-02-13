package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func CheckZombie() {
	result := 0
	Now := time.Now()
	for time.Since(Now).Minutes() <= 5 {
		if len(ZombieAddressTable) > 0 {
			result++
		}
	}
	if result == 0 {
		fmt.Println("Change the Strategy Since there was no Zombie node for 5 mins")
		Strategy = "NORMAL"
		ChangeToNormal()
		Data, _ := json.Marshal(Strategy)
		for _, V := range CurrentAddressTable {
			_, err := http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ConnectionError")
				defer logFile.Close()
			}
		}
		for _, V := range ReadyAddressTable {
			_, err := http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ConnectionError")
				defer logFile.Close()
			}
		}
	}
}
