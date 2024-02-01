package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
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


func NewServer() *mux.Router {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/static/hello", mux.HandlerFunc(helloResource))
	r.Handle("/{res:[^\\/]+}", mux.HandlerFunc(dynamicResource))
	return r
}