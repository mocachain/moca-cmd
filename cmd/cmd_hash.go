package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// cmdMakeBucket create a new Bucket
func cmdCalHash() *cli.Command {
	return &cli.Command{
		Name:      "get-hash",
		Action:    computeHashRoot,
		Usage:     "compute the integrity hash of file",
		ArgsUsage: "filePath",
		Description: `
Compute the integrity hash value of the file which use same algorithm of moca

Examples:
$ moca-cmd object get-hash /home/test.text `,
	}
}

func computeHashRoot(ctx *cli.Context) error {
	// read the local file payload to be uploaded
	filePath := ctx.Args().Get(0)

	exists, objectSize, err := pathExists(filePath)
	if err != nil {
		return toCmdErr(err)
	}

	if !exists {
		return errors.New("upload file not exists")
	} else if objectSize > int64(5*1024*1024*1024) {
		return errors.New("file size should less than 5G")
	}

	mocaClient, err := NewClient(ctx, ClientOptions{IsQueryCmd: false})
	if err != nil {
		return toCmdErr(err)
	}

	// Open the referenced file.
	fReader, err := os.Open(filePath)
	if err != nil {
		return toCmdErr(err)
	}
	defer func() { _ = fReader.Close() }()

	hashes, size, _, err := mocaClient.ComputeHashRoots(fReader, false)
	if err != nil {
		return toCmdErr(fmt.Errorf("compute hash root fail: %w", err))
	}

	fmt.Printf("the primary sp hash root: \n%s\n%s\n", hex.EncodeToString(hashes[0]), "the secondary sp hash list:")

	for _, hash := range hashes[1:] {
		fmt.Println(hex.EncodeToString(hash))
	}

	fmt.Println("file size:", size)

	return nil
}
