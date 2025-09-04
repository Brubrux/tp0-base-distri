package protocol

type OpCode byte

const (
	REGISTER      OpCode = 0x00
	CONFIRM       OpCode = 0x01
	BATCH         OpCode = 0x02
	GET_WINNERS   OpCode = 0x03
	WINNERS       OpCode = 0x04
	NOT_CONDUCTED OpCode = 0x05

	TERMINATE OpCode = 0xFF
)

type BetRegister struct {
	Agency    uint8
	FirstName string
	LastName  string
	ID        string
	BirthDate string
	Number    string
}

type BetConfirmation struct {
	Success bool
	Message string
}
