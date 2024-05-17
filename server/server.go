package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"coap-server/internal"

	piondtls "github.com/pion/dtls/v2"
	coap "github.com/plgd-dev/go-coap/v3"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type serverConfig struct {
	oldUdpPort  int
	newUdpPort  int
	oldDtlsPort int
	newDtlsPort int
}

var config = serverConfig{
	oldUdpPort:  5688,
	newUdpPort:  5683,
	oldDtlsPort: 5689,
	newDtlsPort: 5684,
}

func main() {
	address := flag.String("address", "localhost",
		"The UDP Server listen address, e.g. `localhost` or `0.0.0.0`.")
	password := flag.String("password", "connect:anything",
		"The password to use for the PSK in dTLS.")
	dtls := flag.Bool("dTLS", false, "Start a dTLS server")
	network := flag.String("network", "udp4",
		"The network to use, `udp4` or `udp6`.")
	flag.Parse()

	storageConnectionString, ok := os.LookupEnv("STORAGE_CONNECTION_STRING")
	if !ok {
		log.Fatal("the environment variable 'STORAGE_CONNECTION_STRING' could not be found")
	}
	storageClient, err := azblob.NewClientFromConnectionString(storageConnectionString, nil)
	check(err)
	log.Printf("Azure Storage client URL: %s\n", storageClient.URL())

	containerName, ok := os.LookupEnv("STORAGE_CONTAINER_NAME")
	if !ok {
		log.Fatal("the environment variable 'STORAGE_CONTAINER_NAME' could not be found")
	}
	log.Printf("Container name: %s\n", containerName)

	r := internal.NewServer(storageClient, containerName)

	go func() {
		udpAddr := fmt.Sprintf("%s:%d", *address, config.oldUdpPort)
		dtlsAddr := fmt.Sprintf("%s:%d", *address, config.oldDtlsPort)

		if *dtls {
			log.Printf("dTLS %s Server listening on: %s\n", *network, dtlsAddr)
			log.Printf("dTLS PSK: %s\n", *password)
			log.Fatal(coap.ListenAndServeDTLS(*network, dtlsAddr, &piondtls.Config{
				PSK: func(hint []byte) ([]byte, error) {
					log.Printf("Client's hint: %s \n", hint)
					return []byte(*password), nil
				},
				PSKIdentityHint: []byte("Pion DTLS Client"),
				CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
			}, r))
		} else {
			log.Printf("%s Old Server listening on: %s\n", *network, udpAddr)
			log.Fatal(coap.ListenAndServe(*network, udpAddr, r))
		}
	}()

	udpAddr := fmt.Sprintf("%s:%d", *address, config.newUdpPort)
	dtlsAddr := fmt.Sprintf("%s:%d", *address, config.newDtlsPort)

	if *dtls {
		log.Printf("dTLS %s Server listening on: %s\n", *network, dtlsAddr)
		log.Printf("dTLS PSK: %s\n", *password)
		log.Fatal(coap.ListenAndServeDTLS(*network, dtlsAddr, &piondtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				log.Printf("Client's hint: %s \n", hint)
				return []byte(*password), nil
			},
			PSKIdentityHint: []byte("Pion DTLS Client"),
			CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
		}, r))
	} else {
		log.Printf("%s Server listening on: %s\n", *network, udpAddr)
		log.Fatal(coap.ListenAndServe(*network, udpAddr, r))
	}
}
