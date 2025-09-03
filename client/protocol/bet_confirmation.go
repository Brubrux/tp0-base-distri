package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Serialize BetConfirmation to bytes
func (b *BetConfirmation) ToBytes() []byte {
	var buf bytes.Buffer

	messageBytes := []byte(b.Message)
	payloadSize := uint32(1 + 1 + len(messageBytes)) // success(1) + messageLength(1) + message

	// Write OpCode (1 byte)
	binary.Write(&buf, binary.BigEndian, CONFIRM)

	// Write Payload Length (4 bytes)
	binary.Write(&buf, binary.BigEndian, payloadSize)

	// Write success flag (1 byte)
	var success uint8
	if b.Success {
		success = 1
	}
	binary.Write(&buf, binary.BigEndian, success)

	// Write message length and message
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

	// Skip payload length
	var payloadLength uint32
	if err := binary.Read(buf, binary.BigEndian, &payloadLength); err != nil {
		return bc, err
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
