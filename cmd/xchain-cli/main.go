package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/virtue186/xchain/cmd/xchain-cli/account"
	"github.com/virtue186/xchain/cmd/xchain-cli/balance"
	"github.com/virtue186/xchain/cmd/xchain-cli/transfer"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "xchain-cli",
	Short: "A command-line client for interacting with the xchain blockchain",
	Long: `xchain-cli is a versatile tool to manage accounts, check balances, 
and transfer funds on the xchain network.`,
}

func init() {
	// 在根命令上定义一个持久化的字符串标志 "url"
	// 第一个参数是标志名称，第二个是默认值，第三个是帮助信息
	rootCmd.PersistentFlags().String("url", "http://localhost:8000/rpc", "URL of the xchain RPC API server")
}

func main() {

	// 将子命令添加到根命令中
	rootCmd.AddCommand(account.NewAccountCmd())
	rootCmd.AddCommand(balance.NewBalanceCmd())
	rootCmd.AddCommand(transfer.NewTransferCmd())

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your command '%s'", err)
		os.Exit(1)
	}

}
