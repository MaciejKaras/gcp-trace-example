package shared

import (
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"encoding/hex"
	"fmt"
	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
	"log"
	"net/http"
)

type TraceExporter struct {
	exporter *stackdriver.Exporter
}

func InitTrace() (*TraceExporter, error) {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		return nil, fmt.Errorf("stackdriver.NewExporter: %v", err)
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	trace.RegisterExporter(exporter)

	return &TraceExporter{exporter: exporter}, nil
}

func (t *TraceExporter) Flush() {
	if t.exporter != nil {
		t.exporter.Flush()
	}
}

type rootTraceIDKey struct{}

func StartRequestSpan(r *http.Request) (context.Context, *trace.Span) {
	if parent, ok := (&propagation.HTTPFormat{}).SpanContextFromRequest(r); ok {
		ctx, span := trace.StartSpanWithRemoteParent(r.Context(), r.RequestURI, parent, trace.WithSpanKind(trace.SpanKindServer))
		addRootTraceID(span, parent.TraceID.String())

		return ctx, span
	}

	ctx, span := trace.StartSpan(r.Context(), r.RequestURI, trace.WithSpanKind(trace.SpanKindServer))
	addRootTraceID(span, span.SpanContext().TraceID.String())

	return ctx, span
}

func StartCloudEventSpan(ctx context.Context, spanName string, messageAttributes map[string]string) (context.Context, *trace.Span) {
	ctx, span := trace.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
	addLinkAndLabels(span, messageAttributes)

	return ctx, span
}

func StartSpan(ctx context.Context, spanName string) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, spanName)
}

func StartClientSpan(ctx context.Context, spanName string) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
}

func addRootTraceID(span *trace.Span, rootTraceID string) {
	span.AddAttributes(trace.StringAttribute("rootTraceId", rootTraceID))
}

const (
	traceIDAttrKey = "traceIDAttrKey"
	spanIDAttrKey  = "spanIDAttrKey"
)

func ToMessageAttributes(span *trace.Span) map[string]string {
	return map[string]string{
		traceIDAttrKey: span.SpanContext().TraceID.String(),
		spanIDAttrKey:  span.SpanContext().SpanID.String(),
	}
}

func addLinkAndLabels(span *trace.Span, messageAttributes map[string]string) {
	remoteTraceID, err := getTraceID(messageAttributes[traceIDAttrKey])
	if err != nil {
		log.Printf("Failed reading %s from attributes: %+v", traceIDAttrKey, err)
		return
	}
	addRootTraceID(span, remoteTraceID.String())

	remoteSpanID, err := getSpanID(messageAttributes[spanIDAttrKey])
	if err != nil {
		log.Printf("Failed reading %s from attributes: %+v", spanIDAttrKey, err)
		return
	}

	span.AddLink(trace.Link{
		TraceID: remoteTraceID,
		SpanID:  remoteSpanID,
		Type:    trace.LinkTypeParent,
	})
}

func getTraceID(traceIDString string) (trace.TraceID, error) {
	var traceID trace.TraceID
	decodeString, err := hex.DecodeString(traceIDString)
	if err != nil {
		return trace.TraceID{}, fmt.Errorf("failed to decode traceID: %v", err)
	}

	copy(traceID[:], decodeString)

	return traceID, nil
}

func getSpanID(spanIDString string) (trace.SpanID, error) {
	var spanID trace.SpanID
	decodeString, err := hex.DecodeString(spanIDString)
	if err != nil {
		return trace.SpanID{}, fmt.Errorf("failed to decode spanID: %v", err)
	}

	copy(spanID[:], decodeString)

	return spanID, nil
}
