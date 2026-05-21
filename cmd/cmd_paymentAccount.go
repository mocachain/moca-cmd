package main

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mocachain/moca/v2/sdk/types"
	"github.com/urfave/cli/v2"
)

// cmdCreatePaymentAccount creates a payment account under the owner
func cmdCreatePaymentAccount() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Action:    CreatePaymentAccount,
		Usage:     "create a payment account",
		ArgsUsage: "",
		Description: `
Create a payment account

Examples:
# Create a payment account
$ moca-cmd payment-account create`,
	}
}

func CreatePaymentAccount(ctx *cli.Context) error {
	client, err := NewClient(ctx, ClientOptions{IsQueryCmd: false})
	if err != nil {
		return toCmdErr(err)
	}
	c, createPaymentAccount := context.WithCancel(globalContext)
	defer createPaymentAccount()
	acc, err := client.GetDefaultAccount()
	if err != nil {
		return toCmdErr(err)
	}
	txHash, err := client.CreatePaymentAccount(c, acc.GetAddress().String(), types.TxOption{})
	if err != nil {
		return toCmdErr(err)
	}

	err = waitTxnStatus(client, c, txHash, "CreatePaymentAccount")
	if err != nil {
		return toCmdErr(err)
	}

	fmt.Printf("create payment account for %s succ, txHash: %s\n", acc.GetAddress().String(), txHash)
	return nil
}

// cmdPaymentDeposit makes deposit from the owner account to the payment account
func cmdPaymentDeposit() *cli.Command {
	return &cli.Command{
		Name:   "deposit",
		Action: Deposit,
		Usage:  "deposit into stream(payment) account",
		Description: `
Make a deposit into stream(payment) account 

Examples:
# deposit a stream account
$ moca-cmd payment-account deposit --toAddress 0x.. --amount 12345`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     toAddressFlag,
				Value:    "",
				Usage:    "the stream account",
				Required: true,
			},
			&cli.StringFlag{
				Name:  amountFlag,
				Value: "",
				Usage: "the amount to be deposited",
			},
		},
	}
}

func Deposit(ctx *cli.Context) error {
	client, err := NewClient(ctx, ClientOptions{IsQueryCmd: false})
	if err != nil {
		return toCmdErr(err)
	}

	toAddr := ctx.String(toAddressFlag)
	_, err = sdk.AccAddressFromHexUnsafe(toAddr)
	if err != nil {
		return toCmdErr(err)
	}
	amountStr := ctx.String(amountFlag)
	amount, ok := math.NewIntFromString(amountStr)
	if !ok {
		return toCmdErr(fmt.Errorf("invalid amount %s", amountStr))
	}
	c, deposit := context.WithCancel(globalContext)
	defer deposit()

	txHash, err := client.Deposit(c, toAddr, amount, types.TxOption{})
	if err != nil {
		return toCmdErr(err)
	}

	err = waitTxnStatus(client, c, txHash, "Deposit")
	if err != nil {
		return toCmdErr(err)
	}
	fmt.Printf("Deposit %s amoca to payment account %s succ, txHash=%s\n", amount.String(), toAddr, txHash)
	return nil
}

// cmdPaymentWithdraw makes a withdrawal from payment account to owner account
func cmdPaymentWithdraw() *cli.Command {
	return &cli.Command{
		Name:   "withdraw",
		Action: Withdraw,
		Usage:  "withdraw from stream(payment) account",
		Description: `
Make a withdrawal from stream(payment) account 

Examples:
# withdraw from a stream account back to the creator account
$ moca-cmd payment-account withdraw --fromAddress 0x.. --amount 12345`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     fromAddressFlag,
				Value:    "",
				Usage:    "the stream account",
				Required: true,
			},
			&cli.StringFlag{
				Name:  amountFlag,
				Value: "",
				Usage: "the amount to be withdrew",
			},
		},
	}
}

