package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/luscis/ruijie-telemetry-proto/model"
	pb "github.com/luscis/ruijie-telemetry-proto/proto/pb"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const RuijieGrpcServerPort = ":12345"
const TelegrafURL = "http://xxxx:8881"

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
}

func (s *server) JsonSend(ctx context.Context, in *pb.JsonRequest) (*pb.JsonReply, error) {
	var lines []string
	aTime := time.Now()
	// 可以自行丰富
	if in.JsonEvent == model.IFMDataKey {
		var value model.Response[model.IFMData]

		if err := json.Unmarshal([]byte(in.JsonString), &value); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, fmt.Errorf("failed to parse JSON value: %w", err)
		}
		// 遍历数据并创建数据点
		for _, datum := range value.Data {
			// 创建数据点
			tags := map[string]string{
				"ifx":       strconv.Itoa(datum.Ifx),
				"port_name": datum.PortName,
			}
			fields := map[string]interface{}{
				"outp_drop_pkts": datum.OutpDropPkts,
			}

			line := pointToLineProtocol("ifm_interface", tags, fields, aTime)
			lines = append(lines, line)
		}
	}

	// 拼接所有行
	body := strings.Join(lines, "\n")

	resp, err := http.Post(TelegrafURL+"/write", "text/plain", strings.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "HTTP Post error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("Telegraf returned %d: %s", resp.StatusCode, string(bodyBytes))
		fmt.Fprintln(os.Stderr, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	fmt.Printf("Sent %d points to Telegraf\n", len(lines))
	return &pb.JsonReply{Ret: 1}, nil
}

// escape 用于转义 Line Protocol 中的特殊字符
func escape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `,`, `\,`)
	s = strings.ReplaceAll(s, ` `, `\ `)
	s = strings.ReplaceAll(s, `=`, `\=`)
	return s
}

// pointToLineProtocol 将数据点转换为 InfluxDB Line Protocol 字符串
func pointToLineProtocol(measurement string, tags map[string]string, fields map[string]interface{}, t time.Time) string {
	var line strings.Builder

	// 1. measurement
	line.WriteString(escape(measurement))

	// 2. tags
	tagKeys := make([]string, 0, len(tags))
	for k := range tags {
		tagKeys = append(tagKeys, k)
	}
	// 推荐排序以保证一致性
	for _, k := range tagKeys {
		line.WriteString(",")
		line.WriteString(escape(k))
		line.WriteString("=")
		line.WriteString(escape(tags[k]))
	}

	line.WriteString(" ")

	// 3. fields
	fieldKeys := make([]string, 0, len(fields))
	for k := range fields {
		fieldKeys = append(fieldKeys, k)
	}
	for i, k := range fieldKeys {
		if i > 0 {
			line.WriteString(",")
		}
		line.WriteString(escape(k))
		line.WriteString("=")

		v := fields[k]
		switch val := v.(type) {
		case int, int8, int16, int32, int64:
			line.WriteString(fmt.Sprintf("%di", val))
		case uint, uint8, uint16, uint32, uint64:
			line.WriteString(fmt.Sprintf("%du", val))
		case float32, float64:
			line.WriteString(fmt.Sprintf("%g", val))
		case string:
			escaped := strings.ReplaceAll(val, `"`, `\"`)
			line.WriteString(`"` + escaped + `"`)
		case bool:
			line.WriteString(fmt.Sprintf("%v", val))
		default:
			line.WriteString(`"unknown"`)
		}
	}

	// 4. timestamp (nanoseconds)
	line.WriteString(" ")
	line.WriteString(fmt.Sprintf("%d", t.UnixNano()))

	return line.String()
}

func main() {
	lis, err := net.Listen("tcp", RuijieGrpcServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err != nil {
		log.Fatal("Error creating InfluxDB client:", err)
	}
	s := grpc.NewServer()

	pb.RegisterJsonServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
