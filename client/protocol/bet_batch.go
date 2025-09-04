package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	MAXPACKETSIZE      = 8 * 1024 // 8KB
	PROTOCOL_OVERHEAD  = 1 + 4    // opcode + payload length
	MESSAGE_OVERHEAD   = 1 + 4    // agency id + bet count
	LINE_LENGTH_HEADER = 1
)

type BetBatchRegister struct {
	agencyID    uint8
	betLines    [][]byte
	packetSize  int
	maxBetLines int
}

func NewBatch(agencyID uint8, maxBetLines int) *BetBatchRegister {
	return &BetBatchRegister{
		agencyID:    agencyID,
		betLines:    make([][]byte, 0),
		maxBetLines: maxBetLines,
		packetSize:  PROTOCOL_OVERHEAD + MESSAGE_OVERHEAD,
	}
}

// AddBetLine adds a new bet line to the batch if it fits within the max packet size and the max bet lines limit.
// Returns true if the bet line was added successfully, false otherwise.
func (b *BetBatchRegister) AddBetLine(betLine string) bool {

	if len(b.betLines) >= b.maxBetLines {
		return false
	}

	bet := []byte(betLine)

	betLineSize := LINE_LENGTH_HEADER + len(bet)
	if b.packetSize+betLineSize > MAXPACKETSIZE {
		return false
	}
	b.betLines = append(b.betLines, bet)
	b.packetSize += betLineSize
	return true
}

func (b *BetBatchRegister) ToBytes() []byte {
	var buf bytes.Buffer

	// protocol
	buf.WriteByte(byte(BATCH))
	payloadLength := uint32(b.packetSize - PROTOCOL_OVERHEAD)
	binary.Write(&buf, binary.BigEndian, payloadLength)

	// message
	buf.WriteByte(b.agencyID)
	binary.Write(&buf, binary.BigEndian, uint32(len(b.betLines)))

	// Write each bet line
	for _, line := range b.betLines {
		buf.WriteByte(byte(len(line)))
		buf.Write(line)
	}

	return buf.Bytes()
}

func (b *BetBatchRegister) GetBetCount() int {
	return len(b.betLines)
}
