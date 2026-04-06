# discoverd

discoverd is a lightweight UDP multicast service discovery tool.

## What It Is

discoverd runs in two modes:

- server: listens for discovery requests and replies with server ID and IP
- client: sends discovery requests and prints discovered servers

Discovery messages use timestamp/nonce checks and HMAC signatures with a shared secret.

## Why Use It

- Quickly find services in a local network without hardcoded IPs
- Keep discovery simple for internal tools, labs, and local environments
- Run the same CLI flow on Linux, macOS, and Windows

## Build

Requirements:

- Go 1.26+
- make

Build local binary:

```sh
make build
```

Output binary:

```text
./bin/discoverd
```

Build release binaries for major platforms:

```sh
make release
```

## Quick Check

Run server:

```sh
./bin/discoverd --mode server --secret mysecret
```

Run client (in another terminal):

```sh
./bin/discoverd --mode client --secret mysecret --output text
```

You can also set the secret with environment variable `DISCOVERD_SECRET`.
