# EtherGuard 配置与使用说明

本文档重点说明当前版本的配置文件结构、最新传输层配置，以及常见启动方式。

## 1. 当前版本新增了什么

当前版本在原有配置基础上新增了可配置传输层：

- `Transport.Protocol`
- `Transport.XOR.Key`
- `Transport.XOR.ObfuscateHeaders`
- `Transport.XOR.ReplayWindow`
- `Peers[].TransportEndpoint`

支持的传输协议：

- `udp_xor`
  - 当前默认值。若 `Transport.Protocol` 不写，默认就是它。
  - 只做 UDP 封装和 XOR 混淆，不提供成熟安全隧道协议等级的安全性。
- `tls_tunnel`
  - 传输层接口已预留，当前版本尚未实现。
- `dtls_tunnel`
  - 传输层接口已预留，当前版本尚未实现。

## 2. 启动方式

### 2.1 启动 Edge

```bash
./etherguard-go -mode edge -config /path/to/edge.yaml
```

### 2.2 启动 Super

```bash
./etherguard-go -mode super -config /path/to/super.yaml
```

### 2.3 查看示例配置

```bash
./etherguard-go -mode edge -example
./etherguard-go -mode super -example
```

说明：

- 当前默认传输协议是 `udp_xor`。
- 传输层已经抽象为统一接口，后续可以继续接入 `tls_tunnel`、`dtls_tunnel` 等协议实现。

## 2.5 SuperMode 的控制面与数据面

当前 `supermode` 的设计仍然保留了 full mesh 和智能选路能力：

- SuperNode 主要负责控制面：
  - 接收 `Register`
  - 汇总 `Ping/Pong` 延迟信息
  - 计算并分发 `NextHopTable`
  - 提供 peer/nhtable HTTP API
- EdgeNode 主要负责数据面：
  - 拿到 peer 列表后，会尝试与其他 edge 直接建立底层连接
  - 普通业务流量默认不会强制经过 super
  - 若不能直连，也会按 Floyd-Warshall 算出的下一跳经其他 edge 转发

可以把它理解成：

```text
        control plane
  Edge1 <--------> Super <--------> Edge2
     \                                /
      \------ direct data path ------/

      if direct path is bad/unavailable:
      data may go via another edge next-hop,
      not necessarily via Super
```

## 3. Edge 配置说明

## 3.1 Edge 根字段

常用字段如下：

| Key | 说明 |
| --- | --- |
| `NodeID` | 当前节点 ID |
| `NodeName` | 节点名称 |
| `IdentityPrivateKey` | 节点私钥。当前版本为了兼容内部身份与控制面，建议保留有效值，作为节点身份密钥 |
| `ListenPort` | 底层 UDP 监听端口 |
| `Transport` | 传输层配置，见下文 |
| `Interface` | TAP / sock / stdio 等接入方式 |
| `Peers` | 静态邻居列表 |
| `DynamicRoute` | Super / P2P / 延迟探测相关配置 |
| `DefaultTTL` | 二层转发默认 TTL |
| `ResetEndPointInterval` | 定期重置回初始 endpoint 的间隔 |
| `DisableAf` | 禁用 IPv4 / IPv6 |
| `AfPrefer` | 域名解析优先使用 IPv4 还是 IPv6 |

## 3.2 `Transport` 字段

```yaml
Transport:
  Protocol: udp_xor
  XOR:
    Key: replace-with-shared-xor-key
    ObfuscateHeaders: false
    ReplayWindow: 64
```

字段说明：

| Key | 说明 |
| --- | --- |
| `Transport.Protocol` | 当前可用值：`udp_xor`。预留值：`tls_tunnel`、`dtls_tunnel` |
| `Transport.XOR.Key` | `udp_xor` 模式必填。所有互通节点必须一致 |
| `Transport.XOR.ObfuscateHeaders` | 是否连传输头部一起做 XOR |
| `Transport.XOR.ReplayWindow` | 重放窗口大小，默认 64 |

说明：

- `udp_xor` 模式下，`Transport.XOR.Key` 必填。
- `udp_xor` 不是安全协议，只适合实验、混淆、联调。
- `tls_tunnel` / `dtls_tunnel` 的配置接口已经预留，但当前版本还未实现数据面。

## 3.3 `Peers` 字段

```yaml
Peers:
  - NodeID: 2
    PeerKey: dHeWQtlTPQGy87WdbUARS4CtwVaR2y7IQ1qcX4GKSXk=
    SharedKey: juJMQaGAaeSy8aDsXSKNsPZv/nFiPj4h/1G70tGYygs=
    EndPoint: 10.0.0.2:3002
    TransportEndpoint: 10.0.0.2:3002
    PersistentKeepalive: 30
    Static: true
```

