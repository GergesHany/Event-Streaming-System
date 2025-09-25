module github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC

go 1.24.0

exclude google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013

require (
	github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf v0.0.0
	github.com/GergesHany/Event-Streaming-System/WriteALogPackage v0.0.0
	github.com/stretchr/testify v1.11.1
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tysonmote/gommap v0.0.3 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf => ../StructureDataWithProtobuf

replace github.com/GergesHany/Event-Streaming-System/WriteALogPackage => ../WriteALogPackage
