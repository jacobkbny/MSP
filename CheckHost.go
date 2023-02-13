package main

func CheckHost() {
	if len(CurrentAddressTable) >= 4 {
		var temp []string
		for _, V := range CurrentAddressTable {
			temp = append(temp, V)
		}
		var AllHostAddress string
		// logFile := OpenLogFile("Hosts")
		for i := 0; i < len(temp); i++ {
			// WriteLog(logFile, NodeNameTable[temp[i]]+","+temp[i])
			// log.Println("hosts:")
			// log.Println(temp[i] + ",")
			// defer logFile.Close()
			if i <= len(temp)-2 {
				AllHostAddress += temp[i] + "-"
			} else {
				AllHostAddress += temp[i]
			}
		}
		var AllHostName string
		for i := 0; i < len(temp); i++ {
			if i <= len(temp)-2 {
				AllHostName += NodeNameTable[temp[i]] + "-"
			} else {
				AllHostName += NodeNameTable[temp[i]]
			}
		}
		logFile := OpenLogFile("Hosts")
		WriteLog(logFile, "HostNames,"+AllHostName+","+"HostAddressTable,"+AllHostAddress)
		defer logFile.Close()
	}
}
