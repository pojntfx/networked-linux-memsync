package utils

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

func EncodeJSONFixedLength(writer io.Writer, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := binary.Write(writer, binary.LittleEndian, int32(len(b))); err != nil {
		return err
	}

	_, err = writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func DecodeJSONFixedLength(reader io.Reader, data any) error {
	var bLen int32
	if err := binary.Read(reader, binary.LittleEndian, &bLen); err != nil {
		return err
	}

	b := make([]byte, bLen)
	if _, err := io.ReadFull(reader, b); err != nil {
		return err
	}

	return json.Unmarshal(b, data)
}
