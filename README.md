# Caching Proxy Server

A simple CLI tool that implements a caching proxy server. This tool forwards HTTP requests to an actual server and caches the responses. If a request is made again, it returns the cached response instead of forwarding the request.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Example](#example)
- [Response Headers](#response-headers)

## Features
- Start a caching proxy server on a specified port.
- Forward requests to an origin server.
- Cache responses to improve performance for repeated requests.
- Indicate whether a response is served from cache or the origin server.
- Set a Time-To-Live (TTL) for cached responses.

## Installation
To install the caching proxy server, follow these steps:

1. Clone the repository:
```bash
   git clone https://github.com/yourusername/caching-proxy.git
   cd caching-proxy
```
2. Build the binary:

```bash
go build -o caching-proxy
```
3. (Optional) Move the binary to your PATH:
```bash
mv caching-proxy /usr/local/bin
```

## Usage
To start the caching proxy server, use the following command:

```bash
caching-proxy --port <number> --origin <url> --ttl <duration>
```

Parameters
- `--port`: The port on which the caching proxy server will run.
- `--origin`: The URL of the server to which the requests will be forwarded.
- `--ttl`: The time duration (in seconds) for which cached responses should remain valid. After this duration, cached responses will be considered expired.

## Example
To start the server on port 3000, forward requests to http://dummyjson.com, and set a TTL of 60 seconds, run:

```bash
caching-proxy --port 3000 --origin http://dummyjson.com --ttl 60
```
Now, if you make a request to http://localhost:3000/products, the caching proxy server will forward the request to http://dummyjson.com/products, return the response along with headers, and cache the response.

## Response Headers
The caching proxy server adds the following headers to indicate the source of the response:

- `X-Cache: HIT` - The response was served from the cache.
- `X-Cache: MISS` - The response was fetched from the origin server.
---
project url: https://roadmap.sh/projects/caching-server
