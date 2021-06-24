// codegen tool generates custom (Un)MarshalJSON implementations for protobuf Messages.
// Messages are extracted from proto files in lib/assets/*.proto.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// This codegen tool will output a file containing custom implementations for MarshalJSON & UnmarshalJSON for every ProtoMessage.
// Each method is simply forwarding the (de)serialization to protojson, since it properly handles ProtoMessages.
// Unpopulated fields are emitted to prevent validation issue from the chaincode (which is based on reflection).
func main() {
	var path = flag.String("path", "", "help message for flag n")
	flag.Parse()

	assets := getAssets(*path)

	t := template.Must(template.New("protoMarshal").Parse(protojsonImplementationTemplate))

	cmd := exec.Command("gofmt")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		if err := t.Execute(stdin, assets); err != nil {
			log.Fatalf("execute err: %v", err)
		}
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)
}

// Extract asset names from lib/asset code
func getAssets(path string) []string {
	protoFiles, err := filepath.Glob(path + "/*.proto")
	if err != nil {
		log.Fatalf("failed to find proto files: %s", err)
	}

	protoMessages := make([]string, 0)

	for _, p := range protoFiles {
		f, err := os.Open(p)
		if err != nil {
			log.Fatalf("failed to read %s: %s", p, err)
		}
		defer f.Close()
		proto, err := protoparser.Parse(f)
		if err != nil {
			log.Fatalf("failed to parse %s: %s", p, err)
		}
		for _, item := range proto.ProtoBody {
			if msg, ok := item.(*parser.Message); ok {
				protoMessages = append(protoMessages, msg.MessageName)
			}
		}
	}

	return protoMessages
}

var protojsonImplementationTemplate = `
// THIS FILE IS GENERATED BY codegen/assets DO NOT EDIT !!!

package asset

import "google.golang.org/protobuf/encoding/protojson"

var marshaller protojson.MarshalOptions

func init() {
	marshaller = protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}
}

{{range .}}
// MarshalJSON forward the marshalling to protojson
func (n *{{.}}) MarshalJSON() ([]byte, error) {
	return marshaller.Marshal(n)
}

// UnmarshalJSON forward the unmarshalling to protojson
func (n *{{.}}) UnmarshalJSON(data []byte) error {
	return protojson.Unmarshal(data, n)
}
{{end}}
`
