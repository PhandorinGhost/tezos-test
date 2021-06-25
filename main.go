package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/goat-systems/go-tezos/v4/forge"
	"github.com/goat-systems/go-tezos/v4/keys"
	"github.com/goat-systems/go-tezos/v4/rpc"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	var address, amount string
	var sendToAddress = &cobra.Command{
		Use:   "sendtoaddress [--address ] [--amount ]",
		Short: "send tezos transaction",
		Run: func(cmd *cobra.Command, args []string) {
			public_RPC := os.Getenv("PUBLIC_RPC")

			key, err := keys.FromBytes([]byte(os.Getenv("PRIVATE_KEY")), keys.Ed25519)
			if err != nil {
				fmt.Printf("failed to import keys: %s\n", err.Error())
				os.Exit(1)
			}
			client, err := rpc.New(public_RPC)
			if err != nil {
				fmt.Printf("failed to initialize rpc client: %s\n", err.Error())
				os.Exit(1)
			}
			resp, counter, err := client.ContractCounter(rpc.ContractCounterInput{
				BlockID:    &rpc.BlockIDHead{},
				ContractID: key.PubKey.GetAddress(),
			})
			if err != nil {
				fmt.Printf("failed to get (%s) counter: %s\n", resp.Status(), err.Error())
				os.Exit(1)
			}
			counter++

			respA, balance, err := client.ContractBalance(rpc.ContractBalanceInput{
				BlockID:    &rpc.BlockIDHead{},
				ContractID: key.PubKey.GetAddress(),
			})
			if err != nil {
				fmt.Printf("failed to get (%s) counter: %s\n", respA.Status(), err.Error())
				os.Exit(1)
			}

			amountInt, _ := strconv.Atoi(amount)
			balanceInt, _ := strconv.Atoi(balance)
			if amountInt > balanceInt {
				fmt.Printf("Wallet amount less than transaction amount")
				os.Exit(1)
			}

			big.NewInt(0).SetString("10000000000000000000000000000", 10)

			transaction := rpc.Transaction{
				Source:      key.PubKey.GetPublicKey(),
				Fee:         "2941",
				GasLimit:    "26283",
				Counter:     strconv.Itoa(counter),
				Amount:      amount,
				Destination: address,
			}

			resp, head, err := client.Block(&rpc.BlockIDHead{})
			if err != nil {
				fmt.Printf("failed to get (%s) head block: %s\n", resp.Status(), err.Error())
				os.Exit(1)
			}

			op, err := forge.Encode(head.Hash, transaction.ToContent())
			if err != nil {
				fmt.Printf("failed to forge transaction: %s\n", err.Error())
				os.Exit(1)
			}

			signature, err := key.SignHex(op)
			if err != nil {
				fmt.Printf("failed to sign operation: %s\n", err.Error())
				os.Exit(1)
			}

			resp, ophash, err := client.InjectionOperation(rpc.InjectionOperationInput{
				Operation: signature.AppendToHex(op),
			})
			if err != nil {
				fmt.Printf("failed to inject (%s): %s\n", resp.Status(), err.Error())
				os.Exit(1)
			}

			fmt.Println(ophash)
		},
	}

	sendToAddress.PersistentFlags().StringVarP(&address, "address", "a", "", "tezos address (required)")
	sendToAddress.PersistentFlags().StringVarP(&amount, "amount", "m", "", "transaction sum (required)")

	var rootCmd = &cobra.Command{Use: "tezos"}
	rootCmd.AddCommand(sendToAddress)
	rootCmd.Execute()

}
