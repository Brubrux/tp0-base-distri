package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func (b *BetRegister) ToBytes() []byte {
	var buf bytes.Buffer

	payloadSize := b.getPayloadSize()

	// Write OpCode (1 byte)
	binary.Write(&buf, binary.BigEndian, REGISTER)

	// Write Payload Length (4 bytes)
	binary.Write(&buf, binary.BigEndian, payloadSize)

	// Write Agency (1 byte)
	binary.Write(&buf, binary.BigEndian, b.Agency)

	writeField := func(field string) {
		fieldBytes := []byte(field)
		binary.Write(&buf, binary.BigEndian, uint8(len(fieldBytes)))
		buf.Write(fieldBytes)
	}

	writeField(b.FirstName)
	writeField(b.LastName)
	writeField(b.ID)
	writeField(b.BirthDate)
	writeField(b.Number)

	return buf.Bytes()
}

func DeserializeRegister(data []byte) (BetRegister, error) {
	var br BetRegister
	buf := bytes.NewBuffer(data)

	var opCode OpCode
	if err := binary.Read(buf, binary.BigEndian, &opCode); err != nil {
		return br, err
	}
	if opCode != REGISTER {
		return br, fmt.Errorf("invalid OpCode: %d", opCode)
	}

	// Skip payload length
	var payloadLength uint32
	if err := binary.Read(buf, binary.BigEndian, &payloadLength); err != nil {
		return br, err
	}

	var agency uint8
	if err := binary.Read(buf, binary.BigEndian, &agency); err != nil {
		return br, err
	}
	br.Agency = agency

	readField := func() (string, error) {
		var length uint8
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return "", err
		}
		fieldBytes := make([]byte, length)
		if _, err := buf.Read(fieldBytes); err != nil {
			return "", err
		}
		return string(fieldBytes), nil
	}

	var err error
	br.FirstName, err = readField()
	if err != nil {
		return br, err
	}
	br.LastName, err = readField()
	if err != nil {
		return br, err
	}
	br.ID, err = readField()
	if err != nil {
		return br, err
	}
	br.BirthDate, err = readField()
	if err != nil {
		return br, err
	}
	br.Number, err = readField()
	if err != nil {
		return br, err
	}

	return br, nil
}

func (b *BetRegister) getPayloadSize() uint32 {
	payloadSize := uint32(0)

	// Agency: 1 byte
	payloadSize += 1

	// 1 byte (length) + string bytes
	payloadSize += 1 + uint32(len(b.FirstName))
	payloadSize += 1 + uint32(len(b.LastName))
	payloadSize += 1 + uint32(len(b.ID))
	payloadSize += 1 + uint32(len(b.BirthDate))
	payloadSize += 1 + uint32(len(b.Number))
	return payloadSize
}
