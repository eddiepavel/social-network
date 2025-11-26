package helpers

import (
	"errors"

	"github.com/gofrs/uuid"
)

func GenerateFromBytes(data []byte) (string, error) {
	var id uuid.UUID

	err := id.UnmarshalBinary(data)

	if err != nil {
		return "", errors.New("failed to convert uuid to string")
	}

	return id.String(), nil
}

func GenerateFromString(data string) ([]byte, error) {
	id, err := uuid.FromString(data)
	if err != nil {
		return nil, errors.New("failed to parse uuid string")
	}

	// Then marshal to binary
	binary, err := id.MarshalBinary()
	if err != nil {
		return nil, errors.New("failed to convert uuid to binary")
	}

	return binary, nil
}
