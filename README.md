# httpfromtcp

Building an HTTP server from scratch over raw TCP. No net/http for the core parts.

I wrote this to understand what happens behind the scenes when a web server
handles requests. Good for CV work and for anyone who wants to see how HTTP
works at the TCP level.

## What's inside

- Manual HTTP request parser (state machine that walks through request line,
  headers, and body byte by byte)
- Response writer with chunked transfer encoding and trailer support
- Header parser with validation and case-insensitive keys
- TCP server loop with connection handling
- Example server with a few endpoints

## Project layout

```
cmd/
  httpserver/    - example HTTP server (port 42069)
  tcplistener/   - raw TCP listener that parses requests
  udplistener/   - UDP listener example
  udpsender/     - UDP sender example
internal/
  server/        - server loop, handler types, error helpers
  request/       - HTTP request parsing
  response/      - HTTP response writing
  headers/       - header parsing and validation
```

## Running

```sh
go run ./cmd/httpserver/
```

Server starts on port 42069. Hit it with curl or a browser.

### Endpoints

| Path | What it does |
|------|-------------|
| `/` | Returns a simple HTML page |
| `/error` | Returns a 400 with JSON response |
| `/stream` | Streams random chunks with delay |
| `/trailers` | Proxies httpbin response with trailers |
| `/stream/media` | Streams a PNG image in chunks |

There is also a TCP listener on port 3000 (`go run ./cmd/tcplistener/`) that
parses incoming HTTP requests and prints them.

## Running tests

```sh
go test ./...
```

## License

MIT
