package model

type Response[T any] struct {
	Timestamp int64 `json:"timestamp"`
	Data      []T   `json:"data"`
}
