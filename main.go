package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/things-go/go-socks5"
)

type CustomResolver struct {
	resolver *net.Resolver
}

func (d *CustomResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	ips, err := d.resolver.LookupIPAddr(ctx, name)
	if err != nil {
		return ctx, nil, err
	}
	if len(ips) == 0 {
		return ctx, nil, fmt.Errorf("no IPs found for host: %s", name)
	}

	for _, ipAddr := range ips {
		if ipAddr.IP.To4() != nil {
			return ctx, ipAddr.IP, nil
		}
	}

	return ctx, ips[0].IP, nil
}

func main() {
	bindAddr := flag.String("bind", ":1080", "socks5 listen address (host:port)")
	dnsAddr := flag.String("dns", "", "DNS server address (host:port)")
	ifaceName := flag.String("iface", "", "network interface name to bind outgoing connections (empty to disable)")
	flag.Parse()

	log.Printf("socks5 bind=%s dns=%s iface=%s", *bindAddr, *dnsAddr, *ifaceName)

	var dialer *net.Dialer
	if *ifaceName == "" {
		log.Println("Using default network interface for outgoing connections")
		dialer = &net.Dialer{}
	} else {
		log.Printf("Binding outgoing connections to interface: %s", *ifaceName)
		dialer = &net.Dialer{
			Control: func(network, address string, c syscall.RawConn) error {
				var bindErr error
				err := c.Control(func(fd uintptr) {
					if *ifaceName != "" {
						bindErr = syscall.BindToDevice(int(fd), *ifaceName)
						if bindErr != nil {
							log.Println("Warning: failed to bind to interface", *ifaceName, ":", bindErr)
						}
					}
				})
				if err != nil {
					return err
				}
				return bindErr
			},
		}
	}

	var customNetResolver *net.Resolver
	if *dnsAddr == "" {
		log.Println("Using system default DNS resolver")
		customNetResolver = net.DefaultResolver
	} else {
		log.Printf("Using custom DNS server: %s", *dnsAddr)
		customNetResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", *dnsAddr)
			},
		}
	}

	server := socks5.NewServer(
		socks5.WithDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}),
		socks5.WithResolver(&CustomResolver{customNetResolver}),
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "", log.LstdFlags))),
	)

	if err := server.ListenAndServe("tcp", *bindAddr); err != nil {
		panic(err)
	}
}
