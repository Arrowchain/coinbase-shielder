package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "os/exec"
  "strings"
  "time"
  "github.com/common-nighthawk/go-figure"
  _ "github.com/joho/godotenv/autoload"
)

var version = "1.0.0"
var debug bool
var rpcconnect string
var rpcport string
var rpcuser string
var rpcpass string
var amount int
var batch string
var fromAddress string
var toAddress string
var minConfs int
var pendingAmount float64
var confirmedAmount float64
var txPollInterval int
var defaultCliPath = "/Applications/quiver.app/Contents/MacOS"
var cliPath string
var cli = "/arrow-cli"

func panicIfNil(e error, msg string) {
	if e != nil {
		panic(fmt.Errorf("【ERR】 %s %v", msg, e))
	}
}

func panicWithMsg(msg string) {
  panic(fmt.Errorf(" [ERR] %s", msg))
}

func removeFormatting(s string) string {
  s = strings.ReplaceAll(s, "\\n", "")
  s = strings.ReplaceAll(s, "\\", "")
  s = strings.ReplaceAll(s, " ", "")
  s = strings.Replace(s, "\"{", "{", 1)
  s = strings.Replace(s, "}\"", "}", 1)
  s = strings.Replace(s, "\"[", "[", 1)
  s = strings.Replace(s, "]\"", "]", 1)
  return s
}

func runCmd(inputArgs []string) string {
  defaultArgs := []string{"-rpcconnect=" + rpcconnect, "-rpcport=" + rpcport}
  if rpcuser != "" {
    rpcUserArgs := []string{"-rpcuser=" + rpcuser}
    defaultArgs = append(defaultArgs, rpcUserArgs...)
  }
  if rpcpass != "" {
    rpcPassArgs := []string{"-rpcpassword=" + rpcpass}
    defaultArgs = append(defaultArgs, rpcPassArgs...)
  }
  cmdArgs := []string{}
  cmdArgs = append(cmdArgs, defaultArgs...)
  cmdArgs = append(cmdArgs, inputArgs...)
  cmdCli := cliPath + cli
  cmd := exec.Command(cmdCli, cmdArgs...)
	if debug {
    // print command
    fmt.Println("[CMD]", cmd.Args)
	}
	stderr, err := cmd.StderrPipe()
	panicIfNil(err, "Failed to get stderr pip ")

	stdout, err := cmd.StdoutPipe()
	panicIfNil(err, fmt.Sprintf("Failed to get stdout pipe %v", err))
	err = cmd.Start()
  if err != nil {
    log.Print("********************")
    log.Print("can't find arrow-cli")
    log.Printf("full location: %s", cliPath + cli)
    log.Fatal("exiting...")
  }
	b, err := ioutil.ReadAll(stdout)
	panicIfNil(err, fmt.Sprintf("Failed to read cmd (%v) stdout, %v", cmd, err))
	out := string(b)

	if debug {
    // print response
    fmt.Println(out)
	}

	bo, err := ioutil.ReadAll(stderr)
	panicIfNil(err, "Failed to read stderr")
	out += string(bo)

	cmd.Wait()
	stdout.Close()
	stderr.Close()
	response := strings.TrimSpace(out)
  res, _ := json.Marshal(response)
  formattingRemoved := removeFormatting(string(res))
  if strings.HasPrefix(formattingRemoved, "\"error") {
    panic(fmt.Errorf("【ERR】%s", formattingRemoved))
  }
  if debug {
    log.Print(formattingRemoved)
  }
  return formattingRemoved
}

func getinfo() int32 {
  cmdArgs := []string{"getinfo"}
  response := runCmd(cmdArgs)
  var info InfoResponse
  // response := "{\"version\":1020250,\"protocolversion\":170008,\"walletversion\":60000,\"balance\":148153.20000000,\"blocks\":573997,\"timeoffset\":0,\"connections\":8,\"proxy\":\"\",\"difficulty\":113.1619869062825,\"testnet\":false,\"keypoololdest\":1572054138,\"keypoolsize\":101,\"paytxfee\":0.00000000,\"relayfee\":0.00000100,\"errors\":\"\"}"
  json.Unmarshal([]byte(response), &info)
  return info.Connections
}

func shieldCoinbase() ShieldCoinbaseResponse {
  cmdArgs := []string{"z_shieldcoinbase", fromAddress, toAddress, "0.0001", batch}
  response := runCmd(cmdArgs)
  var shield ShieldCoinbaseResponse
  json.Unmarshal([]byte(response), &shield)
  return shield
}

