package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/plgd-dev/go-coap/v3/udp"

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
		"The UDP Server listen address, e.g. `localhost` or `0.0.0.0`.")
	password := flag.String("password", "connect:anything",
		"The password to use for the PSK in dTLS.")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	id := uuid.New()

	udpAddr := fmt.Sprintf("%s:%d", *address, 5688)
	fmt.Printf("UDP Server listening on: %s\n", udpAddr)
	co, err := udp.Dial(udpAddr)
	check(err)

	internal.TestHello(co, ctx)
	internal.TestPutCustom(co, ctx, fmt.Sprintf("/%s", id))
	internal.TestGetCustom(co, ctx, fmt.Sprintf("/%s", id))

	dtlsAddr := fmt.Sprintf("%s:%d", *address, 5689)
	fmt.Printf("dTLS Server listening on: %s\n", dtlsAddr)
	fmt.Printf("dTLS PSK: %s\n", *password)
	codTLS, err := dtls.Dial(dtlsAddr, &piondtls.Config{
		PSK: func(hint []byte) ([]byte, error) {
			fmt.Printf("Server's hint: %s \n", hint)
			return []byte(*password), nil
		},
		PSKIdentityHint: []byte("Pion DTLS Client"),
		CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
	})
	check(err)

	internal.TestHello(codTLS, ctx)
	internal.TestPutCustom(codTLS, ctx, fmt.Sprintf("/%s", id))
	internal.TestGetCustom(codTLS, ctx, fmt.Sprintf("/%s", id))
}
