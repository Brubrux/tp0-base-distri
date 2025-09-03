package protocol

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
