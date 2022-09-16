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
	NewNode string `json:"node"`
	Type    string `json:"type"`
	Address string `json:"address"`
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
}

var CurrentAddressTable map[string]string
var ReadyAddressTable map[string]string
var ZombieAddressTable map[string]string
var PbftAddressTable map[string]string
var StatusOfAll []*Status
var MyPort string
var Threshold int
var ipBlackList map[int]string
var config ConfigData
var Hash [32]byte
var result bool

func init() {
	temp, err := LoadConfig()
	if err != nil {
		fmt.Println(err)
	}
	config = temp
	num, err := strconv.Atoi(config.Threshold)
	Threshold = num
	Hash = MakeHashofConfig(config)
}
func main() {
	Handlers()
	logFile, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("MSP Server Has Started" + time.Now().String())
	go func() {
		for {
			OpenConnection(config.DataBase, config.User.UserName, config.User.PassWord, config.User.DataBaseName)
			time.Sleep(6 * time.Hour)
		}

	}()
	go func() {
		for {
			CheckConfig()
			time.Sleep(3000 * time.Millisecond)
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
			log.Println(err)
		}
		Threshold = num
		log.Println("Threshold: ", Threshold)
		log.Println("Changes Has been spotted in config.json file")
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
	}, []byte{})
	return data
}
func Handlers() {
	addr := new(Addr)
	http.HandleFunc("/RegNewNode", addr.RegNewNode)
	http.HandleFunc("/PingReq", PingReq)
	http.HandleFunc("/ChangeStrategy", addr.ChangeStrategy)
	http.HandleFunc("/SendBlackIP", SendBlackIP)
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
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(K, V)
		if err != nil {
			log.Println(err)
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
		log.Println("IP BlackList has been Updated")
		ipBlackList = temp
		return true
	}
	return false

}
func TableUpdateAlarm() {
	//open the connection to the DB(Mysql) and Pull IP BlackList from the DB
	if result {
		Data, err := json.Marshal(ipBlackList)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, V := range CurrentAddressTable {
			_, err := http.Post("http://"+V+"/TableUpdateAlarm", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				fmt.Println(err)
			}
		}
		for _, V := range ReadyAddressTable {
			_, err := http.Post("http://"+V+"/TableUpdateAlarm", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				fmt.Println(err)
			}
		}
		for _, V := range ZombieAddressTable {
			_, err := http.Post("http://"+V+"/TableUpdateAlarm", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
func PingReq(w http.ResponseWriter, r *http.Request) {
	for Node, NodeAddress := range CurrentAddressTable {
		res, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
		if err != nil {
			addr := Addr{Node, "1", NodeAddress}
			log.Println("Node's Status right Before Disconnected", addr)
		}
		NodeStatus := new(Status)
		json.NewDecoder(res.Body).Decode(&NodeStatus)
		if NodeStatus.MemoryUsage >= Threshold {
			NodeStatus.Address = NodeAddress
			NodeStatus.GroupName = "Zombie"
			log.Println("Current Node To Zombie Node" + NodeStatus.Address)
			//NodeSwitching
			for K, V := range ReadyAddressTable {
				delete(ReadyAddressTable, K)
				CurrentAddressTable[K] = V
				NewHost, err := json.Marshal(CurrentAddressTable[K])
				if err != nil {
					fmt.Println(err)
				}
				//UpdateHost
				http.Post("http://localhost:7000/UpdateHost", "application/json", bytes.NewBuffer(NewHost))
				delete(CurrentAddressTable, Node)
				ZombieAddressTable[Node] = NodeAddress
				break
			}
		} else {
			NodeStatus.Address = NodeAddress
			NodeStatus.GroupName = "Current"
			log.Println(NodeStatus)
		}
		StatusOfAll = append(StatusOfAll, NodeStatus)
	}

	for Node, NodeAddress := range ReadyAddressTable {
		res, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
		if err != nil {
			addr := Addr{Node, "1", NodeAddress}
			log.Println("Node's Status right Before Disconnected", addr)
		}
		NodeStatus := new(Status)
		json.NewDecoder(res.Body).Decode(&NodeStatus)
		NodeStatus.Address = NodeAddress
		NodeStatus.GroupName = "Ready"
		log.Println(NodeStatus)
		StatusOfAll = append(StatusOfAll, NodeStatus)
	}

	for Node, NodeAddress := range ZombieAddressTable {
		res, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
		if err != nil {
			addr := Addr{Node, "2", NodeAddress}
			log.Println("Node's Status right Before Disconnected", addr)
		}
		NodeStatus := new(Status)
		json.NewDecoder(res.Body).Decode(&NodeStatus)
		if NodeStatus.MemoryUsage <= 20 {
			ReadyAddressTable[Node] = NodeAddress
			delete(ZombieAddressTable, Node)
			NodeStatus.Address = NodeAddress
			NodeStatus.GroupName = "Ready"
			log.Println("Zombie To Ready", NodeStatus)
			StatusOfAll = append(StatusOfAll, NodeStatus)
		} else {
			NodeStatus.Address = NodeAddress
			NodeStatus.GroupName = "Zombie"
			log.Println(NodeStatus)
			StatusOfAll = append(StatusOfAll, NodeStatus)
		}
		SendPingRes(StatusOfAll)
	}
}
func SendPingRes(StatusOfAll []*Status) {
	pid := os.Getpid()
	p, err := procfs.NewProc(pid)
	if err != nil {
		fmt.Println(err)
	}
	stat, err := p.Stat()
	if err != nil {
		fmt.Println(err)
	}
	MyStatus := Status{GetMyIP() + MyPort, "MSP", stat.ResidentMemory(), stat.ResidentMemory()}
	StatusOfAll = append(StatusOfAll, &MyStatus)
	log.Println(MyStatus)
}
func (addr *Addr) RegNewNode(w http.ResponseWriter, req *http.Request) {
	json.NewDecoder(req.Body).Decode(addr)
	ipBlackList := []string{}
	Json_Data, _ := json.Marshal(ipBlackList)
	json.NewEncoder(w).Encode(Json_Data)
	if addr.Type == "1" {
		if len(CurrentAddressTable) < 4 {
			CurrentAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
		} else {
			ReadyAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
		}
	} else {
		ZombieAddressTable[addr.NewNode] = addr.Address + ":" + addr.NewNode
	}
}

func (addr *Addr) ChangeStrategy(w http.ResponseWriter, req *http.Request) {
	Strategy := "ABNORMAL"
	Data, err := json.Marshal(Strategy)
	if err != nil {
		fmt.Println(err)
	}
	for _, V := range CurrentAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			fmt.Println(err)
		}
	}
	for _, V := range ReadyAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			fmt.Println(err)
		}
	}
	for _, V := range ZombieAddressTable {
		_, err = http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
		if err != nil {
			fmt.Println(err)
		}
	}
}
func CheckZombie() {
	if len(ZombieAddressTable) == 0 {
		Strategy := "NORMAL"
		Data, _ := json.Marshal(Strategy)
		for _, V := range CurrentAddressTable {
			_, err := http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				fmt.Println(err)
			}
		}
		for _, V := range ReadyAddressTable {
			_, err := http.Post("http://"+V+"/ChangeStrategy", "application/json", bytes.NewBuffer(Data))
			if err != nil {
				fmt.Println(err)
			}
		}
		log.Println("Changed Strategy to NORMAL")
	}
}

func SendBlackIP(w http.ResponseWriter, r *http.Request) {
	Data := new(ReqData)
	json.NewDecoder(r.Body).Decode(&Data)
	log.Println("SemiBlackIP", Data.Ip)

}

func NodeSwitching() {
	for Node, NodeAddress := range ZombieAddressTable {
		res, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
		if err != nil {
			fmt.Println(err)
		}
		NodeStatus := new(Status)
		Decoder := json.NewDecoder(res.Body)
		Decoder.Decode(&NodeStatus)
		if NodeStatus.MemoryUsage <= 20 {
			delete(ZombieAddressTable, Node)
			ReadyAddressTable[Node] = NodeAddress
			// tell zombie to change the type as "1" cuz it has "2" as a type
		}
	}
}
