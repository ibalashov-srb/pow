package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
)

var ErrMaxInt64Reached = errors.New("reached max int 64")

func CalculateHashWithLeadingZeros(data int64, nonce int64, leadingZeros int) (string, int64, error) {
	for {
		hashInput := fmt.Sprintf("%v%v", data, nonce)
		hash := sha256.Sum256([]byte(hashInput))
		hashString := hex.EncodeToString(hash[:])

		if strings.HasPrefix(hashString, strings.Repeat("0", leadingZeros)) {
			return hashString, nonce, nil
		}

		nonce++
		if nonce == math.MaxInt64 {
			return "", 0, ErrMaxInt64Reached
		}
	}
}
