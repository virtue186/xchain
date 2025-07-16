package node

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/network"
	"time"
)

type Node struct {
	logger           log.Logger
	chainService     *ChainService
	server           *network.Server
	consensusEngine  *ConsensusEngine // 假设我们有一个共识引擎
	broadcastService *BroadcastService
}

type NodeOpts struct {
	Logger     log.Logger
	Transport  network.Transport
	BlockChain *core.BlockChain
	TxPool     *network.TxPool
	PrivateKey *crypto.PrivateKey
	BlockTime  time.Duration
	Encoder    core.Encoder[any]
}

func NewNode(opts NodeOpts) (*Node, error) {

	// 1. 初始化编码器，提供默认值
	encoder := opts.Encoder
	if encoder == nil {
		encoder = core.GOBEncoder[any]{}
	}

	// 2. 初始化 ServerOpts，但先不创建Server，因为Server依赖RPCProcessor
	serverOpts := network.ServerOpts{
		Logger:     opts.Logger,
		Transports: []network.Transport{opts.Transport},
		ID:         fmt.Sprintf("NODE-%s", opts.Transport.Addr()),
		Encoder:    encoder, // 注入统一的编码器
		// Decoder 将在 NewServer 内部使用默认值
	}
	server := network.NewServer(serverOpts)

	// 3. 初始化 BroadcastService，它依赖 Server 和 Encoder
	broadcastService := NewBroadcastService(opts.Logger, server, encoder)

	// 4. 初始化 ChainService，它依赖 BroadcastService
	chainService := NewChainService(
		opts.BlockChain,
		opts.TxPool,
		opts.Logger,
		broadcastService.TxBroadcastChan(), // 现在 broadcastService 已经存在
	)

	optsForCE := ConsensusEngineOpts{
		Logger:           opts.Logger,
		BlockTime:        opts.BlockTime,
		BlockChain:       opts.BlockChain,
		TxPool:           opts.TxPool,
		PrivateKey:       opts.PrivateKey,
		BlockBroadcaster: broadcastService.BlockBroadcastChan(),
	}
	consensusEngine, err := NewConsensusEngine(optsForCE)
	if err != nil {
		return nil, err
	}

	// 6. 将完全初始化的 chainService 设置为 Server 的处理器
	server.RPCProcessor = chainService

	// 7. 返回完全组装好的 Node
	return &Node{
		logger:           opts.Logger,
		chainService:     chainService,
		server:           server,
		consensusEngine:  consensusEngine,
		broadcastService: broadcastService,
	}, nil
}

func (n *Node) Start() {
	n.logger.Log("msg", "starting node...")

	go n.broadcastService.Start()
	// 启动共识引擎（如果它是验证者）
	if n.consensusEngine.IsValidator() {
		go n.consensusEngine.Start()
	}

	// 启动网络服务器
	n.server.Start()
}
