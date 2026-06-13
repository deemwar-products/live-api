package live

import "encoding/json"

// jsonMarshal is a tiny indirection so handler.go can use a one-line
// helper without importing encoding/json itself.
func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }
