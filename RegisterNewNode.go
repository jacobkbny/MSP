package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func RegNewNode(w http.ResponseWriter, req *http.Request) {
	addr := new(Addr)
	json.NewDecoder(req.Body).Decode(addr)
	logFile := OpenLogFile("NewNode")
	WriteLog(logFile, "logging,newnode log,system-name,MSP,"+"nodeName,"+addr.NodeName+",type,"+addr.Type+",address,"+addr.Address+":"+addr.NewNode)
	w.Header().Set("Content-Type", "application/json")
	ipBlackList = append(ipBlackList, "abcd")
	json.NewEncoder(w).Encode(ipBlackList)
	port, err := strconv.Atoi(addr.NewNode)
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"Connection error")
		defer logFile.Close()
	}
	TCPConnection(addr.Address, fmt.Sprint(port+100))
	if addr.Type == "1" && Strategy == "NORMAL" {
		if len(CurrentAddressTable) <= 3 {
			fmt.Println("Current")
			CurrentAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
			fmt.Println(CurrentAddressTable)
			NodeNameTable[addr.Address+":"+addr.NewNode] = addr.NodeName
		} else {
			fmt.Println("Ready")
			ReadyAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
			NodeNameTable[addr.Address+":"+addr.NewNode] = addr.NodeName
		}
	} else if addr.Type == "2" {
		fmt.Println("Zombie")
		ZombieAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
		NodeNameTable[addr.Address+":"+addr.NewNode] = addr.NodeName
	} else if addr.Type == "1" && Strategy == "ABNORMAL" {
		fmt.Println("RegNewNode when Strategy is ABNORMAL")
		CurrentAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
		NodeNameTable[addr.Address+":"+addr.NewNode] = addr.NodeName
		Data, _ := json.Marshal(CurrentAddressTable)
		http.Post("http://"+GetMyIP()+":7000/UpdateHost", "application/json", bytes.NewBuffer(Data))
		WriteHosts()
	}
	if len(CurrentAddressTable) >= 4 && OneTime == 0 {
		Data, err := json.Marshal(CurrentAddressTable)
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"Connection error")
			defer logFile.Close()
		}
		http.Post("http://"+GetMyIP()+":7000/UpdateHost", "application/json", bytes.NewBuffer(Data))
		WriteHosts()
		OneTime++
	}
}
