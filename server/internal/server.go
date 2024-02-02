package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
	"strings"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

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

func dynamicResource(client *azblob.Client, containerName string) func(mux.ResponseWriter, *mux.Message) {
	return func(w mux.ResponseWriter, r *mux.Message) {
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

		key := strings.Split(path[1:], "/")[0]

		switch r.Code() {
		case codes.PUT:
			payloadSize, err := r.BodySize()
			if err != nil {
				log.Fatal(err)
			}
			if (payloadSize > 100000) { // Max size 100 KB
				err := w.SetResponse(codes.RequestEntityTooLarge, message.TextPlain, bytes.NewReader([]byte("Maximum payload size is 100 KB!")))
				if err != nil {
					log.Printf("cannot set response: %v", err)
				}
				return 
			}
			data, err := io.ReadAll(r.Body())
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Stored content: '%s' in '%s'\n", data, path)
			Store(client, containerName, key, data)
			resp.SetBody(bytes.NewReader([]byte("OK")))
		case codes.GET:
			stored, err := Retrieve(client, containerName, key)
			if err != nil {
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
}


func NewServer(client *azblob.Client, containerName string) *mux.Router {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/static/hello", mux.HandlerFunc(helloResource))
	r.Handle("/{res:[^\\/]+}", mux.HandlerFunc(dynamicResource(client, containerName)))
	return r
}