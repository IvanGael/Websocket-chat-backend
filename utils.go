package main

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/rs/xid"
)

func generateUniqueRoomID() string {
	for {
		id := xid.New().String()
		roomsMutex.RLock()
		_, exists := rooms[id]
		roomsMutex.RUnlock()
		if !exists {
			return id
		}
	}
}

func generateNickname() string {
	adjectives := []string{"Happy", "Sleepy", "Grumpy", "Sneezy", "Bashful", "Dopey", "Doc"}
	nouns := []string{"Panda", "Koala", "Penguin", "Tiger", "Lion", "Elephant", "Giraffe"}

	adjectiveIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	nounIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	number, _ := rand.Int(rand.Reader, big.NewInt(1000))

	return fmt.Sprintf("%s%s%03d", adjectives[adjectiveIndex.Int64()], nouns[nounIndex.Int64()], number.Int64())
}

func generateColor() string {
	r, _ := rand.Int(rand.Reader, big.NewInt(256))
	g, _ := rand.Int(rand.Reader, big.NewInt(256))
	b, _ := rand.Int(rand.Reader, big.NewInt(256))
	return fmt.Sprintf("#%02x%02x%02x", r.Int64(), g.Int64(), b.Int64())
}