字段说明：

| Key | 说明 |
| --- | --- |
| `NodeID` | 对端节点 ID |
| `PeerKey` | 对端公钥。当前控制面和身份匹配仍会使用 |
| `SharedKey` | 共享密钥。当前主要用于节点间身份/控制面兼容与静态配对 |
| `EndPoint` | 原有 endpoint 字段 |
| `TransportEndpoint` | 新字段。若填写，优先于 `EndPoint` |
| `PersistentKeepalive` | 保活间隔 |
| `Static` | 是否关闭 roaming，并定期回到初始 endpoint |

建议：

- 新配置优先写 `TransportEndpoint`。
- 为了兼容旧版本或旧脚本，也可以同时保留 `EndPoint`。

## 3.4 `Interface` 字段

`Interface.IType` 决定 VPN 收到的二层数据最终接到哪里，常见值：

- `tap`
- `dummy`
- `stdio`
- `udpsock`
- `tcpsock`
- `unixsock`
- `unixgramsock`
- `unixpacketsock`
- `fd`
- `vpp`

常见最小示例：

```yaml
Interface:
  IType: tap
  Name: eg0
  MacAddrPrefix: "AA:BB:CC:DD"
  MTU: 1404
  IPv4CIDR: 10.23.0.1/24
```

如果是 `udpsock` / `tcpsock` / `unixsock` 这类接口模式，仍然需要按原有规则填写 `RecvAddr` / `SendAddr`。

## 4. Super 配置说明

## 4.1 Super 根字段

常用字段如下：

| Key | 说明 |
| --- | --- |
| `NodeName` | Super 节点名称 |
| `IdentityPrivateKeyV4` | IPv4 通信用私钥 |
| `IdentityPrivateKeyV6` | IPv6 通信用私钥 |
| `ListenPort` | 底层 UDP 监听端口 |
| `Transport` | 传输层配置 |
| `ListenPort_EdgeAPI` | Edge API HTTP 端口 |
| `ListenPort_ManageAPI` | Manage API HTTP 端口 |
| `API_Prefix` | HTTP API 前缀 |
| `Peers` | 受管 Edge 节点列表 |
| `RePushConfigInterval` | 周期性重新推送配置 |
| `SendPingInterval` | Edge 间延迟探测参数 |
| `PeerAliveTimeout` | 判定离线时间 |

## 4.2 Super 的 `Transport`

Super 和 Edge 一样使用同一套 `Transport` 结构：

```yaml
Transport:
  Protocol: udp_xor
  XOR:
    Key: your-shared-key
    ObfuscateHeaders: false
    ReplayWindow: 64
```

要求：

- Super 与所有 Edge 的 `Transport.Protocol` 必须一致。
- 若使用 `udp_xor`，`Transport.XOR.Key` 也必须一致。
- 建议全网统一一种传输协议，不要混用。

## 4.3 Super 的 `Peers`

```yaml
Peers:
  - NodeID: 1
    Name: edge-01
    PeerKey: ZqzLVSbXzjppERslwbf2QziWruW3V/UIx9oqwU8Fn3I=
    SharedKey: iPM8FXfnHVzwjguZHRW9bLNY+h7+B1O2oTJtktptQkI=
    AdditionalCost: 10
    SkipLocalIP: false
    EndPoint: 1.2.3.4:3001
    TransportEndpoint: 1.2.3.4:3001
```

字段说明：

| Key | 说明 |
| --- | --- |
| `NodeID` | Edge 节点 ID |
| `Name` | 节点名称 |
| `PeerKey` | 节点公钥 |
| `SharedKey` | 预共享密钥 |
| `AdditionalCost` | 绕路成本 |
| `SkipLocalIP` | 打洞时是否忽略 Edge 上报的本地地址 |
| `EndPoint` | 原有 endpoint |
| `TransportEndpoint` | 新字段，优先于 `EndPoint` |
| `ExternalIP` | 无 NAT Reflection 等特殊场景下手工指定外网地址 |

## 5. 推荐配置示例

## 5.1 UDP XOR 模式 Edge

```yaml
NodeID: 1
NodeName: edge-01
IdentityPrivateKey: "base64-private-key"
ListenPort: 3001

Transport:
  Protocol: udp_xor
  XOR:
    Key: "same-key-on-all-nodes"
    ObfuscateHeaders: false
    ReplayWindow: 64

Interface:
  IType: tap
  Name: eg0
  MacAddrPrefix: "AA:BB:CC:DD"
  MTU: 1404

DefaultTTL: 200
AfPrefer: 4

Peers:
  - NodeID: 2
    PeerKey: "base64-public-key"
    SharedKey: "base64-psk"
    TransportEndpoint: "203.0.113.2:3002"
    PersistentKeepalive: 30
    Static: true
```

