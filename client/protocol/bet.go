package protocol

type BetRegister struct {
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
