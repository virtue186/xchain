`xchain` 是一个使用 Go 语言从零开始构建的极简区块链项目。

## 实现的功能:

- **账户模型**: 采用基于 ECDSA (P-256) 的公私钥对来创建和管理账户，并从公钥生成唯一的链上地址。
- **状态管理**: 使用 LevelDB 作为底层存储引擎，通过一个专门的状态模块持久化地记录每个账户的余额（Balance）和交易次序（Nonce）。
- **交易处理**: 支持构建、签名、验证和广播交易。交易信息包含了发送方、接收方、金额和 Nonce。交易在被处理前会通过签名进行验证。
- **区块链核心**: 实现了标准的区块（Block）和区块头（Header）数据结构。区块通过存储前一区块哈希（PrevBlockHash）的方式链接起来，形成一条不可篡改的线性链表。
- **P2P 网络**: 节点之间通过 TCP 长连接进行通信。节点启动后可以拨号连接到其他对等节点，并能通过一个事件通道 `peerCh` 感知新加入的节点。
- **区块同步**: 节点间可以请求和发送区块数据。当一个节点发现自己的高度低于对等节点时，会主动请求区块。实现了一次请求多个区块的批量同步逻辑，并能在接收完一批后持续请求下一批，直到追上最新高度。
- **交易池**: 设有一个交易池用于暂存网络中待确认的交易。交易池有最大容量限制，并且在新区块被确认后会从中移除已被打包的交易。
- **JSON-RPC API**: 提供了一个标准的 JSON-RPC 2.0 接口，允许外部应用通过 HTTP 请求查询账户状态 (`get_account_state`) 和提交原始交易 (`send_raw_transaction`)。
- **命令行客户端 (CLI)**: 配套提供了一个功能强大的命令行工具 `xchain-cli`，封装了对 RPC 接口的调用，可用于创建账户、查询余额和发起转账。

## 待完善的功能:

- **共识机制**: 当前的共识模型非常基础，本质上是一种权威证明（Proof-of-Authority）。由启动时提供了私钥的节点作为唯一的验证者，按固定的时间间隔（`BlockTime`）打包交易并创建新区块。这缺乏去中心化网络中应有的竞争和容错机制。
- **未加入树结构**: 项目未使用默克尔树（Merkle Tree）来组织交易。当前是将整个交易列表序列化后进行一次性哈希，这使得轻客户端无法高效地验证单笔交易的存在性。同时，区块链本身也仅是线性的数组结构，无法容纳和处理网络分叉。
- **智能合约**: 代码中包含了一个简单的、基于栈的虚拟机（VM）实现，但它并未被集成到交易处理的核心流程中。因此，系统目前不支持部署和执行智能合约。
- **序列化机制**:最初使用Gob进行序列化，在命令行客户端传输序列化数据时出现了某些BUG。因此暂时采用json进行序列化，后续考虑升级其它序列化方式。

## 如何运行

### 环境要求

- Go 1.20 或更高版本

### 启动网络

项目内置了一个便捷的启动脚本。直接运行根目录下的 `main.go` 即可一键启动一个包含三个节点的本地测试网络。

```cmd
go run main.go
```

该命令会执行以下操作：

1. 启动一个**验证者节点**，监听 `127.0.0.1:3000` (P2P) 和 `127.0.0.1:8000` (RPC)。这个节点拥有私钥，负责创建新的区块。
2. 启动两个**普通节点**（A 和 B），分别监听 `4000` 和 `5000` 端口 (P2P)，以及 `8001` 和 `8002` 端口 (RPC)。
3. 普通节点 A 和 B 会自动连接到验证者节点，并开始同步区块数据。
4. 数据库文件会分别存储在 `./db/node_127.0.0.1:xxxx` 目录下。

您将看到类似以下的日志输出，表示网络已成功运行：

```
Starting blockchain nodes...
node=127.0.0.1:3000 module=blockchain msg="database empty, adding genesis block"
node=127.0.0.1:3000 module=blockchain msg="add block" hash=b8a52fcffdac0ffef98239b4b3188187507e746633cadd7061440987c6802434 height=0 transaction=0
node=127.0.0.1:3000 msg="blockchain is new, performing genesis allocation"
node=127.0.0.1:3000 msg="genesis allocation" address=d55eff4e8c6e1e15740ccf223828cf217d694118 balance=1000000
node=127.0.0.1:3000 msg="starting node..."
node=127.0.0.1:3000 msg="starting broadcast service"
node=127.0.0.1:3000 module=api msg="starting API server" listenAddr=127.0.0.1:8000
```

> 节点数据被持久化存储在本地的 `./db` 目录中。节点重启时，会通过 `loadHeaders` 函数自动加载已有区块数据，从上次停止的高度无缝续传。

### 使用 xchain-cli交互

在网络运行后，您可以打开一个新的终端窗口，使用 `xchain-cli` 工具与区块链进行交互。

1. 创建账户

   ```cmd
   go run ./cmd/xchain-cli account new
   ```

2. 查询余额（d55eff4e8c6e1e15740ccf223828cf217d694118是创世块中预分配了资金的账户）

   ```cmd
   go run ./cmd/xchain-cli balance d55eff4e8c6e1e15740ccf223828cf217d694118
   ```

3. 发起转账

   ```cmd
   # 替换 <SENDER_PRIVATE_KEY> 为创世账户的私钥 (可以修改genesis.json自行设定)
   # 替换 <RECIPIENT_ADDRESS> 为您创建的新地址
   go run ./cmd/xchain-cli transfer \
     --from <SENDER_PRIVATE_KEY> \
     --to <RECIPIENT_ADDRESS> \
     --amount 100
   ```