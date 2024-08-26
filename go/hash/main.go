package main 

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Create return a token along with it's hash value.
func Create(salt string) (string, string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", "", errors.Wrap(err, "[token.Create] failed to create a new UUID")
	}

	token := id.String()
	hash := Hash(token, salt)

	return token, hash, nil
}

// Hash will return the hash result of a token.
func Hash(token, salt string) string {
	saltedToken := fmt.Sprintf("%s.%s", token, salt)

	hash := sha256.New()
	hash.Write([]byte(saltedToken))

	result := hex.EncodeToString(hash.Sum(nil))

	return result
}

func main() {
	fmt.Println(Create("AwwLBAwGCQ0FDAMNBA4IDw"))
}
