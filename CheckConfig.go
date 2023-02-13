package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

func LoadConfig() (ConfigData, error) {
	temp := new(ConfigData)
	file, err := os.Open("config.json")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&temp)
	if err != nil {
		log.Fatal(err)
	}
	return *temp, err
}
func CheckConfig() {
	temp, err := LoadConfig()
	if err != nil {
		fmt.Println(err)
	}
	tempHash := MakeHashofConfig(temp)
	if tempHash != Hash {
		config = temp
		Hash = tempHash
		num, err := strconv.ParseFloat(config.Threshold, 32)
		if err != nil {
			// WriteLog("error: " + err.Error() + "\n")
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ParseError")
			logFile.Close()
		}
		digit, err := strconv.ParseFloat(config.Revive, 32)
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "logging,error log,system-name,MSP,errmsg,"+"ParseError")
			logFile.Close()
		}
		Threshold = num
		Revive = digit
		// WriteLog("Threshold:" + fmt.Sprint(Threshold) + "\n")
		logFile := OpenLogFile("Changes")
		WriteLog(logFile, "logging,changes log,system-name,MSP,"+"threshold,"+fmt.Sprint(Threshold)+",revive,"+fmt.Sprint(Revive))
		defer logFile.Close()
	}
}
func MakeHashofConfig(config ConfigData) [32]byte {
	data := PrepareData(config)
	Hash := sha256.Sum256(data)
	return Hash
}
func PrepareData(config ConfigData) []byte {
	data := bytes.Join([][]byte{
		[]byte(config.DataBase),
		[]byte(config.Port),
		[]byte(config.Threshold),
		[]byte(config.User.DataBaseName),
		[]byte(config.User.PassWord),
		[]byte(config.User.UserName),
		[]byte(config.Revive),
		[]byte(config.PBFT),
		[]byte(config.Gateway),
	}, []byte{})
	return data
}
