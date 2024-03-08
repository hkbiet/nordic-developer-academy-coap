package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
	"github.com/plgd-dev/go-coap/v3/udp/client"

	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/dtls"

	"github.com/google/uuid"

	"coap-client/internal"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	address := flag.String("address", "localhost",
		"The UDP Server listen address, e.g. `localhost`.")
	password := flag.String("password", "connect:anything",
		"The password to use for the PSK in dTLS.")
	udp6 := flag.Bool("udp6", false, "Whether to use IPv6")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	udpPort := 5688
	dTLSPort := 5689
	var udpAddr string
	var dtlsAddr string
	config := client.Config{}
	var opt udp.Option
	if *udp6 {
		ip6, err := net.DefaultResolver.LookupIP(context.Background(), "ip6", *address)
		if err != nil {
			log.Fatal("Failed to resolve IPv6 address: ", err)
		}
		udpAddr = fmt.Sprintf("[%s]:%d", ip6[0].String(), udpPort)
		dtlsAddr = fmt.Sprintf("[%s]:%d", ip6[0].String(), dTLSPort)
		opt = options.WithNetwork("udp6")
	} else {
		ip4, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", *address)
		if err != nil {
			log.Fatal("Failed to resolve IPv4 address: ", err)
		}
		udpAddr = fmt.Sprintf("%s:%d", ip4[0].String(), udpPort)
		dtlsAddr = fmt.Sprintf("%s:%d", ip4[0].String(), dTLSPort)
		opt = options.WithNetwork("udp")
	}

	log.Printf("UDP Server listening on: %s\n", udpAddr)

	opt.UDPClientApply(&config)
	co, err := udp.Dial(udpAddr, opt)

	check(err)

	internal.TestHello(co, ctx)
	id1 := []byte(uuid.New().String())
	internal.TestPutCustom(co, ctx, fmt.Sprintf("/%s", id1), id1)
	readId1, err := internal.TestGetCustom(co, ctx, fmt.Sprintf("/%s", id1))
	check(err)
	if string(readId1) != string(id1) {
		log.Fatalf("Read value %s is not equal written value %s.", id1, readId1)
	}

	log.Printf("dTLS Server listening on: %s\n", dtlsAddr)
	log.Printf("dTLS PSK: %s\n", *password)
	codTLS, err := dtls.Dial(dtlsAddr, &piondtls.Config{
		PSK: func(hint []byte) ([]byte, error) {
			log.Printf("Server's hint: %s \n", hint)
			return []byte(*password), nil
		},
		PSKIdentityHint: []byte("Pion DTLS Client"),
		CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
	}, opt)
	check(err)

	internal.TestHello(codTLS, ctx)
	id2 := []byte(uuid.New().String())
	internal.TestPutCustom(codTLS, ctx, fmt.Sprintf("/%s", id2), id2)
	readId2, err := internal.TestGetCustom(codTLS, ctx, fmt.Sprintf("/%s", id2))
	check(err)
	if string(readId2) != string(id2) {
		log.Fatalf("Read value %s is not equal written value %s.", id2, readId2)
	}
}
