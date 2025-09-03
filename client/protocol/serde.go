package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type OpCode byte

const (
	REGISTER OpCode = 0x00
	CONFIRM  OpCode = 0x01
)

func (b *BetRegister) ToBytes() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, REGISTER)

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
