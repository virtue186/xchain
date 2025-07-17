package transfer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/virtue186/xchain/cmd/xchain-cli/client"
	"github.com/virtue186/xchain/core"
	"github.com/virtue186/xchain/crypto"
	"github.com/virtue186/xchain/types"
	"strconv"
)

// NewTransferCmd 返回一个用于发起交易的 cobra 命令
func NewTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer --from <private_key> --to <recipient_address> --amount <value>",
		Short: "Send funds from one account to another",
		Long: `Constructs a transaction, signs it with the sender's private key, 
and submits it to the blockchain network via RPC.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. 获取并校验命令标志 (flags)
			fromKeyHex, _ := cmd.Flags().GetString("from")
			toAddrHex, _ := cmd.Flags().GetString("to")
			amountStr, _ := cmd.Flags().GetString("amount")

			if fromKeyHex == "" || toAddrHex == "" || amountStr == "" {
				return fmt.Errorf("flags --from, --to, and --amount are all required")
			}

			// 2. 解析参数
			amount, err := strconv.ParseUint(amountStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			fromKey, err := crypto.NewPrivateKeyFromHex(fromKeyHex)
			if err != nil {
				return fmt.Errorf("invalid private key: %w", err)
			}
			fromAddr := fromKey.PublicKey().Address()

			toAddr, err := types.AddressFromHex(toAddrHex)
			if err != nil {
				return fmt.Errorf("invalid recipient address: %w", err)
			}
			apiEndpoint, err := cmd.Flags().GetString("url")
			if err != nil {
				return err // 如果标志不存在或类型错误，这里会报错
			}
			// 3. 【使用 Client】创建API客户端
			cli := client.New(apiEndpoint)

			// 4. 【使用 Client】通过API获取发送方当前的 Nonce
			fmt.Printf("Fetching nonce for sender %s...\n", fromAddr)
			state, err := cli.GetAccountState(fromAddr.String())
			if err != nil {
				return fmt.Errorf("failed to get current nonce for sender: %w", err)
			}
			nonce := state.Nonce
			fmt.Printf("Current nonce is %d. Proceeding to create transaction...\n", nonce)

			// 5. 【核心职责】构建、签名并序列化交易
			tx := core.NewTransaction(nil)
			tx.To = toAddr
			tx.Value = amount
			tx.Nonce = nonce

			if err := tx.Sign(fromKey); err != nil {
				return fmt.Errorf("failed to sign transaction: %w", err)
			}

			buf := new(bytes.Buffer)
			if err := tx.Encode(buf, core.GOBEncoder[*core.Transaction]{}); err != nil {
				return fmt.Errorf("failed to encode transaction: %w", err)
			}
			rawTxHex := hex.EncodeToString(buf.Bytes())
			fmt.Println("Transaction created and signed successfully.")

			// 6. 【使用 Client】发送原始交易
			fmt.Println("Submitting transaction to the network...")
			txHash, err := cli.SendRawTransaction(rawTxHex)
			if err != nil {
				return err // client包中的错误信息已经很清晰了
			}

			// 7. 打印最终结果
			fmt.Println("========================================================================")
			fmt.Printf("Transaction sent successfully!\n")
			fmt.Printf("Transaction Hash: %s\n", txHash)
			fmt.Println("========================================================================")

			return nil
		},
	}

	// 为命令添加必需的标志
	cmd.Flags().String("from", "", "Private key of the sender (in hex format)")
	cmd.Flags().String("to", "", "Recipient's address (in hex format)")
	cmd.Flags().String("amount", "", "Amount to send (as an integer)")

	return cmd
}
