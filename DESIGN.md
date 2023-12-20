# Design

Describing the design goals and keeping track of TODOs.


## What do we need?

 - testable, stable code that is using the go stdlib (or extensions) without dependencies
 - Authentication (various types that alter the request or perform pre request authentication)
 - Rate Limiting (limit the outgoing requests based on a token bucket)
 - Programmable Retries per client instance (want to decide when to retry and how often)
 - Backoff functionality (constant and exponential backoff is what we are using now)
 - post response hooks for metrics
 - pre request hooks for altering requests by adding headers etc.


### Maybe?

- Plugin system to alter the behaviour of the client?


 ## TODO:
 - [ ] add custom errors where fmt.Errof is returned
 - [ ] see if the behavioural error pattern is useful for retries