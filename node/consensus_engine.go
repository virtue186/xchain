package node

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/network"
	"time"
)

type ConsensusEngineOpts struct {
	Logger           log.Logger         // 可选
	BlockTime        time.Duration      // 可选
	PrivateKey       *crypto.PrivateKey // 可选 (决定是否是验证者)
	BlockChain       *core.BlockChain   // 必需
	TxPool           *network.TxPool    // 必需
	BlockBroadcaster chan<- *core.Block // 必需
}

type ConsensusEngine struct {
	logger           log.Logger
	blockTime        time.Duration
	privateKey       *crypto.PrivateKey
	blockChain       *core.BlockChain
	txPool           *network.TxPool
	blockBroadcaster chan<- *core.Block
}

func NewConsensusEngine(opts ConsensusEngineOpts) (*ConsensusEngine, error) {
	// ✅ 在这里进行校验，确保必需的依赖被提供
	if opts.BlockChain == nil {
		return nil, fmt.Errorf("blockchain dependency cannot be nil")
	}
	if opts.TxPool == nil {
		return nil, fmt.Errorf("transaction pool dependency cannot be nil")
	}
	if opts.BlockBroadcaster == nil {
		return nil, fmt.Errorf("block broadcaster channel cannot be nil")
	}

	// 为可选参数设置默认值
	if opts.BlockTime == 0 {
		opts.BlockTime = 5 * time.Second
	}
	if opts.Logger == nil {
		opts.Logger = log.NewNopLogger()
	}

	return &ConsensusEngine{
		logger:           opts.Logger,
		blockTime:        opts.BlockTime,
		privateKey:       opts.PrivateKey,
		blockChain:       opts.BlockChain,
		txPool:           opts.TxPool,
		blockBroadcaster: opts.BlockBroadcaster,
	}, nil
}

func (ce *ConsensusEngine) IsValidator() bool {
	return ce.privateKey != nil
}

func (ce *ConsensusEngine) Start() {
	ticker := time.NewTicker(ce.blockTime)
	ce.logger.Log("msg", "starting consensus engine", "blockTime", ce.blockTime)

	for {
		<-ticker.C
		if err := ce.createNewBlock(); err != nil {
			ce.logger.Log("msg", "failed to create new block", "err", err)
		}
	}
}

func (ce *ConsensusEngine) createNewBlock() error {
	currentHeader, err := ce.blockChain.GetHeader(ce.blockChain.Height())

	if err != nil {
		return err
	}

	txx := ce.txPool.Pending()

	block, err := core.NewBlockFromPreHeader(currentHeader, txx)
	if err != nil {
		return err
	}
	err = block.Sign(*ce.privateKey)
	if err != nil {
		return err
	}

	err = ce.blockChain.AddBlock(block)
	if err != nil {
		return err
	}
	ce.txPool.ClearPending()

	go func() {
		ce.blockBroadcaster <- block
	}()
	ce.logger.Log("msg", "successfully created new block and sent to broadcaster", "hash", block.Hash(core.BlockHasher{}))
	return nil
}
