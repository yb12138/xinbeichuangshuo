package main

import (
	"fmt"
	"starcup-engine/internal/rules"
)

func main() {
	deck := rules.InitDeck()
	fmt.Printf("Total cards: %d\n", len(deck))
	
	// Count by card type
	attackCount := 0
	magicCount := 0
	
	for _, card := range deck {
		if card.Type == "Attack" {
			attackCount++
		} else if card.Type == "Magic" {
			magicCount++
		}
	}
	
	fmt.Printf("Attack cards: %d\n", attackCount)
	fmt.Printf("Magic cards: %d\n", magicCount)
}
