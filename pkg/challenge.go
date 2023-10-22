package pkg

import "math/rand"

func GenerateChallenge() (int64, int) {
	challenge := rand.Int63n(10000)
	return challenge, 5
}
