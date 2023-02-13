package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func PingReq() {
	for Node, NodeAddress := range CurrentAddressTable {
		// var Data []byte
		// Data = make([]byte, 0)
		temp := strings.Split(NodeAddress, ":")
		port, err := strconv.Atoi(temp[1])
		if Client[temp[0]+":"+fmt.Sprint(port+100)] != nil {
			err = json.NewEncoder(Client[temp[0]+":"+fmt.Sprint(port+100)]).Encode("Current")
			// if err exist
			if err != nil {
				logFile := OpenLogFile("Disconnection")
				WriteLog(logFile, "logging,disconnection log,system-name,MSP,node-name,"+NodeNameTable[NodeAddress]+",type,1")
				defer logFile.Close()
				Client[temp[0]+":"+fmt.Sprint(port+100)].Close()
				delete(CurrentAddressTable, Node)
				fmt.Println("delete CurrentAddressTable")
				delete(Client, temp[0]+":"+fmt.Sprint(port+100))
				fmt.Println("delete Client")
				if len(ReadyAddressTable) > 0 {
					for K, V := range ReadyAddressTable {
						NewNode := map[string]string{
							"newIp":    V,
							"zombieIp": NodeAddress,
						}
						address, err := json.Marshal(NewNode)
						if err != nil {
							logFile := OpenLogFile("Error")
							WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ParseError")
							defer logFile.Close()
						}
						_, err = http.Post("http://"+GetMyIP()+":7000/modifyHost", "application/json", bytes.NewBuffer(address))
						fmt.Println("Delete Current and Send Ready")
						if err != nil {
							logFile := OpenLogFile("Error")
							WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"Connection error")
							defer logFile.Close()
						}
						CurrentAddressTable[K] = V
						delete(ReadyAddressTable, K)
						WriteHosts()
						break
					}
				} else {
					delete(CurrentAddressTable, Node)
				}
			} else {
				response, err := bufio.NewReader(Client[temp[0]+":"+fmt.Sprint(port+100)]).ReadString('\n')
				if response != "" {
					temp := response[1 : len(response)-2]
					if err != nil {
						logFile := OpenLogFile("Error")
						WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"Connection error")
						defer logFile.Close()
					}
					// Memory := string(Data[1 : n-2])
					fmt.Println("CPU:", temp)
					MemoryUsage, err := strconv.ParseFloat(temp, 32)
					if err != nil {
						logFile := OpenLogFile("Error")
						WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ParseError")
						defer logFile.Close()
					}
					fmt.Printf("%s's Memory usage: %f\n", NodeNameTable[NodeAddress], MemoryUsage)
					if MemoryUsage >= Threshold {
						fmt.Println("Current to Zombie")
						if Strategy == "NORMAL" {
							ChangeStrategy()
						}
						var NewNode map[string]string
						NewNode = make(map[string]string)
						if len(ReadyAddressTable) == 0 {
							NewNode = map[string]string{
								"zombieIp": NodeAddress,
							}
							fmt.Println("Only Zombie no Ready", NodeAddress)
						}
						if len(ReadyAddressTable) > 0 {
							for _, V := range ReadyAddressTable {
								NewNode = map[string]string{
									"newIp":    V,
									"zombieIp": NodeAddress,
								}
								break
							}
							fmt.Println("Send Zombie And Ready", NodeAddress)
						}
						Data, err := json.Marshal(NewNode)
						res, err := http.Post("http://"+GetMyIP()+":7000/modifyHost", "application/json", bytes.NewBuffer(Data))
						if err != nil {
							logFile := OpenLogFile("Error")
							WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"Connection error")
							defer logFile.Close()
						}
						var ResFromGateway string
						json.NewDecoder(res.Body).Decode(&ResFromGateway)
						fmt.Println("ResFromGateway", ResFromGateway)
						delete(CurrentAddressTable, Node)
						ZombieAddressTable[Node] = NodeAddress
						WriteHosts()
					}
				}
			}
		}
	}
	for Node, NodeAddress := range ReadyAddressTable {
		temp := strings.Split(NodeAddress, ":")
		port, err := strconv.Atoi(temp[1])
		if Client[temp[0]+":"+fmt.Sprint(port+100)] != nil {
			err = json.NewEncoder(Client[temp[0]+":"+fmt.Sprint(port+100)]).Encode("Ready")
			var Data []byte
			if err == nil {
				_, err = Client[temp[0]+":"+fmt.Sprint(port+100)].Read(Data)
				if err != nil {
					logFile := OpenLogFile("Disconnection")
					WriteLog(logFile, "logging,disconnection log,system-name,MSP,node-name,"+NodeNameTable[NodeAddress]+",type,1")
					defer logFile.Close()
				}
			} else {
				logFile := OpenLogFile("Disconnection")
				WriteLog(logFile, "logging,disconnection log,system-name,MSP,node-name,"+NodeNameTable[NodeAddress]+",type,1")
				defer logFile.Close()
				Client[temp[0]+":"+fmt.Sprint(port+100)].Close()
				delete(ReadyAddressTable, Node)
				delete(Client, temp[0]+":"+fmt.Sprint(port+100))
			}
		}
	}
	for Node, NodeAddress := range ZombieAddressTable {
		temp := strings.Split(NodeAddress, ":")
		port, err := strconv.Atoi(temp[1])
		if Client[temp[0]+":"+fmt.Sprint(port+100)] != nil {
			err = json.NewEncoder(Client[temp[0]+":"+fmt.Sprint(port+100)]).Encode("Zombie")
			if err == nil {
				response, err := bufio.NewReader(Client[temp[0]+":"+fmt.Sprint(port+100)]).ReadString('\n')
				if response != "" {
					temp := response[1 : len(response)-2]
					if err != nil {
						logFile := OpenLogFile("Error")
						WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"Connection error")
						defer logFile.Close()
					}
					// Memory := string(Data[1 : n-2])
					fmt.Println("CPU:", temp)
					MemoryUsage, err := strconv.ParseFloat(temp, 32)
					if err != nil {
						logFile := OpenLogFile("Error")
						WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"Connection error")
						defer logFile.Close()
					}
					if MemoryUsage <= Revive {
						fmt.Println("Zombie to Ready")
						delete(ZombieAddressTable, Node)
						ReadyAddressTable[Node] = NodeAddress
						ScaleUp()
					}
				}
			}
		}
	}
}
