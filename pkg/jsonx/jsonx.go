// Package jsonx is the single place where JSON is encoded or decoded.
// It wraps encoding/json/v2 and exposes adapters for Fiber.
package jsonx

import (
	jsonv2 "encoding/json/v2"
)

// Marshal encodes v with project defaults.
func Marshal(v any) ([]byte, error) {
	return jsonv2.Marshal(v)
}

// MarshalDeterministic encodes v with sorted map keys. Use it for artifacts
// that are compared byte-for-byte (OpenAPI golden file, snapshots).
func MarshalDeterministic(v any) ([]byte, error) {
	return jsonv2.Marshal(v, jsonv2.Deterministic(true))
}

// Unmarshal decodes data into v with project defaults.
func Unmarshal(data []byte, v any) error {
	return jsonv2.Unmarshal(data, v)
}

// UnmarshalStrict decodes data and rejects unknown members.
// Use it for every inbound DTO.
func UnmarshalStrict(data []byte, v any) error {
	return jsonv2.Unmarshal(data, v, jsonv2.RejectUnknownMembers(true))
}

// FiberEncoder allows encoding/json/v2 to be plugged into fiber.Config.
func FiberEncoder(v any) ([]byte, error) {
	return jsonv2.Marshal(v)
}

// FiberDecoder allows encoding/json/v2 to be plugged into fiber.Config.
func FiberDecoder(data []byte, v any) error {
	return jsonv2.Unmarshal(data, v)
}
