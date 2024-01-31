package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
	"flag"
	"os"
	"path"
	"crypto/tls"
	"context"
	"crypto/x509"

	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v3/dtls"
	coap "github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/examples/dtls/pki"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/udp/client"
)


func check(e error) {
    if e != nil {
        panic(e)
    }
}

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		log.Printf("ClientAddress %v, %v\n", w.Conn().RemoteAddr(), r.String())
		next.ServeCOAP(w, r)
	})
}

func helloResource(w mux.ResponseWriter, r *mux.Message) {
	err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte(fmt.Sprintf("Hello from the cloud! The time is: %s.", time.Now().Format(time.RFC3339)))))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

// TODO: Replace with DB connection
// @see https://learn.microsoft.com/en-us/azure/azure-sql/database/connect-query-go?view=azuresql
var storage = make(map[string]([]byte))

func dynamicResource(w mux.ResponseWriter, r *mux.Message) {
	resp := w.Conn().AcquireMessage(r.Context())
	defer w.Conn().ReleaseMessage(resp)
	resp.SetToken(r.Token())
	resp.SetContentFormat(message.TextPlain)

	path, pErr := r.Path()
	if pErr != nil {
		resp.SetCode(codes.BadRequest)
		w.Conn().WriteMessage(resp)
		return
	}

	switch r.Code() {
	case codes.PUT:
		data, err := io.ReadAll(r.Body())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Stored content: '%s' in '%s'\n", data, path)
		storage[path] = data
		resp.SetBody(bytes.NewReader([]byte("OK")))
	case codes.GET:
		stored, ok := storage[path]
		if !ok {
			log.Printf("Not content found at '%s'\n", path)
			resp.SetCode(codes.NotFound)
		} else {
			log.Printf("Loaded content: '%s' from '%s'\n", stored, path)
			resp.SetCode(codes.Content)
			resp.SetBody(bytes.NewReader([]byte(stored)))
		}
	}

	err := w.Conn().WriteMessage(resp)
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func main() {
	address := flag.String("address", "localhost",
		"The UDP Server listen address, e.g. `localhost` or `0.0.0.0`.")
	certDir := flag.String("certDir", "/home/coap", "The folder containing the server certificate.")
	udpAddr := fmt.Sprintf("%s:%d", *address, 5688)
	dtlsAddr := fmt.Sprintf("%s:%d", *address, 5689)
	flag.Parse()
	fmt.Printf("UDP Server listening on: %s\n", udpAddr)
	fmt.Printf("dTLS UDP Server listening on: %s\n", dtlsAddr)

	// Read certificate files
	caCert, err := os.ReadFile(path.Join(*certDir, "CA.crt"))
    check(err)
	cert, err := os.ReadFile(path.Join(*certDir, "server.crt"))
    check(err)
	key, err := os.ReadFile(path.Join(*certDir, "server.key"))
    check(err)
	certificate, err := pki.LoadKeyAndCertificate(key, cert)
	check(err)
	certPool, err := pki.LoadCertPool(caCert)
	check(err)

	ctx := context.Background()

	tlsConfig := &piondtls.Config{
		Certificates:         []tls.Certificate{*certificate},
		ExtendedMasterSecret: piondtls.RequireExtendedMasterSecret,
		ClientCAs:            certPool,
		ClientAuth:           piondtls.RequireAndVerifyClientCert,
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(ctx, 30*time.Second)
		},
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/static/hello", mux.HandlerFunc(helloResource))
	r.Handle("/{res:[^\\/]+}", mux.HandlerFunc(dynamicResource))

	log.Println("Server starting")

	log.Fatal(coap.ListenAndServe("udp", udpAddr, r))
	log.Fatal(listenAndServeDTLS("udp", dtlsAddr, tlsConfig, r))
}


func listenAndServeDTLS(network string, addr string, config *piondtls.Config, handler mux.Handler) error {
	l, err := net.NewDTLSListener(network, addr, config)
	if err != nil {
		return err
	}
	defer l.Close()
	s := dtls.NewServer(options.WithMux(handler), options.WithOnNewConn(onNewConn))
	return s.Serve(l)
}

func onNewConn(cc *client.Conn) {
	dtlsConn, ok := cc.NetConn().(*piondtls.Conn)
	if !ok {
		log.Fatalf("invalid type %T", cc.NetConn())
	}
	clientCert, err := x509.ParseCertificate(dtlsConn.ConnectionState().PeerCertificates[0])
	if err != nil {
		log.Fatal(err)
	}
	cc.SetContextValue("client-cert", clientCert)
	cc.AddOnClose(func() {
		log.Println("closed connection")
	})
}