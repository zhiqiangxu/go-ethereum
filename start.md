## 启动说明

### 1. 编译二进制文件
```buildoutcfg
make all
```

在./build/bin 目录下可以看到下列二进制文件
```buildoutcfg
abidump
abigen
bootnode
checkpoint-admin
clef
devp2p
ethkey
evm
faucet
geth
p2psim
puppeth
rlpdump

```

### 2.环境准备
以本地启动四个节点为例
```
node1
node2
node3
node4
```
将geth 拷贝到每个文件夹下，执行：
```buildoutcfg
./geth --datadir ./data account new
```
创建新账户

### 3. 设置创世区块
Clique的共识节点设置在区块头的extraData中，因此创世块需要进行配置，当然后续也可以增减共识节点。可以使用puppeth工具来进行交互式配置。

```
> puppeth
+ — — — — — — — — — — — — — — — — — — — — — — — — — — — — — -+
| Welcome to puppeth, your Ethereum private network manager |
| |
| This tool lets you create a new Ethereum network down to |
| the genesis block, bootnodes, miners and ethstats servers |
| without the hassle that it would normally entail. |
| |
| Puppeth uses SSH to dial in to remote servers, and builds |
| its network components out of Docker containers using the |
| docker-compose toolset. |
+ — — — — — — — — — — — — — — — — — — — — — — — — — — — — — -+
Please specify a network name to administer (no spaces, please)
> o3
Sweet, you can set this via — network=o3 next time!
```

选择 2
```buildoutcfg
What would you like to do? (default = stats)
 1. Show network stats
 2. Configure new genesis
 3. Track new remote server
 4. Deploy network components
> 2

```
选择 1
```buildoutcfg
What would you like to do? (default = create)
 1. Create new genesis from scratch
 2. Import already existing genesis
> 1

```
选择 2 Clique
```buildoutcfg
Which consensus engine to use? (default = clique)
 1. Ethash - proof-of-work
 2. Clique - proof-of-authority
> 2

```

```buildoutcfg
How many seconds should blocks take? (default = 15)
> 

```

增加可以出块的钱包地址，将第一步生成的四个地址加入
```buildoutcfg
Which accounts are allowed to seal? (mandatory at least one)
> 0x

```

填入需要预先分配代币的地址
```buildoutcfg
Which accounts should be pre-funded? (advisable at least one)
> 0x

```

```buildoutcfg
Should the precompile-addresses (0x1 .. 0xff) be pre-funded with 1 wei? (advisable yes)
> no
```
选择2
```buildoutcfg
What would you like to do? (default = stats)
 1. Show network stats
 2. Manage existing genesis
 3. Track new remote server
 4. Deploy network components
> 2

```

选择2
```buildoutcfg
 1. Modify existing configurations
 2. Export genesis configurations
 3. Remove genesis configuration
> 2
```

导出json文件
```buildoutcfg
Which folder to save the genesis specs into? (default = current)
  Will create mochain.json, mochain-aleth.json, mochain-harmony.json, mochain-parity.json
> 

```

生成的json文件：
```buildoutcfg
{
  "config": {
    "chainId": 59569,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "clique": {
      "period": 15,
      "epoch": 30000
    }
  },
  "nonce": "0x0",
  "timestamp": "0x605c499c",
  "extraData": "0x00000000000000000000000000000000000000000000000000000000000000007d84d07aca491a6ccbb1577eccbc5e18cb7de0880000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "gasLimit": "0x47b760",
  "difficulty": "0x1",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "alloc": {},
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```

将`period`改成`3`, `epoch`改成`200`，`gasLimit`改成`0x2625a00`。

在每个文件夹下执行：
```buildoutcfg
./geth --datadir ./data/ init o3.json 
```

### 4. 启动节点
假定四个节点的p2p端口为2000 ~ 2003,我们暂时只开放node1的http端口，默认为8545
首先通过下面的命令打印各个节点的公钥
```buildoutcfg
bootnode -nodekey ./data/geth/nodekey -writeaddress
```
也可以将节点列表放入<datadir>/geth/static-nodes.json文件中，让节点主动连接和重连：https://geth.ethereum.org/docs/interface/peer-to-peer

