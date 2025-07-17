package account

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/virtue186/xchain/crypto"
)

func NewAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new account (key pair)",
		Long:  `Generates a new private key and its corresponding public address.`,
		Run: func(cmd *cobra.Command, args []string) {
			// 1. 生成私钥
			privateKey := crypto.GeneratePrivateKey()

			// 2. 获取对应的地址
			address := privateKey.PublicKey().Address()

			// 3. 打印结果
			fmt.Println("New account created successfully!")
			fmt.Println("=============================================================================================================")
			fmt.Printf("Private Key: %s  (SAVE THIS securely, it cannot be recovered!)\n", privateKey.String())
			fmt.Printf("Address:     %s\n", address.String())
			fmt.Println("=============================================================================================================")
		},
	}
	return cmd
}
