package main

import (
	"context"
	"encoding/base64"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
)

//var logger = log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)

type headerTransport struct {
	transport http.RoundTripper
	headers   map[string]string
}

func (ht *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range ht.headers {
		req.Header.Set(key, value)
	}
	return ht.transport.RoundTrip(req)
}

func getClient() *http.Client {
	auth := "service-id" + ":" + "69d5bfba-ca88-4a6a-8200-78b351526b2b"
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	client := &http.Client{
		Transport: &headerTransport{
			transport: http.DefaultTransport,
			headers: map[string]string{
				"Authorization": authHeader,
			},
		},
	}

	return client
}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer(url string) (func(context.Context) error, error) {
	exporter, err := zipkin.New(
		url,
		//zipkin.WithLogger(logger),
		zipkin.WithClient(getClient()),
	)
	if err != nil {
		return nil, err
	}

	batcher := sdktrace.NewBatchSpanProcessor(exporter)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("zipkin-test"),
		)),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

func main() {
	url := flag.String("zipkin", "https://tracer.stage.kops.dev/api/v2/spans/e4ad4f9f-225b-4bfd-9f9a-39cc62f41598", "zipkin url")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := initTracer(*url)
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")
		tr := otel.GetTracerProvider().Tracer("component-main")
		ctx, span := tr.Start(r.Context(), "handle-request", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		traceID := trace.SpanContextFromContext(ctx).TraceID().String()
		w.Header().Set("X-Trace-ID", traceID)

		switch action {
		case "task1":
			task1(ctx, w)
		case "task2":
			task2(ctx, w)
		default:
			http.Error(w, "Invalid action", http.StatusBadRequest)
		}
	})

	log.Println("Starting server on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("server failed: %s", err)
	}

}

func task1(ctx context.Context, w http.ResponseWriter) {
	tr := otel.GetTracerProvider().Tracer("component-task1")

	// Start a new span for task1
	ctx, span := tr.Start(ctx, "task1")
	defer span.End()

	// Simulate a longer duration of work
	part1(ctx)
	part2(ctx)
	part3(ctx)

	w.Write([]byte("Result from task1"))
}

func part1(ctx context.Context) {
	tr := otel.GetTracerProvider().Tracer("component-part1")

	// Start a new span for part1
	_, span := tr.Start(ctx, "part1")
	defer span.End()

	// Simulate work for part1
	time.Sleep(10 * time.Second)
}

func part2(ctx context.Context) {
	tr := otel.GetTracerProvider().Tracer("component-part2")

	// Start a new span for part2
	_, span := tr.Start(ctx, "part2")
	defer span.End()

	// Simulate work for part2
	time.Sleep(10 * time.Second)
}

func part3(ctx context.Context) {
	tr := otel.GetTracerProvider().Tracer("component-part3")

	// Start a new span for part3
	_, span := tr.Start(ctx, "part3")
	defer span.End()

	// Simulate work for part3
	time.Sleep(10 * time.Second)
}

func task2(ctx context.Context, w http.ResponseWriter) {
	tr := otel.GetTracerProvider().Tracer("component-task2")
	_, span := tr.Start(ctx, "task2")
	defer span.End()

	// Simulate some different work
	time.Sleep(10 * time.Second)
	w.Write([]byte("Result from task2"))
}