注意：

- 这个模式不会提供成熟安全隧道协议的安全保证。
- 更适合实验、抓包、协议调试、被动混淆。

## 5.2 UDP XOR 模式 Super

```yaml
NodeName: super-01
IdentityPrivateKeyV4: "base64-private-key-v4"
IdentityPrivateKeyV6: ""
ListenPort: 3000
ListenPort_EdgeAPI: "3000"
ListenPort_ManageAPI: "3000"
API_Prefix: "/eg_api"

Transport:
  Protocol: udp_xor
  XOR:
    Key: "same-key-on-all-nodes"
    ObfuscateHeaders: false
    ReplayWindow: 64

Peers:
  - NodeID: 1
    Name: edge-01
    PeerKey: "base64-public-key"
    SharedKey: "base64-psk"
    TransportEndpoint: "203.0.113.10:3001"
```

## 5.3 SuperMode + UDP XOR + TAP 示例文件

仓库里已经提供了可直接参考的完整示例：

- `example_config/super_mode/EgNet_super_udp_xor.yaml`
- `example_config/super_mode/EgNet_edge001_tap_udp_xor.yaml`
- `example_config/super_mode/EgNet_edge002_tap_udp_xor.yaml`

这组示例的特点：

- 运行模式是 `supermode`
- 底层传输是 `udp_xor`
- edge 接口类型是 `tap`
- edge 会在启动时把 `IPv4CIDR` / `IPv6CIDR` / `IPv6LLPrefix` 配到 tap 设备上

启动前建议至少修改这些字段：

- `Transport.XOR.Key`
- `Interface.Name`
- `ListenPort`
- `DynamicRoute.SuperNode.EndpointV4`
- `DynamicRoute.SuperNode.EndpointEdgeAPIUrl`

如果 Super 的 peer 条目要主动拨 edge，也建议补：

- `Peers[].TransportEndpoint`

## 5.4 TAP 模式注意事项

`Interface.IType: tap` 与 `stdio` / `dummy` 不同，它会真正创建 Linux tap 设备并配置地址。

因此通常需要：

- Linux 环境
- 可用的 `/dev/net/tun`
- `ip` 命令
- 足够权限，例如 root 或具备 `CAP_NET_ADMIN`

如果只是想在本机快速联调协议连通性，建议先用 `stdio` 示例。
如果要让系统里真实出现一个带 IP 的虚拟接口，再切到 `tap` 示例。

## 6. 使用建议

### 6.1 什么时候用 `udp_xor`

适合：

- 实验传输架构
- 验证“底层传输可替换”
- 需要简单混淆而不是正式加密
- 调试控制面和二层转发逻辑

不适合：

- 需要强安全性的场景
- 直接替代成熟安全隧道协议的生产部署

### 6.2 什么时候等 `tls_tunnel` / `dtls_tunnel`

适合：

- 你需要后续接入更标准化的加密隧道
- 你希望在同一套配置模型下切换不同底层承载
- 你希望保留当前项目的二层转发和路由逻辑，但底层换成 TLS/DTLS

## 7. 迁移建议

从老配置迁移到新配置时，建议按下面做：

1. 先把配置字段替换成新的通用命名：

- `IdentityPrivateKey`
- `IdentityPrivateKeyV4`
- `IdentityPrivateKeyV6`
- `PeerKey`
- `SharedKey`

2. 再把 peer 的地址逐步改成：

```yaml
TransportEndpoint: "host:port"
```

3. 然后统一明确写出：

```yaml
Transport:
  Protocol: udp_xor
  XOR:
    Key: "same-key-on-all-nodes"
```

## 8. 常见问题

### 8.1 不写 `Transport.Protocol` 会怎样

默认按 `udp_xor` 处理。

### 8.2 `TransportEndpoint` 和 `EndPoint` 都写了，谁生效

`TransportEndpoint` 优先。

### 8.3 还能用历史的状态查看方式看运行信息吗

不能。当前项目已经不再暴露旧的状态查询套接字接口。

### 8.4 `udp_xor` 下还要不要保留 `IdentityPrivateKey` / `PeerKey` / `SharedKey`

当前版本建议保留：

- `IdentityPrivateKey` / `IdentityPrivateKeyV4` / `IdentityPrivateKeyV6` 仍参与内部身份与兼容流程
- `PeerKey` / `SharedKey` 仍参与现有控制面和 peer 信息匹配

它们在 `udp_xor` 下不代表成熟隧道协议的数据面安全能力。
