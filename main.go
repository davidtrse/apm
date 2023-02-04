package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	otlptracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	otelTrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/davidtrse/apm/log"
)

const ServiceName = "apm-demo"

var tracer otelTrace.Tracer

func main() {
	agentUrl := os.Getenv("OTLP_ENDPOINT")
	if agentUrl == "" {
		log.Fatalf("please set value of OTLP_ENDPOINT")
	}

	// tracing
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, agentUrl,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("failed to create gRPC connection to collector: %v", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(ServiceName),
		)),
	)
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)
	tracer = otel.Tracer(ServiceName)

	e := echo.New()
	e.GET("/hello/:name", func(c echo.Context) error {
		ctx, span := tracer.Start(c.Request().Context(), "GET /hello/:name")
		defer span.End()

		log.Infof("GET /hello")
		log.Infoff(ctx, "GET /hello/:name spanId=%s", span.SpanContext().SpanID())
		log.Infoff(ctx, "GET /hello/:name traceID=%s", span.SpanContext().TraceID())

		name := c.Param("name")
		msg := helloMsg(ctx, name)
		_ = howareyou(ctx, name)

		return c.JSON(http.StatusOK, msg)
	}, echo.WrapMiddleware(ShortVideoMiddleware))
	e.Logger.Fatal(e.Start(":8080"))
}

func helloMsg(ctx context.Context, name string) string {
	_, span := tracer.Start(ctx, "helloMsg")
	defer span.End()

	return fmt.Sprintf("hello %s", name)
}

func howareyou(ctx context.Context, name string) string {
	_, span := tracer.Start(ctx, "howareyou")
	defer span.End()

	return fmt.Sprintf("howareyou %s", name)
}

type ContextKey string

const ContextUserKey ContextKey = "user"

func ShortVideoMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ctx := context.WithValue(r.Context(), ContextUserKey, "val1ShortVideoMiddleware")

		ctx, span := tracer.Start(r.Context(), "ShortVideoMiddleware")
		defer span.End()

		log.Errorff(r.Context(), "GET /hello/:name traceID=%s", span.SpanContext().TraceID())

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
