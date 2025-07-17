package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/api"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/network"
	"github.com/virtue186/xchain/node"
	"github.com/virtue186/xchain/types"
	"os"
	"path/filepath"
	"time"
)

// Genesis 定义了 genesis.json 文件的结构
type Genesis struct {
	Header       *core.Header                 `json:"header"`
	Transactions []*core.Transaction          `json:"transactions"`
	Alloc        map[string]map[string]uint64 `json:"alloc"`
}

// loadGenesis 从 JSON 文件加载创世配置
func loadGenesis(path string) (*Genesis, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	g := new(Genesis)
	if err := json.Unmarshal(content, g); err != nil {
		return nil, err
	}
	return g, nil
}

// main 函数现在只负责启动网络
func main() {
	// 1. 加载创世配置
	genesisData, err := loadGenesis("genesis.json")
	if err != nil {
		panic(fmt.Errorf("failed to load genesis file: %w", err))
	}

	// 2. 为验证者节点创建一个随机私钥
	validatorKey := crypto.GeneratePrivateKey()

	// 3. 创建并启动三个节点
	fmt.Println("Starting blockchain nodes...")
	validatorTr, _ := makeNode("127.0.0.1:3000", "127.0.0.1:8000", &validatorKey, genesisData)
	remoteATr, _ := makeNode("127.0.0.1:4000", "127.0.0.1:8001", nil, genesisData)
	remoteBTr, _ := makeNode("127.0.0.1:5000", "127.0.0.1:8002", nil, genesisData)

	// 4. 等待节点启动，然后建立P2P连接
	time.Sleep(1 * time.Second)
	fmt.Println("Connecting nodes...")
	if err := remoteATr.Dial(string(validatorTr.Addr())); err != nil {
		fmt.Printf("Error dialing validator from node A: %v\n", err)
	}
	if err := remoteBTr.Dial(string(validatorTr.Addr())); err != nil {
		fmt.Printf("Error dialing validator from node B: %v\n", err)
	}

	fmt.Println("Blockchain network is running. Use xchain-cli to interact.")

	// 5. 永久阻塞，让节点持续运行
	select {}
}

// makeNode 函数负责组装和初始化一个节点
func makeNode(listenAddr, apiListenAddr string, pk *crypto.PrivateKey, genesisData *Genesis) (network.Transport, *node.Node) {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "node", listenAddr)

	genesisBlock, err := core.NewBlock(genesisData.Header, genesisData.Transactions)
	if err != nil {
		panic(err)
	}

	// 初始化存储
	dbPath := filepath.Join("./db", fmt.Sprintf("node_%s", listenAddr))
	if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
		panic(err)
	}
	storage, err := core.NewLeveldbStorage(dbPath)
	if err != nil {
		panic(err)
	}

	// 创建或加载区块链和状态机
	bc, err := core.NewBlockChain(log.With(logger, "module", "blockchain"), storage, genesisBlock)
	if err != nil {
		panic(err)
	}

	// 如果是新链，则进行创世分配
	if bc.Height() == 0 {
		logger.Log("msg", "blockchain is new, performing genesis allocation")
		for addrStr, data := range genesisData.Alloc {
			addrBytes, err := hex.DecodeString(addrStr)
			if err != nil {
				panic(fmt.Errorf("invalid address in genesis alloc: %s", addrStr))
			}
			addr := types.AddressFromBytes(addrBytes)
			balance := data["balance"]
			account := &core.AccountState{
				Address: addr,
				Balance: balance,
				Nonce:   0,
			}
			if err := bc.State.Put(addr, account); err != nil {
				panic(err)
			}
			logger.Log("msg", "genesis allocation", "address", addr, "balance", balance)
		}
	}

	// 初始化交易池、API服务器和节点
	txPool := network.NewTxPool(1000)
	apiServer := api.NewAPIServer(apiListenAddr, log.With(logger, "module", "api"), bc, txPool)

	opts := network.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: network.NOPHandshakeFunc,
		Decoder:       core.JSONDecoder[*network.Message]{},
	}
	tr := network.NewTCPTransport(opts)

	nodeOpts := node.NodeOpts{
		Logger:     logger,
		Transport:  tr,
		BlockChain: bc,
		TxPool:     txPool,
		PrivateKey: pk,
		BlockTime:  5 * time.Second,
		APIServer:  apiServer,
	}
	nodeInstance, err := node.NewNode(nodeOpts)
	if err != nil {
		panic(err)
	}

	// 启动节点的主逻辑
	go nodeInstance.Start()

	return tr, nodeInstance
}
