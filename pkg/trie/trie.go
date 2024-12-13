package trie

import (
	"fmt"
	"net"
)

// Node represents a node in the IP trie
type Node struct {
	children map[byte]*Node
	isEnd    bool
	metadata map[string]interface{}
	cidr     string
}

// IPTrie represents the main trie structure
type IPTrie struct {
	root *Node
}

// NewIPTrie creates a new IP trie
func NewIPTrie() *IPTrie {
	return &IPTrie{
		root: &Node{
			children: make(map[byte]*Node),
			metadata: make(map[string]interface{}),
		},
	}
}

// ipToBytes converts an IP address to a slice of bytes for trie traversal
func ipToBytes(ip net.IP) []byte {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4
	}
	return ip.To16()
}

// Insert adds an IP CIDR with metadata to the trie
func (t *IPTrie) Insert(cidr string, metadata map[string]interface{}) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %v", err)
	}

	node := t.root
	ipBytes := ipToBytes(ipnet.IP)
	ones, total := ipnet.Mask.Size()

	// Convert IP to bits and insert into trie
	for i := 0; i < ones; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		bit := (ipBytes[byteIndex] >> uint(bitIndex)) & 1

		if node.children[bit] == nil {
			node.children[bit] = &Node{
				children: make(map[byte]*Node),
				metadata: make(map[string]interface{}),
			}
		}
		node = node.children[bit]
	}

	// For exact matches (/32 IPv4 or /128 IPv6), we need to handle remaining bits
	if ones == total {
		for i := ones; i < total; i++ {
			byteIndex := i / 8
			bitIndex := 7 - (i % 8)
			bit := (ipBytes[byteIndex] >> uint(bitIndex)) & 1

			if node.children[bit] == nil {
				node.children[bit] = &Node{
					children: make(map[byte]*Node),
					metadata: make(map[string]interface{}),
				}
			}
			node = node.children[bit]
		}
	}

	node.isEnd = true
	node.cidr = cidr
	node.metadata = metadata

	return nil
}

// Find searches for an IP address and returns matching CIDR and metadata
func (t *IPTrie) Find(ip string) (string, map[string]interface{}, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", nil, fmt.Errorf("invalid IP address")
	}

	node := t.root
	var lastMatch *Node
	ipBytes := ipToBytes(parsedIP)
	totalBits := len(ipBytes) * 8

	for i := 0; i < totalBits; i++ {
		if node.isEnd {
			lastMatch = node
		}

		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		bit := (ipBytes[byteIndex] >> uint(bitIndex)) & 1

		node = node.children[bit]
		if node == nil {
			break
		}
	}

	// Check the last node in case it's an exact match
	if node != nil && node.isEnd {
		lastMatch = node
	}

	if lastMatch == nil {
		return "", nil, fmt.Errorf("no matching CIDR found")
	}

	return lastMatch.cidr, lastMatch.metadata, nil
}

// FindAll returns all matching CIDRs and their metadata for an IP
func (t *IPTrie) FindAll(ip string) ([]struct {
	CIDR     string
	Metadata map[string]interface{}
}, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	var matches []struct {
		CIDR     string
		Metadata map[string]interface{}
	}

	node := t.root
	ipBytes := ipToBytes(parsedIP)
	totalBits := len(ipBytes) * 8

	for i := 0; i < totalBits; i++ {
		if node.isEnd {
			matches = append(matches, struct {
				CIDR     string
				Metadata map[string]interface{}
			}{
				CIDR:     node.cidr,
				Metadata: node.metadata,
			})
		}

		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		bit := (ipBytes[byteIndex] >> uint(bitIndex)) & 1

		node = node.children[bit]
		if node == nil {
			break
		}
	}

	// Check the last node in case it's an exact match
	if node != nil && node.isEnd {
		matches = append(matches, struct {
			CIDR     string
			Metadata map[string]interface{}
		}{
			CIDR:     node.cidr,
			Metadata: node.metadata,
		})
	}

	return matches, nil
}

// Delete removes a CIDR and its metadata from the trie
func (t *IPTrie) Delete(cidr string) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %v", err)
	}

	var nodes []*Node
	node := t.root
	ipBytes := ipToBytes(ipnet.IP)
	ones, total := ipnet.Mask.Size()
	totalBits := ones
	if ones == total {
		totalBits = len(ipBytes) * 8
	}

	// Collect nodes along the path
	for i := 0; i < totalBits; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		bit := (ipBytes[byteIndex] >> uint(bitIndex)) & 1

		if node.children[bit] == nil {
			return fmt.Errorf("CIDR not found")
		}
		nodes = append(nodes, node)
		node = node.children[bit]
	}

	// Remove the end marker and clean up empty nodes
	if !node.isEnd {
		return fmt.Errorf("CIDR not found")
	}

	node.isEnd = false
	node.metadata = make(map[string]interface{})
	node.cidr = ""

	// Clean up empty branches
	for i := len(nodes) - 1; i >= 0; i-- {
		parent := nodes[i]
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		bit := (ipBytes[byteIndex] >> uint(bitIndex)) & 1

		child := parent.children[bit]
		if len(child.children) == 0 && !child.isEnd {
			delete(parent.children, bit)
		}
	}

	return nil
}