func getOperationStatus(opid string) GetOperatoinStatusResponse {
  cmdArgs := []string{"z_getoperationstatus", "[\"" + opid + "\"]"}
  response := runCmd(cmdArgs)
  var operationStatus []GetOperatoinStatusResponse
  json.Unmarshal([]byte(response), &operationStatus)
  return operationStatus[0]
}

func getMempool() []string {
  cmdArgs := []string{"getrawmempool"}
  response := runCmd(cmdArgs)
  var mempool []string
  json.Unmarshal([]byte(response), &mempool)
  return mempool
}

func txInMempool(txid string) bool {
  mempool := getMempool()
  inMempool := false
  for _, tx := range mempool {
    if txid == tx {
      inMempool = true
      break
    }
  }
  if (inMempool) {
    log.Printf("transaction in mempool: %s", txid)
  }
  return inMempool
}

func getListUnspent(addr string) []ListUnspentResponse {
  // hard coding min confs to 1, max confs to 20. checking conf threshold elsewhere
  cmdArgs := []string{"z_listunspent", "1", "20", "false", "[\"" + addr + "\"]"}
  response := runCmd(cmdArgs)
  var listUnspent []ListUnspentResponse
  json.Unmarshal([]byte(response), &listUnspent)
  return listUnspent
}

func utxoExists(txid string) (bool, ListUnspentResponse) {
  var utxo ListUnspentResponse
  found := false
  unspents := getListUnspent(toAddress)
  for _, unspent := range unspents {
    if unspent.Txid == txid {
      found = true
      utxo = unspent
      log.Printf("found %s. confirmations: %v", txid, utxo.Confirmations)
    }
  }
  return found, utxo
}

func shieldCoinbaseLoop() {
  // 1 shield coinbase batch at a time. wait for tx to go through, then repeat.

  shieldResponse := shieldCoinbase()
  pendingAmount += shieldResponse.ShieldingValue
  if debug {
    log.Printf("shield amount pending %v", pendingAmount)
  }

  var operationStatusResponse GetOperatoinStatusResponse
  firstrun := true
  for (operationStatusResponse.Status == "executing" || firstrun) {
    log.Print("executing shielding...")
    firstrun = false
    time.Sleep(5 * time.Second)
    operationStatusResponse = getOperationStatus(shieldResponse.Opid)
  }
  // shielding is done, fail if not success
  if operationStatusResponse.Status != "success" {
    log.Print(operationStatusResponse)
    panicWithMsg("shield coinbase operation unsuccessful")
  }
  log.Print("shielding operation successful")
  log.Printf("%v shielded, %v pending, %v remaining", confirmedAmount, pendingAmount, (float64(amount) - confirmedAmount))
  log.Print("waiting for confirmation...")

  txid := operationStatusResponse.Result.Txid
  for {
    // get unspents list
    // is the transaction in a block?
    found, utxo := utxoExists(txid)
    if found {
      if utxo.Confirmations >= int32(minConfs) {
        // success, move on to next, if any
        pendingAmount -= utxo.Amount
        confirmedAmount += utxo.Amount
        log.Printf("total shielded %v", confirmedAmount)
        break
      } else if utxo.Confirmations < int32(minConfs) {
        // waiting for more confirmations
        log.Printf("waiting for at least %v confirmations...", minConfs)
      }
    } else {
      if !found && !txInMempool(txid) {
        // block reorg or tx expired, try again
        pendingAmount -= shieldResponse.ShieldingValue
        log.Print("tx timed out, trying again...")
        break
      } else {
        // waiting for confirmations
        log.Print("waiting for confirmations...")
      }
    }
    time.Sleep(time.Duration(txPollInterval) * time.Second)
  }
  if (confirmedAmount < float64(amount)) {
    shieldCoinbaseLoop()
  }
}

