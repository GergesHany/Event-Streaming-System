# Configuration
CONFIG_PATH := ${HOME}/./Event-Streaming-System/SecureYourServices

# Phony targets
.PHONY: init gencert test clean

# Initialize configuration directory
init:
	mkdir -p ${CONFIG_PATH}

# Generate certificates
gencert:
	cfssl gencert \
		-initca test/ca-csr.json | cfssljson -bare ca
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server \
		test/server-csr.json | cfssljson -bare server
	mv *.pem *.csr ${CONFIG_PATH}

# Run tests
test:
	go test -race ./...

# Clean generated certificates
clean:
	rm -rf ${CONFIG_PATH}