#### Goal

To get profit from price difference from different dex exchanges.

For eg:

TUR/DAI is trading at 1: 5.1 in Uniswap
TUR/DAI is trading at 1: 4.1 in Sushiswap

here,
difference = 5.1 - 4.1 = 1

1. Calculate profit

uniswapCost = transaction fee + platform fee
shusiswapCost = transaction fee + platform fee

profit = difference - (uniswapCost + shusiswapCost)

2. Check Conditions:

- Profit threshold should be there. Else, the transaction won't be worthy one.
    eg: if the profit is 0.012 DAI, the transaction is not worth it. In this case, it shouldn;t proceed
    forward.
- slipage and liquidity

3.

Action:
To
Buy in Sushiswap and,
Sell in Uniswap

The transaction has to be atomic (all or nothing)

#### Action Items

##### Setup

1. Create a trading wallet in metamask and export private key.
2. Connect to ethereum network using go-ethereum package. (Best connect to direct ethereum node)
Some of the public nodes are:
    <https://ethereum.publicnode.com> (Allnodes)
    <https://cloudflare-eth.com/v1/mainnet> (Cloudflare)
    <https://eth.llamarpc.com> (LlamaNodes)
    <https://eth-mainnet.public.blastapi.io> (Blast API)
    <https://public-eth.nownodes.io/> (NOWNodes)
    <https://eth.drpc.org/> (dRPC)

##### Interacting with dex

3. Generate a contract binding to Uniswap and Sushiswap using abigen.

##### Build arbitrage logic

4. Frequently check the price for arbitrage opportunity.

##### Execute the trade

- build and sign the transaction
-

2. Sign a transaction in the script and send it to blockchain using go-ethereum.

TODO: Find a websocket node for fetching swap events for pricing.

To get the pricing from uniswap v3:

steps:

- subscribe to logs using eth_subscribe
 topic: swap event (ie 0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67)
 pool address: WETH/USDT (i.e 0x11b815efB8f581194ae79006d24E0d814B7697F6)

 ```bash
 wscat -c wss://mainnet.infura.io/ws/v3/<api-key> -x '{"jsonrpc": "2.0", "id": 1, "method": "eth_subscribe", "params": ["logs", {"address": "0x8320fe7702b96808f7bbc0d4a888ed1468216cfd", "topics":["0xd78a0cb8bb633d06981248b816e7bd33c2a35a6089241d099fa519e361cab902"]}]}'
 ```
