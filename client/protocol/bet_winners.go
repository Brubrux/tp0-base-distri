package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Winners struct {
	WinnersCount uint32
	WinnerIDs    []uint32
}

// NewGetWinnersMessage creates a GetWinners message for the specified agency
func NewGetWinnersMessage(agencyID uint8) []byte {
	var buf bytes.Buffer

	payloadSize := uint32(1)

	binary.Write(&buf, binary.BigEndian, GET_WINNERS)
	binary.Write(&buf, binary.BigEndian, payloadSize)
	binary.Write(&buf, binary.BigEndian, agencyID)

	return buf.Bytes()
}

// Deserializes the response after a GetWinners request
// If the response indicates winners are available, it returns the winners list. Else it returns false.
func DeserializeWinnersResponse(data []byte) (bool, *Winners, error) {
	if len(data) < 5 {
		return false, nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	buf := bytes.NewBuffer(data)

	var opCode OpCode
	if err := binary.Read(buf, binary.BigEndian, &opCode); err != nil {
		return false, nil, fmt.Errorf("failed to read opcode: %v", err)
	}

	var payloadLength uint32
	if err := binary.Read(buf, binary.BigEndian, &payloadLength); err != nil {
		return false, nil, fmt.Errorf("failed to read payload length: %v", err)
	}

	switch opCode {
	case NOT_CONDUCTED:
		// OpCode 0x05 - Lottery not conducted yet
		return false, nil, nil

	case WINNERS:
		// OpCode 0x04 - Winners available

		var winnersCount uint32
		if err := binary.Read(buf, binary.BigEndian, &winnersCount); err != nil {
			return false, nil, fmt.Errorf("failed to read winners count: %v", err)
		}

		// ID list
		winnerIDs := make([]uint32, winnersCount)
		for i := uint32(0); i < winnersCount; i++ {
			var id uint32
			if err := binary.Read(buf, binary.BigEndian, &id); err != nil {
				return false, nil, fmt.Errorf("failed to read ID %d: %v", i, err)
			}
			winnerIDs[i] = id
		}

		winners := &Winners{
			WinnersCount: winnersCount,
			WinnerIDs:    winnerIDs,
		}

		return true, winners, nil

	default:
		return false, nil, fmt.Errorf("unexpected opcode: 0x%02X", opCode)
	}
}
