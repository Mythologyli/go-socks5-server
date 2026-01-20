# go-socks5-server

SOCK5 server that can bind traffic to a certain interface.

## Guide

`go-socks5-server -bind <address:port> -dns <address:port> -iface <interface-name>`

+ `-bind`: address and port to bind the SOCKS5 server (e.g. `:1080`)
+ `-dns`: address and port of the DNS server to use (e.g. `192.168.225.1:53`)
+ `-iface`: name of the interface to bind traffic to (e.g. `ec20`)