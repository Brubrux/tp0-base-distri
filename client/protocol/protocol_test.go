package protocol

import (
	"testing"
)

// SerDe tests

// Register
func TestBetRegisterSerialization(t *testing.T) {
	bet := BetRegister{
		FirstName: "Santiago Lionel",
		LastName:  "Lorca",
		ID:        "30904465",
		BirthDate: "1999-03-17",
		Number:    "7574",
	}

	// Serialize
	data := bet.ToBytes()

	// Deserialize
	deserializedBet, err := DeserializeRegister(data)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if bet != deserializedBet {
		t.Errorf("Original and deserialized BetRegister do not match:\nOriginal: %+v\nDeserialized: %+v",
			bet, deserializedBet)
	}
}

// Confirmation
func TestBetConfirmationSerialization(t *testing.T) {
	confirmation := BetConfirmation{
		Success: true,
		Message: "",
	}

	// Serialize
	data := confirmation.ToBytes()

	// Deserialize
	deserializedConfirmation, err := DeserializeConfirmation(data)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if confirmation != deserializedConfirmation {
		t.Errorf("Original and deserialized BetConfirmation do not match:\nOriginal: %+v\nDeserialized: %+v",
			confirmation, deserializedConfirmation)
	}
}
