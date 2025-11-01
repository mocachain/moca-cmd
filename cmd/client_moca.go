package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/mocachain/moca-go-sdk/client"
	"github.com/mocachain/moca-go-sdk/types"
)

const (
	iso8601DateFormat = "2006-01-02 15:04:05"
	urlPrefix         = "moca://"
)

func cmdShowVersion() *cli.Command {
	return &cli.Command{
		Name:      "version",
		Action:    showVersion,
		Usage:     "print version info",
		ArgsUsage: "",
		Description: `

Examples:
$ moca-cmd version  `,
	}
}

func showVersion(ctx *cli.Context) error {
	fmt.Println("Moca Cmd Version:", Version)
	return nil
}

// NewClient returns a new moca client
func NewClient(ctx *cli.Context, opts ClientOptions) (client.IClient, error) {
	var (
		account    *types.Account
		err        error
		privateKey string
		cli        client.IClient
	)

	if !opts.IsQueryCmd {
		privateKey, _, err = parseKeystore(ctx)
		if err != nil {
			return nil, err
		}

		account, err = types.NewAccountFromPrivateKey("moca-account", privateKey)
		if err != nil {
			fmt.Println("new account err", err.Error())
			return nil, err
		}
	}

	rpcAddr, chainId, host, evmRpcAddress, err := getConfig(ctx)
	if err != nil {
		return nil, err
	}

	if host != "" {
		cli, err = client.New(chainId, rpcAddr, evmRpcAddress, privateKey, client.Option{DefaultAccount: account, Host: host, ForceToUseSpecifiedSpEndpointForDownloadOnly: opts.ForceToUseSpecifiedSpEndpointForDownloadOnly})
	} else {
		cli, err = client.New(chainId, rpcAddr, evmRpcAddress, privateKey, client.Option{DefaultAccount: account, ForceToUseSpecifiedSpEndpointForDownloadOnly: opts.ForceToUseSpecifiedSpEndpointForDownloadOnly})
	}

	if err != nil {
		fmt.Printf("failed to create client %s \n", err.Error())
		return nil, err
	}

	return cli, nil
}

// ParseBucketAndObject parse the bucket-name and object-name from url
func ParseBucketAndObject(urlPath string) (string, string, error) {
	if strings.Contains(urlPath, urlPrefix) {
		urlPath = urlPath[len(urlPrefix):]
	}

	index := strings.Index(urlPath, "/")

	if index <= -1 {
		return "", "", errors.New("url not right, can not parse bucket name and object name")
	}

	return urlPath[:index], urlPath[index+1:], nil
}

// ParseBucketAndPrefix parse the bucket-name, if prefix exist, return the prefix as well
func ParseBucketAndPrefix(urlPath string) (string, string, error) {
	if strings.Contains(urlPath, urlPrefix) {
		urlPath = urlPath[len(urlPrefix):]
	}

	index := strings.Index(urlPath, "/")

	if index <= -1 {
		return urlPath, "", nil
	}

	return urlPath[:index], urlPath[index+1:], nil
}

// ParseBucket parse the bucket-name from url
func ParseBucket(urlPath string) (bucketName string) {
	if strings.Contains(urlPath, urlPrefix) {
		urlPath = urlPath[len(urlPrefix):]
	}
	splits := strings.SplitN(urlPath, "/", 1)

	return splits[0]
}

func waitTxnStatus(cli client.IClient, ctx context.Context, txnHash string, txnInfo string) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, ContextTimeout)
	defer cancel()

	txnResponse, err := cli.WaitForTx(ctxTimeout, txnHash)
	if err != nil {
		return fmt.Errorf("the %s txn: %s ,has been submitted, please check it later:%v", txnInfo, txnHash, err)
	}
	if txnResponse.TxResult.Code != 0 {
		return fmt.Errorf("the %s txn: %s has failed with response code: %d", txnInfo, txnHash, txnResponse.TxResult.Code)
	}

	return nil
}