```buildoutcfg
[
    "enode://7b9a62ee9350e0d3a86dc29f97875542a3b0a7765c177218bcbcaa2bbb0da945feb87a137f510d6ac0c976456e0d9a624d2534298ed45e07fa455b55ebfa1832@127.0.0.1:2000",
    "enode://d64121d4de07d8acf82e65a8ac7e2e331d4ff77e29496433366570cb0f632f8a60e7e64dfc0853a9f6bb3880b0436df77c9108fbd9fe762980d17d7f1ec92289@127.0.0.1:2001",
    "enode://55f5d42ed95300427742a0cfb80a3143e2b32138f6ae3ce24fdb4393d32bc6040d79fb7cb0aba91352c00b4fcf83ba76df9a75cf3c9f6360740b19f60de57061@127.0.0.1:2002",
    "enode://aa8024d469b6428c169170569aff7f56def36ec419ee80d208ffeebec460c0ce009760683af6ea371eba452e9cdfc3dccd4aed5ec336f26baf2982e6085e6f05@127.0.0.1:2003"
]
```

启动命令
```buildoutcfg
./geth --datadir ./data --networkid 55661 --port 2000  --http --http.addr 0.0.0.0 --http.port 8545 --miner.gasprice 0 --mine --rpc.allow-unprotected-txs --allow-insecure-unlock --bootnodes 'enode://7b9a62ee9350e0d3a86dc29f97875542a3b0a7765c177218bcbcaa2bbb0da945feb87a137f510d6ac0c976456e0d9a624d2534298ed45e07fa455b55ebfa1832@127.0.0.1:2000,enode://d64121d4de07d8acf82e65a8ac7e2e331d4ff77e29496433366570cb0f632f8a60e7e64dfc0853a9f6bb3880b0436df77c9108fbd9fe762980d17d7f1ec92289@127.0.0.1:2001,enode://55f5d42ed95300427742a0cfb80a3143e2b32138f6ae3ce24fdb4393d32bc6040d79fb7cb0aba91352c00b4fcf83ba76df9a75cf3c9f6360740b19f60de57061@127.0.0.1:2002,enode://aa8024d469b6428c169170569aff7f56def36ec419ee80d208ffeebec460c0ce009760683af6ea371eba452e9cdfc3dccd4aed5ec336f26baf2982e6085e6f05@127.0.0.1:2003' --unlock 87ba503cce4ca532b3b31ffa67fbd32fa5409a60 console

./geth --datadir ./data --networkid 55661 --port 2001 --miner.gasprice 0 --mine --rpc.allow-unprotected-txs --bootnodes 'enode://7b9a62ee9350e0d3a86dc29f97875542a3b0a7765c177218bcbcaa2bbb0da945feb87a137f510d6ac0c976456e0d9a624d2534298ed45e07fa455b55ebfa1832@127.0.0.1:2000,enode://d64121d4de07d8acf82e65a8ac7e2e331d4ff77e29496433366570cb0f632f8a60e7e64dfc0853a9f6bb3880b0436df77c9108fbd9fe762980d17d7f1ec92289@127.0.0.1:2001,enode://55f5d42ed95300427742a0cfb80a3143e2b32138f6ae3ce24fdb4393d32bc6040d79fb7cb0aba91352c00b4fcf83ba76df9a75cf3c9f6360740b19f60de57061@127.0.0.1:2002,enode://aa8024d469b6428c169170569aff7f56def36ec419ee80d208ffeebec460c0ce009760683af6ea371eba452e9cdfc3dccd4aed5ec336f26baf2982e6085e6f05@127.0.0.1:2003' --unlock 77237beb384e47d3d1ca5b80b1dd9f3f02807784 console 

./geth --datadir ./data --networkid 55661 --port 2002 --miner.gasprice 0 --mine --rpc.allow-unprotected-txs --bootnodes 'enode://7b9a62ee9350e0d3a86dc29f97875542a3b0a7765c177218bcbcaa2bbb0da945feb87a137f510d6ac0c976456e0d9a624d2534298ed45e07fa455b55ebfa1832@127.0.0.1:2000,enode://d64121d4de07d8acf82e65a8ac7e2e331d4ff77e29496433366570cb0f632f8a60e7e64dfc0853a9f6bb3880b0436df77c9108fbd9fe762980d17d7f1ec92289@127.0.0.1:2001,enode://55f5d42ed95300427742a0cfb80a3143e2b32138f6ae3ce24fdb4393d32bc6040d79fb7cb0aba91352c00b4fcf83ba76df9a75cf3c9f6360740b19f60de57061@127.0.0.1:2002,enode://aa8024d469b6428c169170569aff7f56def36ec419ee80d208ffeebec460c0ce009760683af6ea371eba452e9cdfc3dccd4aed5ec336f26baf2982e6085e6f05@127.0.0.1:2003' --unlock ff05ec6f41b3879dd4d18985bc5958efebf33dff console 

./geth --datadir ./data --networkid 55661 --port 2003 --miner.gasprice 0 --mine --rpc.allow-unprotected-txs --bootnodes 'enode://7b9a62ee9350e0d3a86dc29f97875542a3b0a7765c177218bcbcaa2bbb0da945feb87a137f510d6ac0c976456e0d9a624d2534298ed45e07fa455b55ebfa1832@127.0.0.1:2000,enode://d64121d4de07d8acf82e65a8ac7e2e331d4ff77e29496433366570cb0f632f8a60e7e64dfc0853a9f6bb3880b0436df77c9108fbd9fe762980d17d7f1ec92289@127.0.0.1:2001,enode://55f5d42ed95300427742a0cfb80a3143e2b32138f6ae3ce24fdb4393d32bc6040d79fb7cb0aba91352c00b4fcf83ba76df9a75cf3c9f6360740b19f60de57061@127.0.0.1:2002,enode://aa8024d469b6428c169170569aff7f56def36ec419ee80d208ffeebec460c0ce009760683af6ea371eba452e9cdfc3dccd4aed5ec336f26baf2982e6085e6f05@127.0.0.1:2003' --unlock a369dfa4d618b98ebd61e8725b7782625963fb84 console

```
输入节点密码后，启动挖矿