func setup() {
  var help bool
  flag.BoolVar(&help, "help", false, "print this message")
  var cliPathFlag string
  flag.StringVar(&cliPathFlag, "clipath", "default", "directory containing arrow-cli (no trailing /)\ndefault: " + defaultCliPath + "\ncurrent: current working directory \nor specify your on directory path")
  var rpcConnectFlag string
  flag.StringVar(&rpcConnectFlag, "rpcconnect", "127.0.0.1", "ip address to connect to")
  var rpcPortFlag string
  flag.StringVar(&rpcPortFlag, "rpcport", "6543", "port to connect to")
  var rpcUserFlag string
  flag.StringVar(&rpcUserFlag, "rpcuser", "", "user to use for connection")
  var rpcPassFlag string
  flag.StringVar(&rpcPassFlag, "rpcpass", "", "password to use for connection")
  var amountFlag int
  flag.IntVar(&amountFlag, "amount", 1, "minimum amount to shield")
  var batchFlag string
  flag.StringVar(&batchFlag, "batch", "50", "max number of coinbase UTXOs to shield at once")
  var fromAddressFlag string
  flag.StringVar(&fromAddressFlag, "from", "", "address that has unshielded coinbase UTXOs")
  var toAddressFlag string
  flag.StringVar(&toAddressFlag, "to", "", "address to send the shielded output to")
  var txPollIntervalFlag int
  flag.IntVar(&txPollIntervalFlag, "txpoll", 15, "interval in seconds at which to check if shielded tx confirmed")
  var minConfsFlag int
  flag.IntVar(&minConfsFlag, "minconfs", 2, "minimum block confirmations before shielding more")
  flag.Parse()

  if (os.Getenv("DEBUG") == "true") {
    debug = true
  } else {
    debug = false
  }
  if help {
    flag.PrintDefaults()
    os.Exit(0)
  }
  if cliPathFlag == "default" {
    cliPath = defaultCliPath
  } else if cliPathFlag == "current" {
    path, err := os.Getwd()
    panicIfNil(err, "can't get current directory, please use -clipath")
    cliPath = path
  } else {
    cliPath = cliPathFlag
  }
  rpcconnect = os.Getenv("RPCCONNECT")
  if (rpcconnect == "") {
    rpcconnect = rpcConnectFlag
  }
  rpcport = os.Getenv("RPCPORT")
  if (rpcport == "") {
    rpcport = rpcPortFlag
  }
  rpcuser = os.Getenv("RPCUSER")
  if (rpcuser == "") {
    rpcuser = rpcUserFlag
  }
  rpcpass = os.Getenv("RPCPASS")
  if (rpcpass == "") {
    rpcpass = rpcPassFlag
  }
  amount = amountFlag
  batch = batchFlag
  fromAddress = fromAddressFlag
  toAddress = toAddressFlag
  pendingAmount = 0.0
  confirmedAmount = 0.0
  txPollInterval = txPollIntervalFlag
  minConfs = minConfsFlag
}

func main() {
  figure.NewFigure("coinbase", "isometric1", true).Print()
  figure.NewFigure("shielder", "isometric1", true).Print()
  figure.NewFigure("v" + version, "", true).Print()
  fmt.Print("\n")
  fmt.Print("by j4ys0n")
  fmt.Print("\n\n")

  // set up variables
  setup()

  // get info from node, return if connected to network
  log.Print("checking if node is connected to the network...")
  connected := getinfo()
  if connected == 0 {
    panicWithMsg("node is not connected to the network, please wait and try again")
  }
  log.Printf("node has %v connections!", connected)

  // start shielding coinbase UTXOs
  shieldCoinbaseLoop()
}

type ShieldCoinbaseResponse struct {
  RemainingUTXOs    int32     `json:"remainingUTXOs"`
  RemainingValue    float64   `json:"remainingValue"`
  ShieldingUTXOs    int32     `json:"shieldingUTXOs"`
  ShieldingValue    float64   `json:"shieldingValue"`
  Opid              string    `json:"opid"`
}

type GetOperatoinStatusResponse struct {
  Id                string    `json:"id"`
  Status            string    `json:"status"`
  CreationTime      int64     `json:"creation_time"`
  Result struct {
    Txid            string    `json:"txid"`
  }                           `json:"result,omitempty"`
  ExecutionSecs     float64   `json:"execution_secs,omitempty"`
  Method            string    `json:"method"`
  Params struct {
    FromAddress     string    `json:"fromaddress"`
    ToAddress       string    `json:"toaddress"`
    Fee             float64   `json:"fee"`
  }                           `json:"params"`
}

type ListUnspentResponse struct {
  Txid              string    `json:"txid"`
  Outindex          int32     `json:"outindex"`
  Confirmations     int32     `json:"confirmations"`
  Spendable         bool      `json:"spendable"`
  Address           string    `json:"address"`
  Amount            float64   `json:"amount"`
  Memo              string    `json:"memo"`
  Change            bool      `json:"change"`
}

type InfoResponse struct {
  Version           int32     `json:"version"`
  Protocolversion   int32     `json:"protocolversion"`
  Walletversion     int32     `json:"walletversion"`
  Balance           float64   `json:"balance"`
  Blocks            int32     `json:"blocks"`
  Timeoffset        int64     `json:"timeoffset"`
  Connections       int32     `json:"connections"`
  Proxy             string    `json:"proxy"`
  Difficulty        float64   `json:"difficulty"`
  Testnet           bool      `json:"testnet"`
  Keypoololdest     int64     `json:"keypoololdest"`
  Keypoolsize       int32     `json:"keypoolsize"`
  Paytxfee          float64   `json:"paytxfee"`
  Relayfee          float64   `json:"relayfee"`
  Errors            string    `json:"errors"`
}
