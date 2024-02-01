package main

import (
	"flag"
	"fmt"
	"log"

	coap "github.com/plgd-dev/go-coap/v3"
	piondtls "github.com/pion/dtls/v2"
	"coap-server/internal"
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
	dtls := flag.Bool("dTLS", false, "Start a dTLS server")
	flag.Parse()

	r := internal.NewServer()

	if (*dtls) {
		dtlsAddr := fmt.Sprintf("%s:%d", *address, 5689)
		fmt.Printf("dTLS UDP Server listening on: %s\n", dtlsAddr)
		fmt.Printf("dTLS PSK: %s\n", *password)
		log.Fatal(coap.ListenAndServeDTLS("udp", dtlsAddr, &piondtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				fmt.Printf("Client's hint: %s \n", hint)
				return []byte(*password), nil
			},
			PSKIdentityHint: []byte("Pion DTLS Client"),
			CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
		}, r))
	} else {
		udpAddr := fmt.Sprintf("%s:%d", *address, 5688)
		fmt.Printf("UDP Server listening on: %s\n", udpAddr)
		log.Fatal(coap.ListenAndServe("udp", udpAddr, r))
	}
}
