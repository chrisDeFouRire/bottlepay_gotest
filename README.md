# Bottle Pay test

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
