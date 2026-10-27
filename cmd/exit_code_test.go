package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mocachain/moca-go-sdk/client"
	sdktypes "github.com/mocachain/moca-go-sdk/types"
	sptypes "github.com/mocachain/moca/v2/x/sp/types"
	storagetypes "github.com/mocachain/moca/v2/x/storage/types"
	"github.com/urfave/cli/v2"
)

const (
	runMainEnv   = "MOCA_CMD_TEST_RUN_MAIN"
	fakeChainEnv = "MOCA_CMD_TEST_FAKE_CHAIN"
	deadRPC      = "http://127.0.0.1:1"
)

// TestMain re-executes the test binary as the moca-cmd CLI so the tests see real exit codes.
func TestMain(m *testing.M) {
	if os.Getenv(runMainEnv) != "" {
		if os.Getenv(fakeChainEnv) != "" {
			newSDKClient = func(string, string, string, string, client.Option) (client.IClient, error) {
				return failingChain{}, nil
			}
		}
		main()
		return
	}
	os.Exit(m.Run())
}

var errChain = errors.New("chain unavailable")

// failingChain is a chain client that finds nothing and fails every transaction.
type failingChain struct{ client.IClient }

func (failingChain) ListStorageProviders(context.Context, bool) ([]sptypes.StorageProvider, error) {
	return nil, errChain
}
func (failingChain) HeadObject(context.Context, string, string) (*sdktypes.ObjectDetail, error) {
	return nil, errChain
}
func (failingChain) HeadGroup(context.Context, string, string) (*storagetypes.GroupInfo, error) {
	return nil, errChain
}
func (failingChain) HeadGroupMember(context.Context, string, string, string) bool { return false }
func (failingChain) UpdateBucketInfo(context.Context, string, sdktypes.UpdateBucketOptions) (string, error) {
	return "", errChain
}
func (failingChain) MigrateBucket(context.Context, string, uint32, sdktypes.MigrateBucketOptions) (string, error) {
	return "", errChain
}
func (failingChain) DeleteBucket(context.Context, string, sdktypes.DeleteBucketOption) (string, error) {
	return "", errChain
}
func (failingChain) BuyQuotaForBucket(context.Context, string, uint64, sdktypes.BuyQuotaOption) (string, error) {
	return "", errChain
}
func (failingChain) UpdateObjectVisibility(context.Context, string, string, storagetypes.VisibilityType, sdktypes.UpdateObjectOption) (string, error) {
	return "", errChain
}
func (failingChain) DeleteObject(context.Context, string, string, sdktypes.DeleteObjectOption) (string, error) {
	return "", errChain
}
func (failingChain) GetAccountBalance(context.Context, string) (*sdk.Coin, error) {
	return nil, errors.New("rpc error: key not found")
}

func (failingChain) HeadBucket(_ context.Context, name string) (*storagetypes.BucketInfo, error) {
	if name == "missing" {
		return nil, errChain
	}
	return &storagetypes.BucketInfo{}, nil
}

func (failingChain) ListObjects(context.Context, string, sdktypes.ListObjectsOptions) (sdktypes.ListObjectsResult, error) {
	return sdktypes.ListObjectsResult{Objects: []*sdktypes.ObjectMeta{{ObjectInfo: &storagetypes.ObjectInfo{ObjectName: "dir/a"}}}}, nil
}

type cliResult struct {
	code           int
	stdout, stderr string
}

// runCLI runs the real main() in a child process with a temporary HOME.
func runCLI(t *testing.T, fakeChain bool, args ...string) cliResult {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Env = append(os.Environ(), runMainEnv+"=1", "HOME="+t.TempDir())
	if fakeChain {
		cmd.Env = append(cmd.Env, fakeChainEnv+"=1")
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	res := cliResult{}
	if err = cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run %v: %v", args, err)
		}
		res.code = exitErr.ExitCode()
	}
	res.stdout, res.stderr = stdout.String(), stderr.String()
	return res
}

// assertFailed checks the process exited 1 and reported the failure once on stdout only.
func assertFailed(t *testing.T, res cliResult, want string) {
	t.Helper()
	if res.code != 1 {
		t.Errorf("exit code = %d, want 1\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	if n := strings.Count(res.stdout, "run command error:"); n != 1 {
		t.Errorf("stdout has %d error reports, want 1: %q", n, res.stdout)
	}
	if !strings.Contains(res.stdout, want) {
		t.Errorf("stdout = %q, want it to contain %q", res.stdout, want)
	}
	if res.stderr != "" {
		t.Errorf("stderr = %q, want it empty", res.stderr)
	}
}

func Test_commandFailureExitsNonZero(t *testing.T) {
	home := t.TempDir()
	existingKeystore := filepath.Join(home, "existing.json")
	if err := os.WriteFile(existingKeystore, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	global := []string{"--rpcAddr", deadRPC, "--chainId", "moca_5151-1", "--evmRpcAddr", deadRPC, "--home", home}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"sp ls on an unreachable rpc", []string{"sp", "ls"}, "connection refused"},
		{"bucket head on an unreachable rpc", []string{"bucket", "head", "moca://b"}, "connection refused"},
		{"object head on an unreachable rpc", []string{"object", "head", "moca://b/o"}, "connection refused"},
		{"bucket ls without a keystore", []string{"bucket", "ls"}, "invalid default address"},
		{"bucket create without a keystore", []string{"bucket", "create", "moca://b"}, "invalid default address"},
		{"object get without a keystore", []string{"object", "get", "moca://b/o"}, "invalid default address"},
		{"account import of a missing file", []string{"account", "import", filepath.Join(home, "none.txt")}, "failed to load private key"},
		{"account new over an existing keystore", []string{"--keystore", existingKeystore, "account", "new"}, "key already exists"},
		{"group head-member with one argument", []string{"group", "head-member", "only-one"}, "args number should be 2"},
		{"policy put of an invalid resource", []string{"policy", "put", "--grantee", "0x0000000000000000000000000000000000000001", "--actions", "get", "bad"}, "invalid resource name"},
		{"task status of an unknown task", []string{"task", "status", "--taskId", "none"}, "task not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFailed(t, runCLI(t, false, slices.Concat(global, tt.args)...), tt.want)
		})
	}
}

