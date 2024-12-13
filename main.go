package main

import (
	"fmt"
	"os"

	"github.com/metajar/trie-network/pkg/trie"
)

func main() {
	tmap := trie.NewIPTrie()

	metadata := map[string]interface{}{
		"region":      "us-east-1",
		"environment": "production",
		"owner":       "jared",
	}
	err := tmap.Insert("10.20.20.0/24", metadata)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = tmap.Insert("2001:dead:beef::2/120", metadata)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	matches, err := tmap.FindAll("10.20.20.100")
	for _, match := range matches {
		fmt.Printf("Matching CIDR: %s, Metadata: %v\n", match.CIDR, match.Metadata)
	}

	matches, err = tmap.FindAll("2001:dead:beef::ff")
	for _, match := range matches {
		fmt.Printf("Matching CIDR: %s, Metadata: %v\n", match.CIDR, match.Metadata)
	}
}
