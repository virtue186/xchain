package balance

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/virtue186/xchain/cmd/xchain-cli/client"
)

func NewBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance [address]",
		Short: "Query the balance and nonce of an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]
			apiEndpoint, err := cmd.Flags().GetString("url")
			if err != nil {
				return err // 如果标志不存在或类型错误，这里会报错
			}
			// 1. 创建一个新的 API 客户端
			cli := client.New(apiEndpoint)

			// 2. 调用客户端的方法
			state, err := cli.GetAccountState(address)
			if err != nil {
				return err
			}

			// 3. 打印结果
			fmt.Printf("State for address %s:\n", address)
			fmt.Printf("  Balance: %d\n", state.Balance)
			fmt.Printf("  Nonce:   %d\n", state.Nonce)

			return nil
		},
	}
	return cmd
}
