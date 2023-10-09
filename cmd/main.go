package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
	pb "github.com/luscis/ruijie-telemetry-proto/proto/pb"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/grpc"
	"log"
	"net"
	"strings"
)

const RuijieGrpcServerPort = "172.30.17.163:57000"
const RuijieSwitch = "172.30.19.12"

// Parse path to path-buffer and tag-field
//
//nolint:revive //function-result-limit conditionally 4 return results allowed
func handlePath(gnmiPath *gnmiLib.Path, tags map[string]string, prefix string) (path string, err error) {
	builder := bytes.NewBufferString(prefix)

	// Parse generic keys from prefix
	for _, elem := range gnmiPath.Elem {
		if len(elem.Name) > 0 {
			if _, err := builder.WriteString(strings.ToUpper(elem.Name)); err != nil {
				return "", err
			}
		}
		if tags != nil {
			for key, val := range elem.Key {
				key = strings.ReplaceAll(key, "-", "_")

				// Use short-form of key if possible
				if _, exists := tags[key]; exists {
					tags[elem.Name+"_"+key] = val
				} else {
					tags[key] = val
				}
			}
		}
	}

	return builder.String(), nil
}

type server struct {
	pb.UnimplementedJsonServer
}

func (s *server) JsonSend(ctx context.Context, in *pb.JsonRequest) (*pb.JsonReply, error) {
	//fmt.Printf("\n\nReceived: %v %s\n", in.DeviceInfo, in.SensorPath)

	spath, err := ygot.StringToStructuredPath(in.SensorPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Path value: %w", err)
	}

	tags := make(map[string]string, 2)
	gpath, err := handlePath(spath, tags, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse Path value: %w", err)
	}

	fmt.Printf("\n\nReceived: %v %v\n", in.DeviceInfo, in.SensorPath)

	var value interface{}
	if err := json.Unmarshal([]byte(in.JsonString), &value); err != nil {
		return nil, fmt.Errorf("failed to parse JSON value: %w", err)
	}
	// fmt.Printf("%v\n", value)
	fields := make(map[string]interface{})
	flattener := jsonparser.JSONFlattener{Fields: fields}
	if err := flattener.FullFlattenJSON(strings.Replace(gpath, "-", "_", -1), value, true, true); err != nil {
		return nil, fmt.Errorf("failed to flatten JSON: %w", err)
	}

	label := "{"
	for k, v := range tags {
		label += fmt.Sprintf("%s=%s", k, v)
	}
	label += "}"
	for k, v := range fields {
		fmt.Printf("%s %s %v\n", k, label, v)
	}

	return &pb.JsonReply{Ret: 1}, nil
}

func main() {
	lis, err := net.Listen("tcp", RuijieGrpcServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterJsonServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
