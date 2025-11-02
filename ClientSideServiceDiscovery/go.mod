module github.com/GergesHany/Event-Streaming-System/ClientSideServiceDiscovery

go 1.24.0

exclude google.golang.org/genproto v0.0.0-20200423170343-7949de9c1215
exclude google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
exclude google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55

require (
	github.com/GergesHany/Event-Streaming-System/SecurityAndObservability v0.0.0
	github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC v0.0.0
	github.com/stretchr/testify v1.11.1
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.75.1
)

replace (
	github.com/GergesHany/Event-Streaming-System/SecurityAndObservability => ../SecurityAndObservability
	github.com/GergesHany/Event-Streaming-System/ServeRequestsWithgRPC => ../ServeRequestsWithgRPC
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20251029180050-ab9386a59fda
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
