# Design

Describing the design goals and keeping track of TODOs.

## What do we need?
 - Authentication (various types that alter the request or perform pre request authentication)
 - Rate Limiting (limit the outgoing requests based on a token bucket)
 - Programmable Retries per client instance (want to decide when to retry and how often)
 - Backoff functionality (constant and exponential backoff is what we are using now)
 - post response hooks for metrics
 - pre request hooks for altering requests by adding headers etc.

 ## TODO:
 - [ ] add custom errors where fmt.Errof is returned
 - [ ] see if the behavioural error pattern is useful for retries
 - [ ] pull retry count from the backoff into the retry policy to make it more understandable and shift responsibility to the retry policy