func Withdraw(ctx *cli.Context) error {
	client, err := NewClient(ctx, ClientOptions{IsQueryCmd: false})
	if err != nil {
		return toCmdErr(err)
	}

	fromAddr := ctx.String(fromAddressFlag)
	_, err = sdk.AccAddressFromHexUnsafe(fromAddr)
	if err != nil {
		return toCmdErr(err)
	}
	amountStr := ctx.String(amountFlag)
	amount, ok := math.NewIntFromString(amountStr)
	if !ok {
		return toCmdErr(fmt.Errorf("invalid amount %s", amountStr))
	}
	c, deposit := context.WithCancel(globalContext)
	defer deposit()

	txHash, err := client.Withdraw(c, fromAddr, amount, types.TxOption{})
	if err != nil {
		return toCmdErr(err)
	}

	err = waitTxnStatus(client, c, txHash, "Withdraw")
	if err != nil {
		return toCmdErr(err)
	}

	fmt.Printf("Withdraw %s from %s succ, txHash=%s\n", amount.String(), fromAddr, txHash)
	return nil
}

// cmdListPaymentAccounts list the payment accounts belong to the owner
func cmdListPaymentAccounts() *cli.Command {
	return &cli.Command{
		Name:      "ls",
		Action:    listPaymentAccounts,
		Usage:     "list payment accounts of the owner",
		ArgsUsage: "address of owner",
		Description: `
List payment accounts of the owner.

Examples:
$ moca-cmd payment-account ls `,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  ownerAddressFlag,
				Value: "",
				Usage: "indicate a owner's payment accounts to be list, account address can be omitted for current user's accounts listing",
			},
		},
	}
}

// cmdGetStreamRecord query the stream record of a payment account
func cmdGetStreamRecord() *cli.Command {
	return &cli.Command{
		Name:      "stream-record",
		Action:    getStreamRecord,
		Usage:     "query stream record (balance) of a payment account",
		ArgsUsage: "[payment-account-address]",
		Description: `
Query the stream record of a payment account, which contains balance information.

Examples:
$ moca-cmd payment-account stream-record 0x...`,
	}
}

func getStreamRecord(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return toCmdErr(fmt.Errorf("payment account address is required"))
	}

	paymentAddress := ctx.Args().Get(0)

	client, err := NewClient(ctx, ClientOptions{IsQueryCmd: true})
	if err != nil {
		return toCmdErr(err)
	}

	c, cancel := context.WithCancel(globalContext)
	defer cancel()

	streamRecord, err := client.GetStreamRecord(c, paymentAddress)
	if err != nil {
		return toCmdErr(err)
	}

	fmt.Printf("Stream Record for %s:\n", paymentAddress)
	fmt.Printf("  Account: %s\n", streamRecord.Account)
	fmt.Printf("  Static Balance: %s wei\n", streamRecord.StaticBalance.String())
	fmt.Printf("  Buffer Balance: %s wei\n", streamRecord.BufferBalance.String())
	fmt.Printf("  Lock Balance: %s wei\n", streamRecord.LockBalance.String())
	fmt.Printf("  Netflow Rate: %s wei/sec\n", streamRecord.NetflowRate.String())
	fmt.Printf("  Frozen Netflow Rate: %s wei/sec\n", streamRecord.FrozenNetflowRate.String())
	fmt.Printf("  Status: %s\n", streamRecord.Status.String())
	fmt.Printf("  CRUD Timestamp: %d\n", streamRecord.CrudTimestamp)
	fmt.Printf("  Settle Timestamp: %d\n", streamRecord.SettleTimestamp)
	fmt.Printf("  Out Flow Count: %d\n", streamRecord.OutFlowCount)

	return nil
}

func listPaymentAccounts(ctx *cli.Context) error {
	client, err := NewClient(ctx, ClientOptions{IsQueryCmd: true})
	if err != nil {
		return toCmdErr(err)
	}

	c, cancelCreateBucket := context.WithCancel(globalContext)
	defer cancelCreateBucket()

	var ownerAddr string
	ownerAddrStr := ctx.String(ownerAddressFlag)
	if ownerAddrStr != "" {
		_, err = sdk.AccAddressFromHexUnsafe(ownerAddrStr)
		if err != nil {
			return toCmdErr(err)
		}
		ownerAddr = ownerAddrStr
	} else {
		acct, err := client.GetDefaultAccount()
		if err != nil {
			return toCmdErr(err)
		}
		ownerAddr = acct.GetAddress().String()
	}
	accounts, err := client.GetPaymentAccountsByOwner(c, ownerAddr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fmt.Println("Accounts not exist")
			return nil
		}
		return toCmdErr(err)
	}
	if len(accounts) == 0 {
		fmt.Println("Accounts not exist")
		return nil
	}
	fmt.Println("payment accounts list:")
	for i, a := range accounts {
		fmt.Printf("%d: %s \n", i+1, a)
	}
	return nil
}
