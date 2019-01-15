# go-shadowsocks2

A fresh implementation of Shadowsocks in Go.

GoDoc at https://godoc.org/github.com/shadowsocks/go-shadowsocks2/

[![Build Status](https://travis-ci.org/shadowsocks/go-shadowsocks2.svg?branch=master)](https://travis-ci.org/shadowsocks/go-shadowsocks2)


## Features

- [x] SOCKS5 proxy with UDP Associate
- [x] Support for Netfilter TCP redirect (IPv6 should work but not tested)
- [x] UDP tunneling (e.g. relay DNS packets)
- [x] TCP tunneling (e.g. benchmark with iperf3)


## Install

Pre-built binaries for common platforms are available at https://github.com/shadowsocks/go-shadowsocks2/releases

Install from source

```sh
go get -u -v github.com/shadowsocks/go-shadowsocks2
```


## Basic Usage

### Server

* Way 1

Start a server listening on port 8488 using `AES-192-CFB` AES cipher with password `your-password`.

```sh
shadowsocks2 -s 'ss://AES-192-CFB:your-password@:8488' -d
```

* Way 2

use config file. Manager mode must use config file

```sh
shadowsocks2 -c config.json
```

config file example:

```json
{
    "debug":true,
    "server_port":1066,
    "password":"password",
    "method": "AES-192-CFB",
    "timeout":600,
    "manager_addr":"192.168.1.10:2066",
    "manager_pwd":"password",
    "manager_method":"AES-192-CFB"
}
```

* Way 3

```sh
shadowsocks2 -p 1066 -pwd "password" -m "AES-192-CFB" -t 600 -d
```

Or by default value, only set password

```sh
shadowsocks2 -pwd "password" -d
```

## Advanced Usage


### Netfilter TCP redirect (Linux only)

The client offers `-redir` and `-redir6` (for IPv6) options to handle TCP connections 
redirected by Netfilter on Linux. The feature works similar to `ss-redir` from `shadowsocks-libev`.


Start a client listening on port 1082 for redirected TCP connections and port 1083 for redirected
TCP IPv6 connections.

```sh
shadowsocks2 -c 'ss://AEAD_CHACHA20_POLY1305:your-password@[server_address]:8488' -redir :1082 -redir6 :1083
```


### TCP tunneling

The client offers `-tcptun [local_addr]:[local_port]=[remote_addr]:[remote_port]` option to tunnel TCP.
For example it can be used to proxy iperf3 for benchmarking.

Start iperf3 on the same machine with the server.

```sh
iperf3 -s
```

By default iperf3 listens on port 5201.

Start a client on the same machine with the server. The client listens on port 1090 for incoming connections
and tunnels to localhost:5201 where iperf3 is listening.

```sh
shadowsocks2 -c 'ss://AEAD_CHACHA20_POLY1305:your-password@[server_address]:8488' -tcptun :1090=localhost:5201
```

Start iperf3 client to connect to the tunneld port instead

```sh
iperf3 -c localhost -p 1090
```


## Design Principles

The code base strives to

- be idiomatic Go and well organized;
- use fewer external dependences as reasonably possible;
- only include proven modern ciphers;
