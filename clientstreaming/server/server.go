package main

import (
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/pigfrown/go-grpc-opentracing-streaming-example/clientstreaming/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"

	"google.golang.org/grpc"
)

type server struct{}

func (s server) FetchResponse(stream pb.StreamService_FetchResponseServer) error {
	sumOfIdsSent := int32(0)
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Response{
				Result: fmt.Sprintf("%d", sumOfIdsSent),
			})
		}
		if err != nil {
			return err
		}
		log.Printf("recieved ID from client : %d", msg.Id)
		sumOfIdsSent += msg.Id
	}
}

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
	// Create tracer
	tracer, closer, err := initTracer("localhost", "clientstreaming-server")
	defer closer.Close()

	middleware := grpc_middleware.ChainStreamServer(
		grpc_opentracing.StreamServerInterceptor(
			grpc_opentracing.WithTracer(tracer),
		),
	)
	opts := []grpc.ServerOption{grpc.StreamInterceptor(middleware)}

	// create listener
	lis, err := net.Listen("tcp", ":50005")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	s := grpc.NewServer(opts...)
	pb.RegisterStreamServiceServer(s, server{})

	log.Println("start server")
	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
