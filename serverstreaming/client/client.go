package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"

	pb "github.com/pigfrown/go-grpc-opentracing-streaming-example/serverstreaming/proto"

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
	tracer, closer, err := initTracer("localhost", "exampleclient")
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
	in := &pb.Request{Id: 1}
	stream, err := client.FetchResponse(context.Background(), in)
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}

	//ctx := stream.Context()
	done := make(chan bool)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Printf("Resp received: %s", resp.Result)
		}
	}()

	<-done
	log.Printf("finished")
}
