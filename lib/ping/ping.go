package ping

import (
	"context"
	"fmt"
	"net/http"
	"time"

	libhttp "ping/lib/http"
	"ping/lib/tracing"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
)

var counter int = 0x100

// Ping sends a ping request to the given hostPort, ensuring a new span is created
// for the downstream call, and associating the span to the parent span, if available
// in the provided context.
func Ping(ctx context.Context, hostPort string, useSelfRef bool, tracer opentracing.Tracer) (string, error) {
	url := fmt.Sprintf("http://%s/ping", hostPort)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	var span opentracing.Span
	if useSelfRef {
		counter ++
		ctx := jaeger.NewSpanContext( // see https://github.com/jaegertracing/jaeger-client-go/issues/510
			jaeger.TraceID {
				//High: 0,
				Low:  0x100, // root
			},
			jaeger.SpanID(counter), // this
			jaeger.SpanID(0x100), // parent
			false, // sampled
			nil, // baggage
		)
		span = tracer.StartSpan("ping-send", jaeger.SelfRef(ctx)) // see https://github.com/jaegertracing/jaeger-client-go#selfref
	} else {
		span, _ = opentracing.StartSpanFromContext(ctx, "ping-send")
	}
	time.Sleep(10 * time.Millisecond)
	span.Finish()

	if err := tracing.Inject(span, req); err != nil {
		return "", err
	}

	body, err := libhttp.Do(req)

	return body, err
}
