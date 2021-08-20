package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ping/lib/ping"
	"ping/lib/tracing"

	"github.com/opentracing/opentracing-go"
)

const thisServiceName = "service-b"
const useSelfRef = false

func main() {
	tracer, closer := tracing.Init(thisServiceName)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	outboundHostPort, ok := os.LookupEnv("OUTBOUND_HOST_PORT")
	if !ok {
		outboundHostPort = "localhost:8083"
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http.HandleFunc(/ping) // r=%+v", r)

		span := tracing.StartSpanFromRequest(tracer, r)
		time.Sleep(50 * time.Millisecond)
		span.Finish()

		ctx := opentracing.ContextWithSpan(context.Background(), span)
		response, err := ping.Ping(ctx, outboundHostPort, useSelfRef, tracer)
		if err != nil {
			log.Fatalf("Error occurred: %s", err)
		}

		log.Printf("http.HandleFunc(/ping) // .Write()")
		w.Write([]byte(fmt.Sprintf("%s -> %s", thisServiceName, response)))
	})
	log.Printf("Listening on localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
