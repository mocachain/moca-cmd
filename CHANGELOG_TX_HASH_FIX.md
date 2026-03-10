# Changelog - Transaction Hash Fix

## [2026-01-16] - Transaction Hash User Experience Improvement

### Problem
- `moca-cmd` 返回 EVM 交易哈希（0x...格式）
- `mocad query tx` 需要 Cosmos 交易哈希（大写十六进制，无 0x 前缀）
- 用户无法直接使用 `moca-cmd` 返回的哈希查询交易

### Added

#### 新命令: `moca-cmd tx find-cosmos-hash`
- 从最近的区块中查找 Cosmos 交易哈希
- 支持 `--blocks N` 参数指定搜索范围（默认 20 个区块）
- 显示每个交易的详细信息：
  - Cosmos 哈希
  - 交易大小
  - 查询命令示例
  - Base64 编码的交易数据（前 100 字符）

#### 新文件: `cmd/cmd_tx.go`
- 实现 `tx` 命令组
- 实现 `find-cosmos-hash` 子命令
- 包含 Cosmos 哈希计算逻辑

#### 新文件: `cmd/utils/hash.go`
- `GetCosmosHashFromTxBytes()` - 从交易字节计算 Cosmos 哈希
- `GetCosmosHashFromBase64()` - 从 Base64 编码计算 Cosmos 哈希
- `FormatTxHashOutput()` - 格式化哈希输出
- `NormalizeEvmHash()` - 规范化 EVM 哈希格式
- `NormalizeCosmosHash()` - 规范化 Cosmos 哈希格式

### Changed

#### `cmd/cmd_account.go`
- **修改**: `cmdTransfer()` 输出
- **新增**: 在交易成功后添加说明，告知用户返回的是 EVM 哈希
- **新增**: 提示用户如何查找 Cosmos 哈希

```go
fmt.Printf("transfer %s amoca to address %s succ, txHash: %s\n", amountStr, toAddr, txHash)
fmt.Printf("\nNote: This is an EVM transaction hash. To query with mocad, you need the Cosmos tx hash.\n")
fmt.Printf("The transaction will be included in the next block. You can find it by searching recent blocks.\n")
```

#### `cmd/cmd_bucket.go`
- **修改**: `cmdCreateBucket()` 输出
- **新增**: 在 bucket 创建成功后添加相同的 EVM 哈希说明

```go
fmt.Printf("make_bucket: %s \n", bucketName)
fmt.Println("transaction hash: ", txnHash)
fmt.Printf("\nNote: This is an EVM transaction hash. To query with mocad, you need the Cosmos tx hash.\n")
fmt.Printf("The transaction will be included in the next block. You can find it by searching recent blocks.\n")
```

#### `cmd/main.go`
- **新增**: 注册 `cmdTx()` 命令到主命令列表

### Testing

#### 新文件: `test/tx_hash_test.sh`
- 自动化测试脚本验证修复
- 测试场景：
  1. Transfer 输出包含 EVM 哈希说明
  2. 可以从最近区块找到 Cosmos 哈希
  3. Cosmos 哈希可以用于 `mocad query tx`
  4. EVM 哈希在交易数据中可见

### Documentation

#### 新文件: `TX_HASH_FIX_SUMMARY.md`
- 完整的问题分析和解决方案文档
- 使用示例和工作流程
- 未来改进建议

## Usage Examples

### 基本使用

```bash
# 1. 创建交易
$ moca-cmd bank transfer --toAddress 0x... --amount 1000000000000000000
transfer 1000000000000000000 amoca to address 0x... succ, txHash: 0x4c1e32f3...

Note: This is an EVM transaction hash. To query with mocad, you need the Cosmos tx hash.
The transaction will be included in the next block. You can find it by searching recent blocks.

# 2. 查找 Cosmos 哈希
$ moca-cmd tx find-cosmos-hash --blocks 5
Searching last 5 blocks (from height 2256)...

Block 2252, TX #0:
  Cosmos Hash: B1845069B7DB7B0208EF84B412DDBFE6F1CC5A6A24169C9F4E8ECEE8605DFF50
  TX Size: 381 bytes
  Query command: mocad query tx B1845069B7DB7B0208EF84B412DDBFE6F1CC5A6A24169C9F4E8ECEE8605DFF50

# 3. 查询交易
$ mocad query tx B1845069B7DB7B0208EF84B412DDBFE6F1CC5A6A24169C9F4E8ECEE8605DFF50
code: 0
...
```

### 查看所有最近交易

```bash
# 查看最近 20 个区块的所有交易
$ moca-cmd tx find-cosmos-hash --blocks 20
```

## Technical Details

### Cosmos Hash Calculation

```go
func calculateCosmosHash(txBytes []byte) string {
    hash := sha256.Sum256(txBytes)
    return strings.ToUpper(hex.EncodeToString(hash[:]))
}
```

### Hash Format Comparison

| Type | Format | Example | Usage |
|------|--------|---------|-------|
| EVM Hash | 0x + lowercase hex | `0x4c1e32f38757654e...` | `moca-cmd` 返回值 |
| Cosmos Hash | Uppercase hex | `B1845069B7DB7B02...` | `mocad query tx` 参数 |

## Breaking Changes

无破坏性变更。所有修改都是向后兼容的：
- 现有命令的输出格式保持不变，只是添加了额外的说明信息
- 新增的命令不影响现有功能

## Migration Guide

不需要迁移。用户可以：
1. 继续使用 EVM 哈希进行 EVM 相关操作
2. 使用新的 `tx find-cosmos-hash` 命令查找 Cosmos 哈希
3. 使用 Cosmos 哈希进行 `mocad query tx` 查询

## Future Improvements

1. **哈希转换 API**: 提供 EVM 哈希到 Cosmos 哈希的直接转换（如果能找到映射关系）
2. **自动查询选项**: 添加 `--wait-for-cosmos-hash` 标志自动等待并返回 Cosmos 哈希
3. **交易跟踪**: 实现交易状态跟踪功能
4. **批量查询**: 支持批量查询多个交易
5. **缓存机制**: 缓存最近的哈希映射以提高查询速度

## Contributors

- AI Assistant (Implementation)
- User (Requirements and Testing)

## References

- Cosmos SDK Transaction Hash: SHA256(tx_bytes)
- EVM Transaction Hash: Ethereum transaction hash format
- Related Issue: moca-cmd 返回的交易哈希无法用于 mocad query tx 查询
