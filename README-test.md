# Bottle Pay test

This document presents my entry as the set of modifications I've performed in this repository.

## Refactor the mock Wallet/Exchange service 

First I chose to use a mono-repo to hold the various microservices. It makes it easier to share model definition for instance, and devops can be much easier too:

- only one CI/CD pipeline, simpler build
- only one docker image to host (different `cmd` though)
- easier synchronization: when you're deploying v1.43 of two different micro services, you know they're from the exact same build and they should work well together

So I've extracted the Cobra `rootCmd` into its own `data` command. The parameters remain the same.
This way it will be easier to add another microservice, like the tracker service.

