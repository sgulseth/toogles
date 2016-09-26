# Toogles

[![Docker Repository on Quay](https://quay.io/repository/sgulseth/toogles/status "Docker Repository on Quay")](https://quay.io/repository/sgulseth/toogles)

Toogles, pronounced `/'tugels/`, aims to be a high performance self hosted feature flag service(Typo intended). It implements different strategies to serve your users different sets of flags. The only dependency are at the moment Redis, and the goal is to keep the requirements at a minimum.

## Getting started

Toogles requires two environment variables to be set when started

 * `REDIS_ADDR`: Which host and port are redis running on. ie: `REDIS_ADDR="127.0.0.1:6379"`.
 * `API_KEY`: The api key to manage and set new features through the [toogles-console](https://github.com/sgulseth/toogles-console).

 Optional environment variables:

 * `PORT`: Which port should toogles listen on. Defaults to `8080`. 
 * `CORS`: Which cors origins should be allowed. Defaults to `*`.

### Run with docker

```
docker pull quay.io/sgulseth/toogles
docker run -e "REDIS_ADDR=redis.something:6379" -e "API_KEY=..." -p 8080:8080 quay.io/sgulseth/toogles
```

The app should now be listening on port 8080

## Manage feature flags

To manage feature flags it is recommended to use the [toogles-console](https://github.com/sgulseth/toogles-console) application.

### Managing through the api

Documentation for this are coming soon