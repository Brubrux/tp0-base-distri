package bet

import "bytes"

type OpCode byte

const (
	OpCodeBetRegistration OpCode = 0x00
	OpCodeBetConfirmation OpCode = 0x01
)

type Bet struct {
	FirstName string
	LastName  string
	ID        string
	BirthDate string
	Number    string
}

func (b *Bet) ToBytes() []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(OpCodeBetRegistration))

	buf.WriteByte(byte(len(b.FirstName)))
	buf.WriteString(b.FirstName)

	buf.WriteByte(byte(len(b.LastName)))
	buf.WriteString(b.LastName)

	buf.WriteByte(byte(len(b.ID)))
	buf.WriteString(b.ID)

	buf.WriteByte(byte(len(b.BirthDate)))
	buf.WriteString(b.BirthDate)

	buf.WriteByte(byte(len(b.Number)))
	buf.WriteString(b.Number)
	return buf.Bytes()
}
