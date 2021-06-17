# Golang grpc server/client streaming example with tracing via grpc-middleware

Simple grpc streaming examples for both client and server streaming, with automatic tracing via grpc-middleware library.

# Usage

Spin up a jaeger docker to collect traces:

```
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.23
  ```

Choose an example (clientstreaming or serverstreaming) and in one terminal move to the server directory and start the server with ```go run server.go```. In another terminal move to the client directory and start the client with ```go run client.go```

You should see output in the termainl and traces in your jaeger UI (http://localhost:16686)

