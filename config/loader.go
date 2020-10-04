package config

import (
	"bytes"
	"encoding/json"
	"io"

	"v2ray.com/core/common/errors"
	json_reader "v2ray.com/core/infra/conf/json"
)

type offset struct {
	line int
	char int
}

func findOffset(b []byte, o int) *offset {
	if o >= len(b) || o < 0 {
		return nil
	}

	line := 1
	char := 0
	for i, x := range b {
		if i == o {
			break
		}
		if x == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}

	return &offset{line: line, char: char}
}

// decodeJSONConfig reads from reader and decode the config into config
// syntax error could be detected.
func decodeJSONConfig(reader io.Reader, config interface{}) error {
	jsonContent := bytes.NewBuffer(make([]byte, 0, 10240))
	jsonReader := io.TeeReader(&json_reader.Reader{
		Reader: reader,
	}, jsonContent)
	decoder := json.NewDecoder(jsonReader)

	if err := decoder.Decode(config); err != nil {
		var pos *offset
		cause := errors.Cause(err)
		switch tErr := cause.(type) {
		case *json.SyntaxError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		case *json.UnmarshalTypeError:
			pos = findOffset(jsonContent.Bytes(), int(tErr.Offset))
		}
		if pos != nil {
			return errors.New("failed to read config file at line ", pos.line, " char ", pos.char).Base(err)
		}
		return errors.New("failed to read config file").Base(err)
	}

	return nil
}
