package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/procfs"
)

type ReqData struct {
	Ip            string `json:"ip"`
	IsBad         string `json:"isBad"`
	Alias         string `json:"alias"`
	Txid          string `json:"txid"`
	WalletAddress string `json:"walletAddress"`
	URL_Path      string `json:"urlPath"`
	Tx            Transaction
}
type Status struct {
	Address     string `json:"address"`
	GroupName   string `json:"groupName"`
	MemoryUsage int    `json:"memoryUsage"`
	CpuUsage    int    `json:"cpuUsage"`
}
type Addr struct {
	NewNode  string `json:"node"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	NodeName string `json:"nodeName"`
}
type Transaction struct {
	TxID        [32]byte `json:"TxID"`
	TimeStamp   []byte   `json:"Timestamp"`   // 블럭 생성 시간
	Applier     []byte   `json:"Applier"`     // 신청자
	Company     []byte   `json:"Company"`     // 경력회사
	CareerStart []byte   `json:"CareerStart"` // 경력기간
	CareerEnd   []byte   `json:"CareerEnd"`
	Payment     []byte   `json:"Payment"` // 결제수단
	Job         []byte   `json:"Job"`     // 직종, 업무
	Proof       []byte   `json:"Proof"`   // 경력증명서 pdf
	WAddr       []byte   `json:"Address"` // 지갑 주소
	Sign        []byte   `json:"Sign"`
}
type ConfigData struct {
	Port     string `json:"port"`
	DataBase string `json:"dataBase"`
	User     struct {
		UserName     string `json:"userName"`
		PassWord     string `json:"passWord"`
		DataBaseName string `json:"dataBaseName"`
	}
	Threshold string `json:"threshold"`
	Revive    string `json:"revive"`
	PBFT      string `json:"pBFT"`
	Gateway   string `json:"gateway"`
}

var CurrentAddressTable map[string]string
var ReadyAddressTable map[string]string
var ZombieAddressTable map[string]string
var PbftHostAddressTable map[string]string
var PbftReadyAddressTable map[string]string
var StatusOfAll []*Status
var MyPort string
var Threshold int
var Revive int
var ipBlackList map[int]string
var config ConfigData
var Hash [32]byte
var result bool
var NodeNameTable map[string]string
var Strategy string
var Boot time.Time
var HostAddressTable map[string]string

func init() {
	temp, err := LoadConfig()
	if err != nil {
		fmt.Println(err)
	}
	config = temp
	num, err := strconv.Atoi(config.Threshold)
	digit, err := strconv.Atoi(config.Revive)
	Threshold = num
	Revive = digit
	Hash = MakeHashofConfig(config)
	Strategy = "NORMAL"
}
func Server() {
	Boot = time.Now()
	fmt.Println(GetMyIP())
	CurrentAddressTable = make(map[string]string)
	ReadyAddressTable = make(map[string]string)
	ZombieAddressTable = make(map[string]string)
	ipBlackList = make(map[int]string)
	NodeNameTable = make(map[string]string)
	PbftHostAddressTable = make(map[string]string)
	PbftReadyAddressTable = make(map[string]string)
	HostAddressTable = make(map[string]string)
	Handlers()
	// OpenLogFile()
	logFile := OpenLogFile("General")
	// WriteLog("\n" + "starttime:" + time.Now().Format("2006-01-02 15:04:05") + "\n")
	WriteLog(logFile, "starttime"+","+Boot.Format("2006-01-02 15:04:05"))
	// WriteLog("name:Athena" + "\n")
	WriteLog(logFile, "name,Athena")
	// WriteLog("strategy:" + Strategy + "\n")
	WriteLog(logFile, "strategy,"+Strategy)
	logFile.Close()
	// go func() {
	// 	for {
	// 		OpenConnection(config.DataBase, config.User.UserName, config.User.PassWord, config.User.DataBaseName)
	// 		time.Sleep(6 * time.Hour)
	// 	}
	// }()
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			// WriteLog("enlapsedTime:" + time.Since(Boot).String() + "\n")
			logFile := OpenLogFile("General")
			WriteLog(logFile, "enlapsedTime,"+time.Since(Boot).String())
			defer logFile.Close()
		}
	}()
	go func() {
		for {
			CheckConfig()
			time.Sleep(3000 * time.Millisecond)
		}
	}()
	go func() {
		for {
			// WriteLog("power:on" + "\n")
			logFile := OpenLogFile("General")
			WriteLog(logFile, "power,on")
			defer logFile.Close()
			time.Sleep(10000 * time.Millisecond)
		}
	}()
	go func() {
		for {
			if len(HostAddressTable) > 0 {
				time.Sleep(1 * time.Minute)
				CheckHost()
			}
		}
	}()
	go func() {
		for {
			if Strategy != "NORMAL" {
				CheckLazyBoyHost()
				CheckZombie()
				// WriteLog("strategy:" + Strategy + "\n")
				logFile := OpenLogFile("General")
				WriteLog(logFile, "strategy,"+Strategy)
			}
		}
	}()
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			PingReq()
		}
	}()
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
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
		num, err := strconv.Atoi(config.Threshold)
		if err != nil {
			// WriteLog("error: " + err.Error() + "\n")
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			logFile.Close()
		}
		digit, err := strconv.Atoi(config.Revive)
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			logFile.Close()
		}
		Threshold = num
		Revive = digit
		// WriteLog("Threshold:" + fmt.Sprint(Threshold) + "\n")
		logFile := OpenLogFile("Changes")
		WriteLog(logFile, "Threshold,"+fmt.Sprint(Threshold)+",Revive,"+fmt.Sprint(Revive))
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
func Handlers() {
	http.HandleFunc("/RegNewNode", RegNewNode)
	http.HandleFunc("/ChangeStrategy", ChangeStrategy)
	// http.HandleFunc("/SendBlackIP", SendBlackIP)
	http.HandleFunc("/GetLazyBoy", GetLazyBoy)
	http.HandleFunc("/Rest", Rest)
}
func GetMyIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// UserName , PassWord , DataBaseName
// Connection to the DataBase
func OpenConnection(DataBase string, UserName string, PassWord string, DataBaseName string) {
	DB, err := sql.Open(DataBase, UserName+":"+PassWord+"@tcp(127.0.0.1:3306)/"+DataBaseName)
	if err != nil {
		fmt.Println(err)
	}
	defer DB.Close()
	var K int
	var V string
	var temp map[int]string
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
func CompareIpBlackList(temp map[int]string) bool {
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
func PingReq() {
	for Node, NodeAddress := range CurrentAddressTable {
		GroupName := "Current"
		Json_Data, _ := json.Marshal(GroupName)
		res, err := http.Post("http://"+NodeAddress+"/PingReq", "applicaion/json", bytes.NewBuffer(Json_Data))
		if err != nil {
			delete(CurrentAddressTable, Node)
			if len(ReadyAddressTable) > 0 {
				for K, V := range ReadyAddressTable {
					CurrentAddressTable[K] = V
					delete(ReadyAddressTable, K)
					break
				}
			}
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			WriteLog(logFile, "disconnected,"+NodeAddress+","+"type,"+"1")
			defer logFile.Close()
			// WriteLog("Disconnected: " + NodeNameTable[Node] + NodeAddress + " " + "1" + "\n")
			// WriteLog("error: " + err.Error() + "\n")
			// CloseLogFile()
		} else {
			NodeStatus := new(Status)
			json.NewDecoder(res.Body).Decode(&NodeStatus)
			if NodeStatus.MemoryUsage >= Threshold {
				NodeStatus.Address = NodeAddress
				NodeStatus.GroupName = "Zombie"
				//NodeSwitching
				for K, V := range ReadyAddressTable {
					delete(ReadyAddressTable, K)
					CurrentAddressTable[K] = V
					HostAddressTable[K] = V
					NewHost, err := json.Marshal(CurrentAddressTable[K])
					if err != nil {
						logFile := OpenLogFile("Error")
						WriteLog(logFile, "error,"+err.Error())
						defer logFile.Close()
					}
					//UpdateHost
					_, err = http.Post("http://localhost:7000/UpdateHost", "application/json", bytes.NewBuffer(NewHost))
					logFile := OpenLogFile("Error")
					WriteLog(logFile, "error,"+err.Error())
					defer logFile.Close()
					delete(CurrentAddressTable, Node)
					delete(HostAddressTable, Node)
					ZombieAddressTable[Node] = NodeAddress
					break
				}
			} else {
				NodeStatus.Address = NodeAddress
				NodeStatus.GroupName = "Current"
			}
		}
	}
	for _, NodeAddress := range ReadyAddressTable {
		GroupName := "Ready"
		Json_Data, _ := json.Marshal(GroupName)
		res, err := http.Post("http://"+NodeAddress+"/PingReq", "application/json", bytes.NewBuffer(Json_Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			WriteLog(logFile, "disconnected,"+NodeAddress+","+"type,"+"1")
			defer logFile.Close()
		} else {
			NodeStatus := new(Status)
			json.NewDecoder(res.Body).Decode(&NodeStatus)
			NodeStatus.Address = NodeAddress
			NodeStatus.GroupName = "Ready"
		}
	}
	for Node, NodeAddress := range ZombieAddressTable {
		GroupName := "Zombie"
		Json_Data, _ := json.Marshal(GroupName)
		res, err := http.Post("http://"+NodeAddress+"/PingReq", "application/json", bytes.NewBuffer(Json_Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			WriteLog(logFile, "disconnected,"+NodeAddress+","+"type,"+"2")
			defer logFile.Close()
		} else {
			NodeStatus := new(Status)
			json.NewDecoder(res.Body).Decode(&NodeStatus)
			if NodeStatus.MemoryUsage <= 20 {
				ReadyAddressTable[Node] = NodeAddress
				delete(ZombieAddressTable, Node)
				NodeStatus.Address = NodeAddress
				NodeStatus.GroupName = "Ready"
			} else {
				NodeStatus.Address = NodeAddress
				NodeStatus.GroupName = "Zombie"
			}
		}
	}
	if len(PbftHostAddressTable) > 0 {
		for Node, NodeAddress := range PbftHostAddressTable {
			_, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "error,"+err.Error())
				WriteLog(logFile, "pBFTDisconnected,"+NodeAddress+","+"type,"+"3")
				defer logFile.Close()
				delete(PbftHostAddressTable, Node)
			}
		}
	}
	for Node, NodeAddress := range PbftReadyAddressTable {
		_, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			WriteLog(logFile, "pBFTDisconnected,"+NodeAddress+","+"type,"+"3")
			defer logFile.Close()
			delete(PbftReadyAddressTable, Node)
		}
	}
}
func SendPingRes(StatusOfAll []*Status) {
	pid := os.Getpid()
	p, err := procfs.NewProc(pid)
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "error,"+err.Error())
		defer logFile.Close()
	}
	stat, err := p.Stat()
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "error,"+err.Error())
		defer logFile.Close()
	}
	MyStatus := Status{GetMyIP() + MyPort, "MSP", stat.ResidentMemory(), stat.ResidentMemory()}
	StatusOfAll = append(StatusOfAll, &MyStatus)
}
func RegNewNode(w http.ResponseWriter, req *http.Request) {
	addr := new(Addr)
	json.NewDecoder(req.Body).Decode(addr)
	logFile := OpenLogFile("NewNode")
	WriteLog(logFile, "nodeName,"+addr.NodeName+",type,"+addr.Type+",address,"+addr.Address+":"+addr.NewNode)
	w.Header().Set("Content-Type", "application/json")
	ipBlackList := "abcd"
	json.NewEncoder(w).Encode(ipBlackList)
	if addr.Type == "1" {
		if len(CurrentAddressTable) < 4 {
			CurrentAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
			NodeNameTable[addr.NewNode] = addr.NodeName
		} else {
			ReadyAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
			NodeNameTable[addr.NewNode] = addr.NodeName
		}
	} else if addr.Type == "3" {
		PbftReadyAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
		NodeNameTable[addr.NewNode] = addr.NodeName
	} else {
		ZombieAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
		NodeNameTable[addr.NewNode] = addr.NodeName
	}
	if len(CurrentAddressTable) >= 3 && len(HostAddressTable) == 0 {
		var once sync.Once
		FirstHosts := func() {
			var Hosts []string
			for K, V := range CurrentAddressTable {
				HostAddressTable[K] = V
				Hosts = append(Hosts, V)
				if len(HostAddressTable) == 3 {
					// Data, _ := json.Marshal(Hosts)
					// http.Post("http://localhost:7000/UpdateHost", "application/json", bytes.NewBuffer(Data))
					break
				}
			}
			logFile := OpenLogFile("Hosts")
			WriteLog(logFile, "address,"+Hosts[0]+","+"address,"+Hosts[1]+","+"address,"+Hosts[2])
			defer logFile.Close()
		}
		once.Do(FirstHosts)
	}
}

//	func InitialHosts() {
//		var Hosts []string
//		for K, V := range CurrentAddressTable {
//			HostAddressTable[K] = V
//			Hosts = append(Hosts, V)
//			if len(HostAddressTable) == 3 {
//				// Data, _ := json.Marshal(Hosts)
//				// http.Post("http://localhost:7000/UpdateHost", "application/json", bytes.NewBuffer(Data))
//				break
//			}
//		}
//		WriteLog("Hosts:" + Hosts[0] + Hosts[1] + Hosts[2])
//	}
func CheckHost() {
	var temp []string
	for _, V := range HostAddressTable {
		temp = append(temp, V)
	}
	logFile := OpenLogFile("Hosts")
	WriteLog(logFile, "address,"+temp[0]+","+"address,"+temp[1]+","+"address,"+temp[2])
	defer logFile.Close()
}
func ChangeStrategy(w http.ResponseWriter, req *http.Request) {
	Strategy = "ABNORMAL"
	Data, err := json.Marshal(Strategy)
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "error,"+err.Error())
		defer logFile.Close()
	}
	for _, V := range CurrentAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			defer logFile.Close()
		}
	}
	for _, V := range ReadyAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			defer logFile.Close()
		}
	}
	for _, V := range ZombieAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			defer logFile.Close()
		}
	}
}
func CheckZombie() {
	result := 0
	Now := time.Now()
	for time.Since(Now).Milliseconds() <= 3000 {
		if len(ZombieAddressTable) > 0 {
			result++
		}
	}
	if result == 0 {
		Strategy = "NORMAL"
		logFile := OpenLogFile("General")
		WriteLog(logFile, "strategy,"+Strategy)
		defer logFile.Close()
		Data, _ := json.Marshal(Strategy)
		for _, V := range CurrentAddressTable {
			_, err := http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "error,"+err.Error())
				defer logFile.Close()
			}
		}
		for _, V := range ReadyAddressTable {
			_, err := http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				logFile := OpenLogFile("Error")
				WriteLog(logFile, "error,"+err.Error())
				defer logFile.Close()
			}
		}
	}
}

// Node -> MSP GetLazyBoy
// pbft -> MSP FinishDelay
// MSP -> Node DelayResult
var ADT int64
var Start time.Time

func GetLazyBoy(w http.ResponseWriter, r *http.Request) {
	logFile := OpenLogFile("Performance")
	WriteLog(logFile, "startDelay,"+time.Now().Format("2006-01-02 15:04:05"))
	defer logFile.Close()
	Start = time.Now()
	if len(PbftHostAddressTable) == 0 {
		for K, V := range PbftReadyAddressTable {
			PbftHostAddressTable[K] = V
			err := Awake(V)
			logFile := OpenLogFile("Error")
			WriteLog(logFile, "error,"+err.Error())
			defer logFile.Close()
			json.NewEncoder(w).Encode(V)
			break
		}
	} else {
		for _, V := range PbftHostAddressTable {
			json.NewEncoder(w).Encode(V)
		}
	}
}
func Awake(address string) error {
	_, err := http.Post("http://"+address+"/awake", "text/plain", nil)
	return err
}
func FinishDelay(w http.ResponseWriter, r *http.Request) {
	reqData := new(ReqData)
	json.NewDecoder(r.Body).Decode(&reqData)
	DelayResult(*reqData)
	ADT = time.Since(Start).Milliseconds()
	logFile := OpenLogFile("Performance")
	WriteLog(logFile, "ADT,"+fmt.Sprint(ADT))

}
func DelayResult(reqData ReqData) {
	Data, err := json.Marshal(reqData)
	if err != nil {
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "error,"+err.Error())
		defer logFile.Close()
	}
	for _, V := range HostAddressTable {
		_, err = http.Post("http://"+V+"/DelayResult", "application/json", bytes.NewBuffer(Data))
		logFile := OpenLogFile("Error")
		WriteLog(logFile, "error,"+err.Error())
		defer logFile.Close()
	}
}
func CheckLazyBoyHost() {
	if len(PbftHostAddressTable) > 0 {
		for _, V := range PbftHostAddressTable {
			logFile := OpenLogFile("Hosts")
			WriteLog(logFile, "lazyBoyHost,"+V)
			defer logFile.Close()
		}
	}
}
func Rest(w http.ResponseWriter, r *http.Request) {
	for Node := range PbftHostAddressTable {
		for K, V := range PbftReadyAddressTable {
			PbftHostAddressTable[K] = V
		}
		delete(PbftHostAddressTable, Node)
	}
}