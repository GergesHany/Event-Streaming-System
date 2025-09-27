module github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC

go 1.24.0

exclude google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013

exclude cloud.google.com/go v0.26.0

require (
	github.com/GergesHany/Event-Streaming-System/SecurityAndObservability v0.0.0
	github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf v0.0.0
	github.com/GergesHany/Event-Streaming-System/WriteALogPackage v0.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/stretchr/testify v1.11.1
	go.opencensus.io v0.24.0
	go.uber.org/zap v1.27.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.1
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/casbin/casbin v1.9.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tysonmote/gommap v0.0.3 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf => ../StructureDataWithProtobuf

replace github.com/GergesHany/Event-Streaming-System/WriteALogPackage => ../WriteALogPackage

replace github.com/GergesHany/Event-Streaming-System/SecurityAndObservability => ../SecurityAndObservability