// writeTestKeystore stores the testdata key as the default account of home and returns the password file.
func writeTestKeystore(t *testing.T, home string) (passwordFile, address string) {
	t.Helper()
	privateKey, addr, err := loadKey("testdata/private_key_file")
	if err != nil {
		t.Fatal(err)
	}
	// scrypt N=2 keeps the keystore cheap to decrypt in every child process
	blob, err := EncryptKey(&Key{Address: addr, PrivateKey: privateKey}, "pw", 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	lower := convertAddressToLower(addr.String())
	files := map[string][]byte{
		filepath.Join(home, DefaultKeyDir, "ts--"+lower): blob,
		filepath.Join(home, DefaultAccountPath):          []byte(lower),
		filepath.Join(home, "pw.txt"):                    []byte("pw"),
	}
	for path, content := range files {
		if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(home, "pw.txt"), addr.String()
}

// Failures reported by the chain used to be printed and then ignored.
func Test_chainErrorExitsNonZero(t *testing.T) {
	home := t.TempDir()
	passwordFile, owner := writeTestKeystore(t, home)
	global := []string{"--rpcAddr", deadRPC, "--chainId", "moca_5151-1", "--evmRpcAddr", deadRPC, "--home", home, "-p", passwordFile}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"sp ls", []string{"sp", "ls"}, "fail to list SP"},
		{"bucket head", []string{"bucket", "head", "moca://missing"}, "no such bucket"},
		{"object head", []string{"object", "head", "moca://present/o"}, "no such object"},
		{"group head", []string{"group", "head", "--groupOwner", owner, "g"}, "no such group"},
		{"group head-member", []string{"group", "head-member", "--groupOwner", owner, owner, "g"}, "does not exist in the group"},
		{"bucket ls", []string{"bucket", "ls"}, "fail to get SP info to list bucket"},
		{"bucket update", []string{"bucket", "update", "--visibility", "private", "moca://present"}, "update bucket error"},
		{"bucket migrate", []string{"bucket", "migrate", "--dstPrimarySPID", "2", "moca://present"}, "migrate bucket error"},
		{"bucket rm", []string{"bucket", "rm", "moca://present"}, "delete bucket error"},
		{"bucket buy-quota", []string{"bucket", "buy-quota", "--chargedQuota", "100", "moca://present"}, "buy quota error"},
		{"object update", []string{"object", "update", "--visibility", "private", "moca://present/o"}, "update object visibility error"},
		{"object rm", []string{"object", "rm", "moca://present/o"}, "failed to delete object o"},
		{"object rm --recursive", []string{"object", "rm", "--recursive", "moca://present/dir/"}, "failed to delete 1 object(s)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFailed(t, runCLI(t, true, slices.Concat(global, tt.args)...), tt.want)
		})
	}
}

func Test_noBalanceExitsNonZero(t *testing.T) {
	home := t.TempDir()
	_, owner := writeTestKeystore(t, home)
	res := runCLI(t, true, "--rpcAddr", deadRPC, "--chainId", "moca_5151-1", "--evmRpcAddr", deadRPC, "--home", home, "bank", "balance", "--address", owner)
	if want := "The operator account have no balance, please transfer token to your account\n"; res.code != 1 || res.stdout != want || res.stderr != "" {
		t.Errorf("exit code = %d, stdout = %q, stderr = %q, want 1, %q, empty", res.code, res.stdout, res.stderr, want)
	}
}

func Test_successExitsZero(t *testing.T) {
	res := runCLI(t, false, "version")
	if res.code != 0 || !strings.Contains(res.stdout, "Moca Cmd Version") || res.stderr != "" {
		t.Errorf("exit code = %d, stdout = %q, stderr = %q, want 0 and the version", res.code, res.stdout, res.stderr)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = stdout }()
	fn()
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func Test_toCmdErr(t *testing.T) {
	var err error
	out := captureStdout(t, func() { err = toCmdErr(errors.New("boom")) })
	var exitErr cli.ExitCoder
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || err.Error() != "" {
		t.Fatalf("toCmdErr returned %#v, want an ExitCoder with code 1 and no message", err)
	}
	if out != "run command error: boom\n" {
		t.Errorf("printed %q", out)
	}

	// an error that was already reported is not printed again, even when wrapped
	for _, again := range []error{err, fmt.Errorf("upload: %w", err)} {
		var next error
		if out = captureStdout(t, func() { next = toCmdErr(again) }); out != "" {
			t.Errorf("printed %q for a reported error", out)
		}
		if !errors.As(next, &exitErr) || exitErr.ExitCode() != 1 {
			t.Errorf("toCmdErr returned %#v for a reported error, want an ExitCoder with code 1", next)
		}
	}
}
