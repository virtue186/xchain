package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/network"
	"github.com/virtue186/xchain/types"
	"net/http"
)

type APIServer struct {
	listenAddr string
	logger     log.Logger
	bc         *core.BlockChain // 持有对区块链核心的引用，以便查询数据
	txPool     *network.TxPool
}

func NewAPIServer(listenAddr string, logger log.Logger, bc *core.BlockChain, txPool *network.TxPool) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		logger:     logger,
		bc:         bc,
		txPool:     txPool,
	}
}

func (s *APIServer) Run() error {
	s.logger.Log("msg", "starting API server", "listenAddr", s.listenAddr)

	// 创建一个新的 HTTP 请求多路复用器 (router)
	mux := http.NewServeMux()
	// 为我们的 RPC 端点注册一个处理器
	mux.HandleFunc("/rpc", s.handleRPC)

	// 启动服务器
	return http.ListenAndServe(s.listenAddr, mux)
}

// JSONRPCRequest 定义了 JSON-RPC 2.0 请求的结构
type JSONRPCRequest struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"` // 使用 RawMessage 延迟解析参数
	ID      int             `json:"id"`
}

// JSONRPCResponse 定义了 JSON-RPC 2.0 响应的结构
type JSONRPCResponse struct {
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      int         `json:"id"`
}

// RPCError 定义了 JSON-RPC 错误对象的结构
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// handleRPC 是处理所有RPC请求的核心函数
func (s *APIServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	var req JSONRPCRequest
	// 解码请求体中的JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 如果解码失败，返回一个格式错误的JSON-RPC响应
		writeError(w, -32700, "Parse error", 0)
		return
	}

	s.logger.Log("msg", "received rpc request", "method", req.Method, "id", req.ID)

	// 根据请求的 method 字段，调用不同的处理函数
	switch req.Method {
	case "get_account_state":
		s.handleGetAccountState(w, req)
	case "send_raw_transaction": // 【新增】
		s.handleSendRawTransaction(w, req)
	default:
		// 如果方法不存在
		writeError(w, -32601, fmt.Sprintf("method not found: %s", req.Method), req.ID)
	}
}

// writeError 是一个辅助函数，用于方便地写入JSON-RPC错误响应
func writeError(w http.ResponseWriter, code int, message string, id int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest) // 通常RPC错误也使用400或500状态码
	resp := JSONRPCResponse{
		Version: "2.0",
		Error:   &RPCError{Code: code, Message: message},
		ID:      id,
	}
	json.NewEncoder(w).Encode(resp)
}

type GetAccountStateParams struct {
	Address string `json:"address"`
}

// AccountStateResponse 定义了返回给客户端的账户状态
type AccountStateResponse struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
	Nonce   uint64 `json:"nonce"`
}

// handleGetAccountState 处理查询账户状态的请求
func (s *APIServer) handleGetAccountState(w http.ResponseWriter, req JSONRPCRequest) {
	var params GetAccountStateParams
	// 解析特定于此方法的参数
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeError(w, -32602, "Invalid params", req.ID)
		return
	}

	// 将字符串地址转换为 types.Address
	addr, err := types.AddressFromHex(params.Address)
	if err != nil {
		writeError(w, -32602, fmt.Sprintf("invalid address format: %s", params.Address), req.ID)
		return
	}

	// 【核心逻辑】: 通过 blockchain 的 state 查询账户状态
	// 这里我们假设 BlockChain 有一个方法可以暴露它的 State
	// 或者直接让 APIServer 持有 State 的引用
	accountState, err := s.bc.State.Get(addr) // 假设 s.bc.State() 可以获取到 State 对象
	if err != nil {
		// 数据库层面的错误
		writeError(w, -32000, fmt.Sprintf("internal server error: %s", err), req.ID)
		return
	}

	// 准备成功的响应
	respBody := AccountStateResponse{
		Address: params.Address,
		Balance: accountState.Balance,
		Nonce:   accountState.Nonce,
	}
	resp := JSONRPCResponse{
		Version: "2.0",
		Result:  respBody,
		ID:      req.ID,
	}

	// 发送成功的响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type SendRawTxParams struct {
	TxData string `json:"tx_data"`
}

// 【新增】handleSendRawTransaction 处理提交原始交易的请求
func (s *APIServer) handleSendRawTransaction(w http.ResponseWriter, req JSONRPCRequest) {
	var params SendRawTxParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeError(w, -32602, "Invalid params", req.ID)
		return
	}

	// 1. 将十六进制字符串解码为字节
	txBytes, err := hex.DecodeString(params.TxData)
	if err != nil {
		writeError(w, -32602, "Invalid tx_data: not a valid hex string", req.ID)
		return
	}

	// 2. 将字节反序列化为 Transaction 对象
	tx := new(core.Transaction)
	if err := tx.Decode(bytes.NewReader(txBytes), core.JSONDecoder[*core.Transaction]{}); err != nil {
		writeError(w, -32602, fmt.Sprintf("Invalid tx_data: failed to decode transaction: %s", err), req.ID)
		return
	}
	// 3. 【核心逻辑】将交易添加到交易池
	s.txPool.Add(tx)

	// 4. 如果成功，返回交易的哈希值
	hash := tx.Hash(core.TxHasher{})
	resp := JSONRPCResponse{
		Version: "2.0",
		Result:  hash.String(), // 返回交易哈希
		ID:      req.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	s.logger.Log("msg", "transaction received via api", "hash", hash)
}
