package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Serialize BetConfirmation to bytes
func (b *BetConfirmation) ToBytes() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, CONFIRM)

	var success uint8
	if b.Success {
		success = 1
	}
	binary.Write(&buf, binary.BigEndian, success)

	messageBytes := []byte(b.Message)
	binary.Write(&buf, binary.BigEndian, uint8(len(messageBytes)))
	buf.Write(messageBytes)

	return buf.Bytes()
}

// Deserializes BetConfirmation from bytes
func DeserializeConfirmation(data []byte) (BetConfirmation, error) {
	var bc BetConfirmation
	buf := bytes.NewBuffer(data)

	// OpCode check
	var opCode OpCode
	if err := binary.Read(buf, binary.BigEndian, &opCode); err != nil {
		return bc, err
	}
	if opCode != CONFIRM {
		return bc, fmt.Errorf("invalid OpCode: %d", opCode)
	}

	// Success bool
	var success uint8
	if err := binary.Read(buf, binary.BigEndian, &success); err != nil {
		return bc, err
	}
	bc.Success = success == 1

	// Message length
	var messageLength uint8
	if err := binary.Read(buf, binary.BigEndian, &messageLength); err != nil {
		return bc, err
	}
	if messageLength > 0 {
		messageBytes := make([]byte, messageLength)
		if _, err := buf.Read(messageBytes); err != nil {
			return bc, err
		}
		bc.Message = string(messageBytes)
	}

	return bc, nil
}
