package main

import (
	"flag"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/mux"
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

type serverPorts struct {
	oldUdpPort  int
	newUdpPort  int
	oldDtlsPort int
	newDtlsPort int
}

type flags struct {
	address  string
	password string
	dtls     bool
	network  string
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

	var flagValues = flags{
		address:  *address,
		password: *password,
		dtls:     *dtls,
		network:  *network,
	}

	var ports = serverPorts{
		oldUdpPort:  5688,
		newUdpPort:  5683,
		oldDtlsPort: 5689,
		newDtlsPort: 5684,
	}

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

	// old ports
	go launchServer(flagValues, ports.oldUdpPort, ports.oldDtlsPort, r)

	// new ports
	launchServer(flagValues, ports.newUdpPort, ports.newDtlsPort, r)

}

func launchServer(flagValues flags, udpPort int, dtlsPort int, r *mux.Router) {
	udpAddr := fmt.Sprintf("%s:%d", flagValues.address, udpPort)
	dtlsAddr := fmt.Sprintf("%s:%d", flagValues.address, dtlsPort)

	if flagValues.dtls {
		log.Printf("dTLS %s Server listening on: %s\n", flagValues.network, dtlsAddr)
		log.Printf("dTLS PSK: %s\n", flagValues.password)
		log.Fatal(coap.ListenAndServeDTLS(flagValues.network, dtlsAddr, &piondtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				log.Printf("Client's hint: %s \n", hint)
				return []byte(flagValues.password), nil
			},
			PSKIdentityHint: []byte("Pion DTLS Client"),
			CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
		}, r))
	} else {
		log.Printf("%s Server listening on: %s\n", flagValues.network, udpAddr)
		log.Fatal(coap.ListenAndServe(flagValues.network, udpAddr, r))
	}
}
