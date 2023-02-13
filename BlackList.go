package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func OpenConnection(DataBase string, UserName string, PassWord string, DataBaseName string) {
	DB, err := sql.Open(DataBase, UserName+":"+PassWord+"@tcp(127.0.0.1:3306)/"+DataBaseName)
	if err != nil {
		fmt.Println(err)
	}
	defer DB.Close()
	var K int
	var V string
	var temp []string
	rows, err := DB.Query("SELECT * FROM IPBlackList")
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "error,"+err.Error())
		logFile.Close()
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(K, V)
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			logFile.Close()
		}
		temp[K] = V
	}
	if ipBlackList == nil {
		ipBlackList = temp
	} else {
		result = CompareIpBlackList(temp)
	}
	TableUpdateAlarm()
}
func CompareIpBlackList(temp []string) bool {
	if len(ipBlackList) != len(temp) {
		// WriteLog("IP BlackList has been Updated" + "\n")
		ipBlackList = temp
	}
	return true
}
func TableUpdateAlarm() {
	//open the connection to the DB(Mysql) and Pull IP BlackList from the DB
	if result {
		Data, err := json.Marshal(ipBlackList)
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			logFile.Close()
		}
		for _, V := range CurrentAddressTable {
			_, err := http.Post("http://"+V+"/TableUpdateAlarm", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "error,"+err.Error())
				logFile.Close()
			}
		}
		for _, V := range ReadyAddressTable {
			_, err := http.Post("http://"+V+"/TableUpdateAlarm", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "error,"+err.Error())
				logFile.Close()
			}
		}
		for _, V := range ZombieAddressTable {
			_, err := http.Post("http://"+V+"/TableUpdateAlarm", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "error,"+err.Error())
				logFile.Close()
			}
		}
	}
}
