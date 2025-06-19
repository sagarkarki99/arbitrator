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

##### Interacting with dex

3. Generate a contract binding to Uniswap and Sushiswap using abigen.

##### Build arbitrage logic

4. Frequently check the price for arbitrage opportunity.

##### Execute the trade

- build and sign the transaction
-

2. Sign a transaction in the script and send it to blockchain using go-ethereum.
