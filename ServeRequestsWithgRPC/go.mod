module github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC

go 1.24.0

// This file describes the dependencies of this module
replace github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC => ../ServeRequestsWithgRPC

exclude google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013

require (
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.1
)

require (
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
)
