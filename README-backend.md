# Bottlepay Interview Task

## Background
A user of Bottlepay has the ability to hold funds from their native fiat currency (`EUR`, `GBP`, etc) and cryptocurrencies (`BTC`), as well as making exchanges and payments in those currencies.

However, there are a few use cases that Bottlepay currently does not account for:
- An investor may be forced to use multiple exchanges if they're interested in investing in alternative cryptocurrencies that Bottlepay does not support (such as Dogecoin).
- A power user may wish to hold the majority of their funds in a custodial wallet, and only keep a small portion within Bottlepay or an alternative exchange.

For power users and investors, keeping track of the total value of a portfolio held across multiple wallets and/or exchanges can be cumbersome, which is where a portfolio tracker that can consume data from multiple external sources can come in handy.

## The Task

You should design a multi-tenant system for tracking users' asset portfolios across multiple third-party wallets and exchanges that the user has linked to the system.

At a basic level, the system should track:
- The total amount held for each currency/asset across all wallets and exchanges.
- Movement of funds between wallets and exchanges.
- External deposits and withdrawals
- Exchanges of funds between assets on an exchange.

You may use any method or technology you wish to accomplish this task, but don't spend more than a day or two on it.

Your solution can be in any form you wish, but it should be relevant to the position you're applying for. For example, we would not expect a primarily React powered UX solution for a backend development position.

If your solution extends to writing code, we have included a mock data service that can dummy as an integration against multiple exchanges.

## Mock Wallet/Exchange Data

Included in this package is a small application that will generate and serve dummy exchange-like data for you to consume.

For your convenience, it is available as a docker-compose service. Simply run `docker-compose up -d` and it will listen on port `9999`. All CORS Origins and Headers are allowed.

The service contains dummy `Custodians`, which represent a third-party exchange or wallet that holds one or more of a user's assets. By default, 4 `Custodians` will be created
to represent the following third-party sources belonging to a single user:
- Bitcoin Wallet (ID 1)
- Bottlepay (ID 2)
- Coinbase (ID 3)
- Binance (ID 4)

Each `Custodian` has a list of `Transactions`. A `Transaction` can represent one of the following operations:
- A deposit from an external source.
- A withdrawal to an external source.
- A transfer to another exchange or wallet
- An exchange between another asset (e.g. GBP to BTC).

By default, a new operation is executed every 10 seconds. This can be configured in the `docker-compose.yaml` file. The data generated is persisted between runs.

### `GET /custodian/{id}`

This endpoint retrieves an individual `Custodian` by its ID. The output contains the same information
as the list endpoint, but also includes individual transactions:

```json
{
	"id": 2,
	"assets": [
		{
			"code": "BTC",
			"balance": "15.99455449"
		},
		{
			"code": "GBP",
			"balance": "742999.5341094"
		}
	],
	"transactions": [
		{
			"id": 1,
			"asset": "GBP",
			"amount": "9000.0009",
			"direction": "OUT"
		},
		{
			"id": 2,
			"asset": "BTC",
			"amount": "0.5000000005",
			"direction": "IN",
			"related_custodian_id": 3,
			"related_custodian_transaction_id": 1
		},
		{
			"id": 3,
			"asset": "GBP",
			"amount": "7280.000728",
			"direction": "OUT",
			"related_custodian_id": 4,
			"related_custodian_transaction_id": 1
		},
		{
			"id": 4,
			"asset": "BTC",
			"amount": "0.6300000006",
			"direction": "OUT"
		},
		{
			"id": 5,
			"asset": "BTC",
			"amount": "1.8753000019",
			"direction": "OUT",
			"related_custodian_id": 2,
			"related_custodian_transaction_id": 6
		},
		{
			"id": 6,
			"asset": "GBP",
			"amount": "75012.000076",
			"direction": "IN",
			"related_custodian_id": 2,
			"related_custodian_transaction_id": 5
		}
	]
}
```

### `GET /generate`

Generate a new random event. The `count` query param can be passed to generate up to 1000 events at a time.