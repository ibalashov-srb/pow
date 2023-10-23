package pkg

import "math/rand"

type Challenger struct {
	leadingZeros int
	keyRange     int64
}

func NewChallenger(leadingZeros int, keyRange int64) *Challenger {
	return &Challenger{
		leadingZeros: leadingZeros,
		keyRange:     keyRange,
	}
}

func (c *Challenger) GenerateChallenge() (int64, int) {
	challenge := rand.Int63n(c.keyRange)
	return challenge, c.leadingZeros
}
