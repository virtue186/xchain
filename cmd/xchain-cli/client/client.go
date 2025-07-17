package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client 是一个与 xchain 节点 RPC API 交互的客户端
type Client struct {
	Endpoint string
}

// New 创建一个新的 Client 实例
func New(endpoint string) *Client {
	return &Client{Endpoint: endpoint}
}

// GetAccountState 调用 get_account_state RPC 方法
func (c *Client) GetAccountState(address string) (*AccountStateResponse, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "get_account_state",
		"params":  map[string]string{"address": address},
	})

	resp, err := http.Post(c.Endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API server: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var rpcResp struct {
		Result *AccountStateResponse `json:"result"`
		Error  *RPCError             `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w\nResponse body: %s", err, string(bodyBytes))
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", rpcResp.Error.Message)
	}
	if rpcResp.Result == nil {
		return nil, fmt.Errorf("received empty result from API")
	}

	return rpcResp.Result, nil
}

// SendRawTransaction 调用 send_raw_transaction RPC 方法
func (c *Client) SendRawTransaction(txHex string) (string, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "send_raw_transaction",
		"params":  map[string]string{"tx_data": txHex},
	})

	resp, err := http.Post(c.Endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to API server: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var rpcResp struct {
		Result string    `json:"result"` // 交易哈希是字符串
		Error  *RPCError `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &rpcResp); err != nil {
		return "", fmt.Errorf("failed to parse RPC response: %w\nResponse body: %s", err, string(bodyBytes))
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("API error: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// --- 辅助数据结构 ---

type AccountStateResponse struct {
	Balance uint64 `json:"balance"`
	Nonce   uint64 `json:"nonce"`
}

type RPCError struct {
	Message string `json:"message"`
}
