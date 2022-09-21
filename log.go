package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kjk/dailyrotate"
)

func OnLogClose(path string, didRotate bool) {
	if !didRotate {
		return
	}
	go func() {

	}()
}

var LogFile *dailyrotate.File

func OpenLogFile(pathFormat string, onClose func(string, bool)) error {
	w, err := dailyrotate.NewFile(pathFormat, OnLogClose)
	if err != nil {
		return err
	}
	LogFile = w
	return nil
}

func CloseLogFile() error {
	return LogFile.Close()
}

func WriteLog(msg string) error {
	_, err := LogFile.Write([]byte(msg))
	CloseLogFile()
	return err
}

func Work() {
	logDir := "logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("os.MkdirAll()(")
	}
	pathFormat := filepath.Join(logDir, "2006-01-02"+"MSP"+".txt")
	err = OpenLogFile(pathFormat, OnLogClose)
	if err != nil {
		log.Fatalf("openLogFile failed with '%s'\n", err)
	}
	LogFile.Location, _ = time.LoadLocation("Local")
	if err != nil {
		log.Fatalf("writeToLog() failed with '%s'\n", err)
	}
	if err != nil {
		log.Fatalf("closeLogFile() failed with '%s'\n", err)
	}

}
