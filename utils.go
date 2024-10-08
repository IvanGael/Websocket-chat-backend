package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

func generateUniqueRoomID() string {
	for {
		// Generate 3 parts of lowercase letters
		part1 := generateRandomLetters(3)
		part2 := generateRandomLetters(4)
		part3 := generateRandomLetters(3)

		// Generate random 3-digit number
		hashNum := generateRandomNumber(100, 999)

		// Combine parts
		id := fmt.Sprintf("%s-%s-%s?hs=%d", part1, part2, part3, hashNum)

		roomsMutex.RLock()
		_, exists := rooms[id]
		roomsMutex.RUnlock()

		if !exists {
			return id
		}
	}
}

func generateRandomLetters(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		randIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[randIndex.Int64()]
	}
	return string(b)
}

func generateRandomNumber(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}

func isValidRoomID(id string) bool {
	pattern := regexp.MustCompile(`^[a-z]{3}-[a-z]{4}-[a-z]{3}\?hs=[1-9]\d{2}$`)
	return pattern.MatchString(id)
}

func extractBaseRoomID(fullRoomID string) string {
	parts := strings.Split(fullRoomID, "?")
	return parts[0]
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
