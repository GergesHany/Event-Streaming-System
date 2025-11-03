module github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC

go 1.24.0

exclude google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
exclude google.golang.org/genproto v0.0.0-20200423170343-7949de9c1215
exclude google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55

exclude cloud.google.com/go v0.26.0

require (
	github.com/GergesHany/Event-Streaming-System/SecurityAndObservability v0.0.0
	github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf v0.0.0
	github.com/GergesHany/Event-Streaming-System/WriteALogPackage v0.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/stretchr/testify v1.11.1
	go.opencensus.io v0.24.0
	go.uber.org/zap v1.27.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.10
)

require (
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/casbin/casbin v1.9.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/go-hclog v1.6.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.4 // indirect
	github.com/hashicorp/go-msgpack/v2 v2.1.2 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/raft v1.7.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/tysonmote/gommap v0.0.3 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf => ../StructureDataWithProtobuf

replace github.com/GergesHany/Event-Streaming-System/WriteALogPackage => ../WriteALogPackage

replace github.com/GergesHany/Event-Streaming-System/SecurityAndObservability => ../SecurityAndObservability

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20251029180050-ab9386a59fda
