package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Data struct {
	Data string `json:"data"`
}

func SwtichStrategy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Switch Called")
	data := new(Data)
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
	}
	if data.Data == "0" {
		Strategy = "NORMAL"
		fmt.Println("Strategy", Strategy)
		var tempAddress map[string]string
		tempAddress = make(map[string]string)
		for V, K := range CurrentAddressTable {
			if len(tempAddress) <= 3 {
				tempAddress[V] = K
				break
			}
		}
		Data, _ := json.Marshal(tempAddress)
		_, err = http.Post("http://localhost:7000/UpdateHost", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "errormsg,"+err.Error())
		}
	}
	if data.Data == "1" {
		ChangeStrategy()
		ScaleUp()
		fmt.Println("ChangeStrategy")
	}

}
