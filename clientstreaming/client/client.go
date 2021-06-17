package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"

	pb "github.com/pigfrown/go-grpc-opentracing-streaming-example/clientstreaming/proto"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"

	"time"

	"google.golang.org/grpc"
)

func initTracer(jaegerHostname, service string) (opentracing.Tracer, io.Closer, error) {
	endpoint := fmt.Sprintf("http://%s:14268/api/traces", jaegerHostname)
	cfg := &config.Configuration{
		ServiceName: service,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: endpoint,
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		return nil, nil, fmt.Errorf("Could not init Jaeger : %v", err)
	}
	return tracer, closer, nil
}

func main() {
	tracer, closer, err := initTracer("localhost", "clientstreaming-client")
	defer closer.Close()

	rand.Seed(time.Now().Unix())

	tracingInterceptor := grpc_middleware.ChainStreamClient(
		grpc_opentracing.StreamClientInterceptor(
			grpc_opentracing.WithTracer(tracer),
		),
	)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(tracingInterceptor),
	}

	// dail server
	conn, err := grpc.Dial("localhost:50005", opts...)
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewStreamServiceClient(conn)
	stream, err := client.FetchResponse(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}
	defer stream.CloseSend()

	msgCount := 5
	for i := 0; i < msgCount; i++ {
		if err := stream.Send(&pb.Request{Id: int32(i)}); err != nil {
			log.Fatalf("Cound not send : %v", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Could not close : %v", err)
	}
	log.Printf("finished with reply : %v", reply)
}
