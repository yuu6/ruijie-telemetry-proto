package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	client "github.com/influxdata/influxdb1-client"
	"github.com/luscis/ruijie-telemetry-proto/model"
	pb "github.com/luscis/ruijie-telemetry-proto/proto/pb"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

const RuijieGrpcServerPort = ":12345"
const TelegrafURL = "http://10.56.113.92:8881"

// Parse path to path-buffer and tag-field
//
//nolint:revive //function-result-limit conditionally 4 return results allowed
func handlePath(gnmiPath *gnmiLib.Path, tags map[string]string, prefix string) (path string, err error) {
	builder := bytes.NewBufferString(prefix)

	// Parse generic keys from prefix
	for _, elem := range gnmiPath.Elem {
		if len(elem.Name) > 0 {
			if _, err := builder.WriteString(strings.ToUpper(elem.Name)); err != nil {
				fmt.Fprintln(os.Stderr, err)
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
	client *client.Client
}

func (s *server) JsonSend(ctx context.Context, in *pb.JsonRequest) (*pb.JsonReply, error) {
	//fmt.Printf("\n\nReceived Data : %v \n", in)

	//// 创建写入的批次（BatchPoints）
	//bp, err := client.NewBatchPoints(client.BatchPointsConfig{
	//	Database:        "switch-telemetry", // 指定数据库（db）
	//	RetentionPolicy: "autogen",          // 可选：保留策略，通常为 autogen
	//	Precision:       "ms",               // 时间精度：ns（纳秒）、us、ms、s
	//})
	//if err != nil {
	//	log.Fatal("Error creating batch points:", err)
	//}
	batchPoints := client.BatchPoints{
		Database: "switch-telemetry",
	}
	now := time.Now().Add(-1 * time.Minute)

	if in.JsonEvent == model.IFMDataKey {
		var value model.Response[model.IFMData]

		if err := json.Unmarshal([]byte(in.JsonString), &value); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, fmt.Errorf("failed to parse JSON value: %w", err)
		}
		// 遍历数据并创建数据点
		for _, datum := range value.Data {
			// 创建数据点
			pt := client.Point{
				Measurement: "ifm_interface", // measurement
				Tags: map[string]string{ // tags
					"ifx":       datum.PortName,
					"port_name": datum.PortName,
				},
				Time: now,
				Fields: map[string]interface{}{ // fields
					"outp_drop_pkts": datum.OutpDropPkts,
				},
			}
			batchPoints.Points = append(batchPoints.Points, pt)
		}
	}

	fmt.Printf("Batch points: %v\n", batchPoints)

	write, err := s.client.Write(batchPoints)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	log.Printf("wrote %d", write)

	return &pb.JsonReply{Ret: 1}, nil
}

func main() {
	lis, err := net.Listen("tcp", RuijieGrpcServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	parse, err := url.Parse(TelegrafURL)
	if err != nil {
		return
	}

	c, err := client.NewClient(client.Config{
		URL: *parse, // InfluxDB v1 地址
		// Username: "your-username",   // 如果启用了认证
		// Password: "your-password",
	})
	if err != nil {
		log.Fatal("Error creating InfluxDB client:", err)
	}
	s := grpc.NewServer()

	pb.RegisterJsonServer(s, &server{
		client: c,
	})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