测试转账：
```buildoutcfg
eth.sendTransaction({from:'87ba503cce4ca532b3b31ffa67fbd32fa5409a60', to:'a369dfa4d618b98ebd61e8725b7782625963fb84', value: web3.toWei(0.05, "ether")})
```

### 5. 新增共识节点
取得节点公钥后，以相同的networkid 启动节点 node5
```buildoutcfg
./geth --datadir ./data --networkid 55661 --port 2004 --miner.gasprice 0 --rpc.allow-unprotected-txs --bootnodes 'enode://7b9a62ee9350e0d3a86dc29f97875542a3b0a7765c177218bcbcaa2bbb0da945feb87a137f510d6ac0c976456e0d9a624d2534298ed45e07fa455b55ebfa1832@127.0.0.1:2000,enode://d64121d4de07d8acf82e65a8ac7e2e331d4ff77e29496433366570cb0f632f8a60e7e64dfc0853a9f6bb3880b0436df77c9108fbd9fe762980d17d7f1ec92289@127.0.0.1:2001,enode://55f5d42ed95300427742a0cfb80a3143e2b32138f6ae3ce24fdb4393d32bc6040d79fb7cb0aba91352c00b4fcf83ba76df9a75cf3c9f6360740b19f60de57061@127.0.0.1:2002,enode://aa8024d469b6428c169170569aff7f56def36ec419ee80d208ffeebec460c0ce009760683af6ea371eba452e9cdfc3dccd4aed5ec336f26baf2982e6085e6f05@127.0.0.1:2003' --unlock a369dfa4d618b98ebd61e8725b7782625963fb84 console
```

node1~ node4 在控制台执行：
```buildoutcfg
admin.addPeer("enode://1290aad1ef5b457e219668f02814883236ed805a2f1ac87188d44bac67e0908e5f48fa5bbb09d8f875c3641215d6a5710c31aabe875c943e7ab1f2c6cf28f33f@127.0.0.1:2004")
```

由共识节点发起新增节点交易
```buildoutcfg
> clique.propose(<node5钱包地址>, true)
```

node5的控制台中执行
```buildoutcfg
miner.start()
```