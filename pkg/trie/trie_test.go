package trie

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
)

func TestIPv4Insertion(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		ip       string
		metadata map[string]interface{}
		want     bool
	}{
		{
			name: "basic IPv4 /24",
			cidr: "192.168.1.0/24",
			ip:   "192.168.1.100",
			metadata: map[string]interface{}{
				"region": "us-west",
			},
			want: true,
		},
		{
			name: "IPv4 outside range",
			cidr: "192.168.1.0/24",
			ip:   "192.168.2.100",
			metadata: map[string]interface{}{
				"region": "us-west",
			},
			want: false,
		},
		{
			name: "IPv4 /32 exact match",
			cidr: "192.168.1.1/32",
			ip:   "192.168.1.1",
			metadata: map[string]interface{}{
				"region": "us-west",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := NewIPTrie()
			err := trie.Insert(tt.cidr, tt.metadata)
			if err != nil {
				t.Fatalf("Failed to insert CIDR: %v", err)
			}

			_, metadata, err := trie.Find(tt.ip)
			if tt.want && (err != nil || metadata == nil) {
				t.Errorf("Expected to find IP %s in CIDR %s, but didn't", tt.ip, tt.cidr)
			} else if !tt.want && err == nil {
				t.Errorf("Expected not to find IP %s in CIDR %s, but did", tt.ip, tt.cidr)
			}
		})
	}
}

func TestIPv6Insertion(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		ip       string
		metadata map[string]interface{}
		want     bool
	}{
		{
			name: "basic IPv6 /120",
			cidr: "2001:dead:beef::2/120",
			ip:   "2001:dead:beef::ff",
			metadata: map[string]interface{}{
				"region": "eu-west",
			},
			want: true,
		},
		{
			name: "IPv6 outside range",
			cidr: "2001:dead:beef::2/120",
			ip:   "2001:dead:beef:1::2",
			metadata: map[string]interface{}{
				"region": "eu-west",
			},
			want: false,
		},
		{
			name: "IPv6 /128 exact match",
			cidr: "2001:dead:beef::2/128",
			ip:   "2001:dead:beef::2",
			metadata: map[string]interface{}{
				"region": "eu-west",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := NewIPTrie()
			err := trie.Insert(tt.cidr, tt.metadata)
			if err != nil {
				t.Fatalf("Failed to insert CIDR: %v", err)
			}

			_, metadata, err := trie.Find(tt.ip)
			if tt.want && (err != nil || metadata == nil) {
				t.Errorf("Expected to find IP %s in CIDR %s, but didn't", tt.ip, tt.cidr)
			} else if !tt.want && err == nil {
				t.Errorf("Expected not to find IP %s in CIDR %s, but did", tt.ip, tt.cidr)
			}
		})
	}
}

func TestOverlappingRanges(t *testing.T) {
	trie := NewIPTrie()

	// Insert overlapping ranges
	ranges := []struct {
		cidr     string
		metadata map[string]interface{}
	}{
		{
			cidr: "192.168.0.0/16",
			metadata: map[string]interface{}{
				"scope": "wide",
			},
		},
		{
			cidr: "192.168.1.0/24",
			metadata: map[string]interface{}{
				"scope": "narrow",
			},
		},
	}

	for _, r := range ranges {
		err := trie.Insert(r.cidr, r.metadata)
		if err != nil {
			t.Fatalf("Failed to insert CIDR: %v", err)
		}
	}

	matches, err := trie.FindAll("192.168.1.100")
	if err != nil {
		t.Fatalf("Failed to find IP: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}
}

// Benchmarks
func BenchmarkIPv4Insert(b *testing.B) {
	trie := NewIPTrie()
	metadata := map[string]interface{}{"region": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cidr := fmt.Sprintf("192.168.%d.0/24", i%256)
		_ = trie.Insert(cidr, metadata)
	}
}

func BenchmarkIPv6Insert(b *testing.B) {
	trie := NewIPTrie()
	metadata := map[string]interface{}{"region": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cidr := fmt.Sprintf("2001:dead:beef:%d::0/64", i%65536)
		_ = trie.Insert(cidr, metadata)
	}
}

func BenchmarkIPv4Find(b *testing.B) {
	trie := NewIPTrie()
	metadata := map[string]interface{}{"region": "test"}

	// Insert some test data
	for i := 0; i < 1000; i++ {
		cidr := fmt.Sprintf("192.168.%d.0/24", i%256)
		_ = trie.Insert(cidr, metadata)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := fmt.Sprintf("192.168.%d.%d", i%256, i%256)
		_, _, _ = trie.Find(ip)
	}
}

func BenchmarkIPv6Find(b *testing.B) {
	trie := NewIPTrie()
	metadata := map[string]interface{}{"region": "test"}

	// Insert some test data
	for i := 0; i < 1000; i++ {
		cidr := fmt.Sprintf("2001:dead:beef:%d::0/64", i%65536)
		_ = trie.Insert(cidr, metadata)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := fmt.Sprintf("2001:dead:beef:%d::%d", i%65536, i%65536)
		_, _, _ = trie.Find(ip)
	}
}

func BenchmarkLargeScale(b *testing.B) {
	b.Run("1K_CIDRs", func(b *testing.B) { benchmarkWithSize(b, 1000) })
	b.Run("10K_CIDRs", func(b *testing.B) { benchmarkWithSize(b, 10000) })
	b.Run("100K_CIDRs", func(b *testing.B) { benchmarkWithSize(b, 100000) })
}

func benchmarkWithSize(b *testing.B, size int) {
	trie := NewIPTrie()
	metadata := map[string]interface{}{"region": "test"}

	// Generate random CIDRs
	for i := 0; i < size; i++ {
		ip := make(net.IP, 4)
		rand.Read(ip)
		cidr := fmt.Sprintf("%s/24", ip.String())
		_ = trie.Insert(cidr, metadata)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ip := make(net.IP, 4)
		rand.Read(ip)
		_, _, _ = trie.Find(ip.String())
	}
}
