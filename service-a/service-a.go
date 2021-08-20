package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"ping/lib/ping"
	"ping/lib/tracing"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
)

const thisServiceName = "service-a"
const useSelfRef = true

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

		var span opentracing.Span
		if useSelfRef {
			ctx := jaeger.NewSpanContext( // see https://github.com/jaegertracing/jaeger-client-go/issues/510
				jaeger.TraceID {
					//High: 0,
					Low:  0x100, // root
				},
				jaeger.SpanID(0x100), // this
				jaeger.SpanID(0), // parent
				false, // sampled
				nil, // baggage
			)
			span = tracer.StartSpan("ping-receive", jaeger.SelfRef(ctx)) // see https://github.com/jaegertracing/jaeger-client-go#selfref
		} else {
			span = tracing.StartSpanFromRequest(tracer, r)
		}
		time.Sleep(40 * time.Millisecond)
		span.Finish()

		ctx := opentracing.ContextWithSpan(context.Background(), span)
		var response1 string
		var response2 string
		var err error
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			response1, err = ping.Ping(ctx, outboundHostPort, useSelfRef, tracer)
			if err != nil {
				log.Fatalf("Error occurred: %s", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			response1, err = ping.Ping(ctx, outboundHostPort, useSelfRef, tracer)
			if err != nil {
				log.Fatalf("Error occurred: %s", err)
			}
		}()

		wg.Wait()

		log.Printf("http.HandleFunc(/ping) // .Write()")
		w.Write([]byte(fmt.Sprintf("%s -> %s", thisServiceName, response1 + response2)))
	})
	log.Printf("Listening on localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
