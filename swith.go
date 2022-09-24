package main

import (
	"encoding/json"
	"net/http"
)

func SwtichStrategy(w http.ResponseWriter, r *http.Request) {
	var Data string
	json.NewDecoder(r.Body).Decode(&Data)
	if Data == "0" {
		Strategy = "NORMAL"
	}
	if Data == "1" {
		ChangeStrategy()
	}

}
