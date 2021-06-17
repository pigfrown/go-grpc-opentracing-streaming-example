module github.com/pigfrown/go-grpc-opentracing-streaming-example/client

replace github.com/pigfrown/go-grpc-opentracing-streaming-example/serverstreaming/proto => ../proto

go 1.16

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pigfrown/go-grpc-opentracing-streaming-example/serverstreaming/proto v0.0.0-00010101000000-000000000000
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	google.golang.org/grpc v1.38.0
)
