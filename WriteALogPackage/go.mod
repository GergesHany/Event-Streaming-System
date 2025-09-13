module github.com/GergesHany/Event-Streaming-System/WriteALogPackage

go 1.24.0

require (
	github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf v0.0.0
	github.com/stretchr/testify v1.11.1
	github.com/tysonmote/gommap v0.0.3
	google.golang.org/protobuf v1.25.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/GergesHany/Event-Streaming-System/StructureDataWithProtobuf => ../StructureDataWithProtobuf
