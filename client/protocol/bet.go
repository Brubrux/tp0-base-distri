package protocol

type OpCode byte

const (
	REGISTER OpCode = 0x00
	CONFIRM  OpCode = 0x01
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
