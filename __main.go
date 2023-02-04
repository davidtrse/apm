package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {

	// common init
	// You may also want to set them as globals
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
	bsp := sdktrace.NewSimpleSpanProcessor(exp) // You should use batch span processor in prod
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)

	propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	ctx, span := tp.Tracer("foo").Start(context.Background(), "parent-span-name")
	defer span.End()

	// Serialize the context into carrier
	carrier := propagation.MapCarrier{}
	propgator.Inject(ctx, carrier)
	// This carrier is sent accros the process
	fmt.Println(carrier)

	ctxx, _ := tp.Tracer("foo").Start(ctx, "parent1-span-name")

	// Serialize the context into carrier
	propgator.Inject(ctxx, carrier)
	// This carrier is sent accros the process
	fmt.Println(carrier)

	// Extract the context and start new span as child
	// In your receiving function

	parentCtx := propgator.Extract(context.Background(), carrier)

	propgator2 := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	parentCtx2 := propgator2.Extract(context.Background(), carrier)
	// Serialize the context into carrier
	propgator.Inject(parentCtx, carrier)
	// This carrier is sent accros the process
	fmt.Println(carrier)
	// Serialize the context into carrier
	propgator.Inject(parentCtx2, carrier)
	// This carrier is sent accros the process
	fmt.Println(carrier)

	if parentCtx != parentCtx2 {
		fmt.Println("different")
	}

	_, childSpan := tp.Tracer("foo").Start(parentCtx, "child-span-name")
	childSpan.AddEvent("some-dummy-event")
	childSpan.End()
}
