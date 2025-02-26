package auth

import (
    "testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestHashPassword(t *testing.T) {
	my_passwd := "Egorhahaha"
	hash, err := HashPassword(my_passwd)
	if err != nil {
		t.Fatalf(`Something has gone wrong when hashing password`)
	}
	err = CheckPasswordHash(my_passwd, hash)
	if err != nil {
		t.Fatalf(`Something has gone wrong when comparing passwords`)
	}
}
