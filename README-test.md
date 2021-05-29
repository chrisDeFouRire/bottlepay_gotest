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

## Add a CustodianSvc to collectd the custodians data from HTTP requests

To perform HTTP requests to the mock Wallet/Exchange service, I'm using a CustodianSvc service. It can fetch any number of custodians by IDs.

The first implementation is not concurrent (each custodian is fetched in order) but an alternative implementation could perform the HTTP requests in parallel to reduce latency.

I've added simple integration tests for the CustodianSvc in `service/custodianSvc_test.go`. They require the mock service to be running on localhost:9999 though.

## Add an HTTP route to aggregate the custodians data 

I first need to identify and authenticate the user before aggregating his data.

For this, I'll rely of JWT: I'll assume there's an authentication service somewhere and it will provide the user with a JWT token once authenticated. This kind of service would receive credentials (user/password, Google or Apple token) from the mobile app, verify them and return a JWT token to the app. This token would then allow the mobile app to call the Tracker service.

This way the Tracker service can check the user's identity against the JWT token in HTTP requests. There's a clean separation of concerns, the tracker service doesn't need any access to user's private data.

