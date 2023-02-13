package main

import (
	_ "bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
	MemoryUsage int    `json:"mem"`
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

// var PbftHostAddressTable map[string]string
// var PbftReadyAddressTable map[string]string
var StatusOfAll []*Status
var MyPort string
var Threshold float64
var Revive float64
var ipBlackList []string
var config ConfigData
var Hash [32]byte
var result bool

// Key = address , value = NodeName
var NodeNameTable map[string]string
var Strategy string
var Boot time.Time
var SUT string
var OneTime int

func init() {
	temp, err := LoadConfig()
	if err != nil {
		fmt.Println(err)
	}
	config = temp
	num, err := strconv.ParseFloat(config.Threshold, 32)
	digit, err := strconv.ParseFloat(config.Revive, 32)
	Threshold = num
	Revive = digit
	Hash = MakeHashofConfig(config)
	Strategy = "NORMAL"
	OneTime = 0
}
func Server() {
	Boot = time.Now()
	fmt.Println(GetMyIP())
	fmt.Println("MSP is running")
	CurrentAddressTable = make(map[string]string)
	ReadyAddressTable = make(map[string]string)
	ZombieAddressTable = make(map[string]string)
	ipBlackList = make([]string, 0)
	NodeNameTable = make(map[string]string)
	Client = make(map[string]net.Conn)
	// PbftHostAddressTable = make(map[string]string)
	// PbftReadyAddressTable = make(map[string]string)
	Handlers()
	// OpenLogFile()
	logFile := OpenLogFile("General")
	WriteLog(logFile, "logging,General log,systemname,MSP"+",starttime,"+Boot.Format("2006-01-02 15:04:05")+","+"name,Atena,"+"power,on,"+"strategy,"+Strategy+","+"enlapsedTime,"+time.Since(Boot).String())
	logFile.Close()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			// var Total float64
			Total := float64(len(CurrentAddressTable) + len(ReadyAddressTable) + len(ZombieAddressTable))
			CurrentLen := float64(len(CurrentAddressTable))
			ZombieLen := float64(len(ZombieAddressTable))
			if Total > 0 {
				temp := fmt.Sprint((1.0 / CurrentLen) * 100.0)
				PPS := strings.Split(temp, ".")
				Haza := fmt.Sprint((ZombieLen / Total) * 100.0)
				Hazardeous := strings.Split(Haza, ".")
				logFile := OpenLogFile("Performance")
				WriteLog(logFile, "ppn,"+PPS[0]+",sut,"+SUT+",hazardeous,"+Hazardeous[0]+",totalNode,"+fmt.Sprint(Total))
				defer logFile.Close()
			}
		}
	}()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			logFile := OpenLogFile("General")
			WriteLog(logFile, "starttime,"+Boot.Format("2006-01-02 15:04:05")+","+"name,Atena,"+"power,on,"+"strategy,"+Strategy+","+"enlapsedTime,"+time.Since(Boot).String())
			defer logFile.Close()
		}
	}()
	//check config file if any changes are made
	go func() {
		for {
			CheckConfig()
			time.Sleep(3000 * time.Millisecond)
		}
	}()
	// go func() {
	// 	for {
	// 		// WriteLog("power:on" + "\n")
	// 		logFile := OpenLogFile("General")
	// 		WriteLog(logFile, "starttime,"+Boot.Format("2006-01-02 15:04:05")+","+"name,Atena,"+"power,on,"+"strategy,"+Strategy+","+"enlapsedTime,"+time.Since(Boot).String())
	// 		defer logFile.Close()
	// 		time.Sleep(10000 * time.Millisecond)
	// 	}
	// }()
	// Check who is working as a host in the GateWay
	go func() {
		for {
			if len(CurrentAddressTable) >= 4 {
				WriteHosts()
				time.Sleep(1 * time.Minute)
			}
		}
	}()
	go func() {
		for {
			//Check if the attack is over by checking the length of ZombieAddressTable for 5 min
			if Strategy != "NORMAL" {
				//CheckLazyBoyHost()
				CheckZombie()
				// WriteLog("strategy:" + Strategy + "\n")
				// logFile := OpenLogFile("General")
				// WriteLog(logFile, "starttime,"+Boot.Format("2006-01-02 15:04:05")+","+"name,Atena,"+"power,on,"+"strategy,"+Strategy+","+"enlapsedTime,"+time.Since(Boot).String())
			}
		}
	}()
	// put the Node in Ready to Current when strategy is not 'NORMAL'
	// go func() {
	// 	for {
	// 		if Strategy != "NORMAL" && len(ReadyAddressTable) >0 {
	// 			time.Sleep(1 * time.Minute)
	// 			ScaleUp()
	// 		}
	// 	}
	// }()
	go func() {
		for {
			Total := len(CurrentAddressTable) + len(ReadyAddressTable) + len(ZombieAddressTable)
			if Total > 0 {
				time.Sleep(500 * time.Millisecond)
				PingReq()
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
	defer CloseConnection(Client)

}

func Handlers() {
	http.HandleFunc("/RegNewNode", RegNewNode)
	http.HandleFunc("/Switch", SwtichStrategy)
	// http.HandleFunc("/SendBlackIP", SendBlackIP)
	// http.HandleFunc("/GetLazyBoy", GetLazyBoy)
	// http.HandleFunc("/Rest", Rest)
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

//	if len(PbftHostAddressTable) > 0 {
//		for Node, NodeAddress := range PbftHostAddressTable {
//			_, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
//			if err != nil {
//				logFile := OpenLogFile("Error")
//				WriteLog(logFile, "error,"+err.Error())
//				WriteLog(logFile, "pBFTDisconnected,"+NodeAddress+","+"type,"+"3")
//				defer logFile.Close()
//				delete(PbftHostAddressTable, Node)
//			}
//		}
//	}
//
//	for Node, NodeAddress := range PbftReadyAddressTable {
//		_, err := http.Post("http://"+NodeAddress+"/PingReq", "text/plain", nil)
//		if err != nil {
//			logFile := OpenLogFile("Error")
//			WriteLog(logFile, "error,"+err.Error())
//			WriteLog(logFile, "pBFTDisconnected,"+NodeAddress+","+"type,"+"3")
//			defer logFile.Close()
//			delete(PbftReadyAddressTable, Node)
//		}
//	}
//
// }
//
//	func SendPingRes(StatusOfAll []*Status) {
//		pid := os.Getpid()
//		p, err := procfs.NewProc(pid)
//		if err != nil {
//			logFile := OpenLogFile("Error")
//			WriteLog(logFile, "error,"+err.Error())
//			defer logFile.Close()
//		}
//		stat, err := p.Stat()
//		if err != nil {
//			logFile := OpenLogFile("Error")
//			WriteLog(logFile, "error,"+err.Error())
//			defer logFile.Close()
//		}
//		MyStatus := Status{GetMyIP() + MyPort, "MSP", stat.ResidentMemory(), stat.ResidentMemory()}
//		StatusOfAll = append(StatusOfAll, &MyStatus)
//	}

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

//----------------------- LazyBoy-----------------------------------
// Node -> MSP GetLazyBoy
// pbft -> MSP FinishDelay
// MSP -> Node DelayResult
// var ADT int64
// var Start time.Time

//	func GetLazyBoy(w http.ResponseWriter, r *http.Request) {
//		logFile := OpenLogFile("Performance")
//		WriteLog(logFile, "startDelay,"+time.Now().Format("2006-01-02 15:04:05"))
//		defer logFile.Close()
//		Start = time.Now()
//		if len(PbftHostAddressTable) == 0 {
//			for K, V := range PbftReadyAddressTable {
//				PbftHostAddressTable[K] = V
//				err := Awake(V)
//				logFile := OpenLogFile("Error")
//				WriteLog(logFile, "error,"+err.Error())
//				defer logFile.Close()
//				json.NewEncoder(w).Encode(V)
//				break
//			}
//		} else {
//			for _, V := range PbftHostAddressTable {
//				json.NewEncoder(w).Encode(V)
//			}
//		}
//	}
// func Awake(address string) error {
// 	_, err := http.Post("http://"+address+"/awake", "text/plain", nil)
// 	return err
// }
// func FinishDelay(w http.ResponseWriter, r *http.Request) {
// 	reqData := new(ReqData)
// 	json.NewDecoder(r.Body).Decode(&reqData)
// 	DelayResult(*reqData)
// 	ADT = time.Since(Start).Milliseconds()
// 	logFile := OpenLogFile("Performance")
// 	WriteLog(logFile, "ADT,"+fmt.Sprint(ADT))

// }
// func DelayResult(reqData ReqData) {
// 	Data, err := json.Marshal(reqData)
// 	if err != nil {
// 		logFile := OpenLogFile("Error")
// 		WriteLog(logFile, "error,"+err.Error())
// 		defer logFile.Close()
// 	}
// 	for _, V := range CurrentAddressTable {
// 		_, err = http.Post("http://"+V+"/DelayResult", "application/json", bytes.NewBuffer(Data))
// 		logFile := OpenLogFile("Error")
// 		WriteLog(logFile, "error,"+err.Error())
// 		defer logFile.Close()
// 	}
// }
//	func CheckLazyBoyHost() {
//		if len(PbftHostAddressTable) > 0 {
//			for _, V := range PbftHostAddressTable {
//				logFile := OpenLogFile("Hosts")
//				WriteLog(logFile, "lazyBoyHost,"+V)
//				defer logFile.Close()
//			}
//		}
//	}
//
//	func Rest(w http.ResponseWriter, r *http.Request) {
//		for Node := range PbftHostAddressTable {
//			for K, V := range PbftReadyAddressTable {
//				PbftHostAddressTable[K] = V
//			}
//			delete(PbftHostAddressTable, Node)
//		}
//	}
