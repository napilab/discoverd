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

## Linux: systemd Service

This repository includes a production-oriented unit file at `./deploy/systemd/discoverd.service`.

Build and install binary:

```sh
make build
sudo install -m 0755 ./bin/discoverd /usr/local/bin/discoverd
```

Create a dedicated system user:

```sh
sudo useradd --system --no-create-home --shell /usr/sbin/nologin discoverd
```

Install service and environment file:

```sh
sudo install -m 0644 ./deploy/systemd/discoverd.service /etc/systemd/system/discoverd.service
sudo install -D -m 0600 ./deploy/systemd/discoverd.env.example /etc/default/discoverd
```

Edit `/etc/default/discoverd` and set a strong `DISCOVERD_SECRET` value.

Enable and start service:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now discoverd
sudo systemctl status discoverd --no-pager
```
