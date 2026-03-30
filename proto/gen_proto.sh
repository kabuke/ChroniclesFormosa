#!/bin/bash
set -ex

# Ensure GOPATH/bin is in PATH for protoc plugins
export PATH=$PATH:$(go env GOPATH)/bin

# Install plugins if missing
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.0

DIR=/Users/black/SourceCode/ChroniclesFormosa/proto
OUT_DIR=/Users/black/SourceCode/ChroniclesFormosa/resource

# macOS Homebrew protoc path or global path
if command -v protoc &> /dev/null; then
    PROTOC_PATH="protoc"
else
    echo "ERROR: protoc is not installed or not in PATH."
    echo "Please install protoc: brew install protobuf"
    exit 1
fi

$PROTOC_PATH $DIR/*.proto -I=$DIR --go_out=$OUT_DIR --go_opt=paths=source_relative

echo "Proto generation completed successfully!"
