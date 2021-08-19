package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"ping/lib/tracing"

	"github.com/opentracing/opentracing-go"
)

const thisServiceName = "service-c"

func main() {
	tracer, closer := tracing.Init(thisServiceName)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http.HandleFunc(/ping) // r=%+v", r)

		span := tracing.StartSpanFromRequest(tracer, r)
		time.Sleep(10 * time.Millisecond)
		span.Finish()

		log.Printf("http.HandleFunc(/ping) // .Write()")
		w.Write([]byte(fmt.Sprintf("%s", thisServiceName)))
	})
	log.Printf("Listening on localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}
