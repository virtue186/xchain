package core

import (
	"encoding/json"
	"io"
)

// Encoder 定义了统一的、无状态的编码器接口
type Encoder[T any] interface {
	Encode(w io.Writer, v T) error
}

// Decoder 定义了统一的、无状态的解码器接口
type Decoder[T any] interface {
	Decode(r io.Reader, v T) error
}

// ======================= 新增 JSON 实现 =======================
// JSONEncoder 是一个具体的 json 编码器实现
type JSONEncoder[T any] struct{}

func (e JSONEncoder[T]) Encode(w io.Writer, v T) error {
	return json.NewEncoder(w).Encode(v)
}

// JSONDecoder 是一个具体的 json 解码器实现
type JSONDecoder[T any] struct{}

func (d JSONDecoder[T]) Decode(r io.Reader, v T) error {
	return json.NewDecoder(r).Decode(v)
}

//// GOBEncoder 是一个具体的 gob 编码器实现
//type GOBEncoder[T any] struct{}
//
//func (e GOBEncoder[T]) Encode(w io.Writer, v T) error {
//	return gob.NewEncoder(w).Encode(v)
//}
//
//// GOBDecoder 是一个具体的 gob 解码器实现
//type GOBDecoder[T any] struct{}
//
//func (d GOBDecoder[T]) Decode(r io.Reader, v T) error {
//	return gob.NewDecoder(r).Decode(v)
//}
