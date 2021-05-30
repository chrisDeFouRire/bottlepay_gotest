# Bottle Pay test

I'm a Backend Developer, so I'll limit myself to backend stuff. Also, because BottlePay is a mobile app, it makes sense to me to provide an API-only entry.

This document presents my entry as the set of modifications I've performed in this repository.

## Refactor the mock Wallet/Exchange service 

First I chose to use a mono-repo to hold the various microservices. It makes it easier to share model definition for instance, and devops can be much easier too:

- only one CI/CD pipeline, simpler build
- only one docker image to host (different `cmd` though)
- easier synchronization: when you're deploying v1.43 of two different micro services, you know they're from the exact same build and they should work well together

So I've extracted the Cobra `rootCmd` into its own `data` command. The parameters remain the same.
This way it will be easier to add another microservice, like the tracker service.

## Add a User model and UserStore

I've added a User model in `model.go`. A User is identified by his ID, and holds references to the linked Custodian IDs.

I've chosen to implement a UserStore interface, because the first implementation will be a memory based fake datasource. A database-driven implementation would make more sense of course.
The UserStore functions has context.Context and errors because any UserStore but the most basic will need them.

In our FakeUserStore, User 1 has linked Custodians 1,2,3 and 4.

I've added unit tests for the UserStore in `store/userStore_test.go`.

## Add a CustodianSvc to collect the custodians data from HTTP requests

To perform HTTP requests to the mock Wallet/Exchange service, I'm using a CustodianSvc service. It can fetch any number of custodians by IDs.

The first implementation is not concurrent (each custodian is fetched in order) but an alternative implementation could perform the HTTP requests in parallel to reduce latency.

I've added simple integration tests for the CustodianSvc in `service/custodianSvc_test.go`. They require the mock service to be running on localhost:9999 though.

## Enrich the model package

To answer requests, I need filter and aggregation functions. I've chosen to put them in the model package as these functions relate directly to the model.

I've added an AssetList type to make it easier to handle Asset lists when adding Assets or Transactions. I've also added TransactionType's to make it easier to filter transactions.

I've also added AssetExchanges to track exchanges of asset in the same custodian.

Finally I've added simple unit tests in `model/aggregation_test.go`.

## Secure HTTP routes

I would normally add authentication on such a service. A JWT bearer token feels appropriate in this case.

- the mobile app calls a "login" route of an Auth service and receives a JWT token with claims (among them, the ID of the user). Could be AWS Cognito, Auth0 or a custom service
- the mobile app adds the token to its requests to the tracker service, usually in the `Authentication` HTTP header
- the tracker service can verify the token (issuer, signature and expiry) to certify it's from the Auth service, then use the JWT claims to get the identity of the user, and check the IDs are the same

I didn't implement this because of the development overhead, and because it would make testing more cumbersome.

## Add an HTTP route to aggregate the custodians holdings 

`GET /user/{id}/holdings`

Now it's just a matter of getting the user from the user store, getting his custodians data, aggregating his holding, and sending the json result.

```
$ curl -v http://0.0.0.0:9998/user/1/holdings
[{"code":"BTC","balance":"131.72086218"},{"code":"GBP","balance":"939940.97444372"}]
```

I've added a 30s timeout to fetch the data from the fake service.

I'd never used chi before, it's nice and powerful!

## Add an HTTP route to get the list of custodians for a user

At this point, I'm dealing with spec uncertainty: the specs don't seem precise enough to determine exactly what is expected from me, so I'll start going uphill thinking about what the abstract feature set should be. In real life, I'd turn to Slack to ask for more information while reading the specs :-)

I know I would need to access the list of custodians linked with the user, so I'll add that route. `GET /user/{id}`

I also started adding HTTP integration tests in `src/cmd/track_test.go` using `github.com/gavv/httpexpect/v2`, and made it easy to run the two services in docker (to run the integration tests).

## Add a transactions route for a custodian, with optional filtering

`GET /user/{id}/custodian/{custId}/transactions?type=[0-3]&summary`

I think we'll need to get a list of transactions for a custodian, filtered or unfiltered... So I addedd this route, and an integration test for it. This integration mandates to run against the generator with docker-compose, with --time=0 and with the provided data/state.json file mounted in the generator container. Otherwise the test data won't match.

This route provides easy external deposits and withdrawals lists, as well as summaries.

The following example shows the total withdrawal amounts for custodian 2, grouped by asset. Removing the `summary` parameter returns the list of transactions instead.

```
$ curl 'http://localhost:9998/user/1/custodian/2/transactions?type=1&summary'
[{"code":"BTC","balance":"7.6262799625"},{"code":"GBP","balance":"43046.9044724478"}]

$ curl 'http://localhost:9998/user/1/custodian/2/transactions?type=1'        
[{"id":8,"asset":"GBP","amount":"12246.121224612","direction":"OUT"},{"id":20,"asset":"BTC","amount":"4.6265955702","direction":"OUT"},{"id":25,"asset":"BTC","amount":"2.5370189581","direction":"OUT"},{"id":29,"asset":"BTC","amount":"0.4626654342","direction":"OUT"},{"id":40,"asset":"GBP","amount":"5265.0911534763","direction":"OUT"},{"id":43,"asset":"GBP","amount":"25535.6920943595","direction":"OUT"}]
```

Type 0 is Deposit, Type 1 is Withdrawal.

## Internal and External asset exchanges

For these last 2 requests, the requirements are even fuzzier. It could be argued that the above route already answers the requested data, albeit in a basic form. It can return the internal or external transactions, along with summaries.

However, I've added GetAssetExchanges() while thinking about this requirement. It's not used in routes though.

For this coding test, I'll choose to stop here, as the UX would be the determining factor, and I'm not doing UX. That's what teams are for :-)

## Next steps

I've already mentioned authentication which is missing from my entry for obvious reasons. 

Next I would definitely add caching to the Tracker service: it's getting data from foreign sources, which could have rate limiting, could be down, could be slow, etc... and caching the returned data in a Redis instance (LRU) would make plenty of sense. 

I'll implement caching before making a concurrent version of the custodians HTTP requester. Caching will limit the absence of concurrent HTTP requests and provide other benefits.

I also wanted to spend some time and provide a Helm chart and deploy it on a K8S cluster of mine, and you can ask me to do it, but I think I've already spent my time envelope, one or two days.

## Wrap up

All in all, I'm not sure if I'm not overthinking everything in this test :-) But it's a test.

I hope you'll enjoy my entry, and I want to thank you for making me try `Chi`, I already love it! and also `httpexpect` for testing. No matter what experience we have, there's always something new to learn from others!

## How to run

In my latest commit I've switched the time parameter of the generator back to 0, to turn data generation off, to make test data stable. I've included a `data/state.json` file in the commit to ensure the development environement is stable and self-contained.

`docker-compose up -d` should build and start everything. The two containers reuse the same docker image layers, so the build is pretty fast.

`go test ./...` will run all my tests against the running services.

```
$ go test ./... -count=1
?       github.com/bottlepay/portfolio-data     [no test files]
ok      github.com/bottlepay/portfolio-data/cmd 0.487s
ok      github.com/bottlepay/portfolio-data/model       0.456s
ok      github.com/bottlepay/portfolio-data/service     1.056s
ok      github.com/bottlepay/portfolio-data/store       0.460s
```
