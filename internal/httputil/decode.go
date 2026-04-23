package httputil

import (
	"encoding/json"
	"io"
)

const defaultMaxBodyBytes = 1 << 20 // 1MB

func DecodeJSON(r io.Reader, dst any) error {
	return json.NewDecoder(io.LimitReader(r, defaultMaxBodyBytes)).Decode(dst)
}
