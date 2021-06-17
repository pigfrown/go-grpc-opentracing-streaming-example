package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/pramonow/go-grpc-server-streaming-example/src/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"

	"google.golang.org/grpc"
)

type server struct{}

func (s server) FetchResponse(in *pb.Request, srv pb.StreamService_FetchResponseServer) error {

	log.Printf("fetch response for id : %d", in.Id)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(count int64) {
			defer wg.Done()
			time.Sleep(time.Duration(count) * time.Second)
			resp := pb.Response{Result: fmt.Sprintf("Request #%d For Id:%d", count, in.Id)}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("finishing request number : %d", count)
		}(int64(i))
	}

	wg.Wait()
	return nil
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
	tracer, closer, err := initTracer("localhost", "exampleserver")
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
