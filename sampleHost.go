package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func WriteHosts() {
	jsonTest()
}

type HostInfo struct {
	Hostname string
	Hostip   string
}
type HostInfoJSON struct {
	Systemname      string
	Hostinformation []HostInfo
	Timestamp       string
}

func jsonTest() {
	//hostinfos := []HostInfo{}
	hostinfos := make([]HostInfo, len(CurrentAddressTable))
	// hostinfos[0].Hostname = "N1"
	// hostinfos[0].Hostip = "111.111.111.111"
	// hostinfos[1].Hostname = "N2"
	// hostinfos[1].Hostip = "111.111.111.112"
	var AllHostName []string
	var AllHostAddress []string
	AllHostName = make([]string, 0)
	AllHostAddress = make([]string, 0)
	for _, V := range CurrentAddressTable {
		AllHostName = append(AllHostName, NodeNameTable[V])
		AllHostAddress = append(AllHostAddress, V)
	}
	for i := 0; i < len(CurrentAddressTable); i++ {
		hostinfos[i].Hostname = AllHostName[i]
		hostinfos[i].Hostip = AllHostAddress[i]
	}
	hostinfo := &HostInfoJSON{
		Systemname:      "system",
		Hostinformation: hostinfos,
		Timestamp:       "YYYY-MM-DD",
	}
	byteData, _ := json.Marshal(hostinfo)
	// fmt.Println(string(byteData))
	logFolderPath := "./log"
	logFile, err := os.OpenFile(logFolderPath+"/"+time.Now().Format("2006-01-02")+"-Hosts.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	logFile.Write(byteData)
	defer logFile.Close()
}
