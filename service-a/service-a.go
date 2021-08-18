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

const thisServiceName = "service-a"

func main() {
	tracer, closer := tracing.Init(thisServiceName)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	outboundHostPort, ok := os.LookupEnv("OUTBOUND_HOST_PORT")
	if !ok {
		outboundHostPort = "localhost:8082"
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http.HandleFunc(/ping) // r=%+v", r)
		span := tracing.StartSpanFromRequest(tracer, r)
		defer span.Finish()

		time.Sleep(40 * time.Millisecond)

		ctx := opentracing.ContextWithSpan(context.Background(), span)
		response, err := ping.Ping(ctx, outboundHostPort)
		if err != nil {
			log.Fatalf("Error occurred: %s", err)
		}

		log.Printf("http.HandleFunc(/ping) // .Write()")
		w.Write([]byte(fmt.Sprintf("%s -> %s", thisServiceName, response)))
	})
	log.Printf("Listening on localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
