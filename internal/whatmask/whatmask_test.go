package whatmask

import "testing"

func TestParseMaskCIDR(t *testing.T) {
	tests := []struct {
		input    string
		cidr     int
		netmask  string
		hex      string
		wildcard string
		usable   int
	}{
		{"/24", 24, "255.255.255.0", "0xffffff00", "0.0.0.255", 254},
		{"/32", 32, "255.255.255.255", "0xffffffff", "0.0.0.0", 1},
		{"/31", 31, "255.255.255.254", "0xfffffffe", "0.0.0.1", 1},
		{"/30", 30, "255.255.255.252", "0xfffffffc", "0.0.0.3", 2},
		{"/16", 16, "255.255.0.0", "0xffff0000", "0.0.255.255", 65534},
		{"/0", 0, "0.0.0.0", "0x00000000", "255.255.255.255", 1},
		{"/1", 1, "128.0.0.0", "0x80000000", "127.255.255.255", 1},
	}
	for _, tt := range tests {
		t.Run("cidr_"+tt.input, func(t *testing.T) {
			res, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			mask, ok := res.(*MaskResult)
			if !ok {
				t.Fatalf("expected MaskResult, got %T", res)
			}
			if mask.CIDR != tt.cidr {
				t.Errorf("CIDR: got %d, want %d", mask.CIDR, tt.cidr)
			}
			if mask.Netmask != tt.netmask {
				t.Errorf("Netmask: got %s, want %s", mask.Netmask, tt.netmask)
			}
			if mask.Hex != tt.hex {
				t.Errorf("Hex: got %s, want %s", mask.Hex, tt.hex)
			}
			if mask.Wildcard != tt.wildcard {
				t.Errorf("Wildcard: got %s, want %s", mask.Wildcard, tt.wildcard)
			}
			if mask.Usable != tt.usable {
				t.Errorf("Usable: got %d, want %d", mask.Usable, tt.usable)
			}
		})
	}
}

func TestParseMaskDotted(t *testing.T) {
	tests := []struct {
		input string
		cidr  int
	}{
		{"255.255.255.0", 24},
		{"255.255.0.0", 16},
		{"255.0.0.0", 8},
		{"255.255.255.255", 32},
		{"255.255.255.252", 30},
	}
	for _, tt := range tests {
		t.Run("dotted_"+tt.input, func(t *testing.T) {
			res, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			mask, ok := res.(*MaskResult)
			if !ok {
				t.Fatalf("expected MaskResult, got %T", res)
			}
			if mask.CIDR != tt.cidr {
				t.Errorf("CIDR: got %d, want %d", mask.CIDR, tt.cidr)
			}
		})
	}
}

func TestParseMaskHex(t *testing.T) {
	tests := []struct {
		input string
		cidr  int
	}{
		{"0xffffff00", 24},
		{"0xffff0000", 16},
		{"0xffffffff", 32},
	}
	for _, tt := range tests {
		t.Run("hex_"+tt.input, func(t *testing.T) {
			res, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			mask, ok := res.(*MaskResult)
			if !ok {
				t.Fatalf("expected MaskResult, got %T", res)
			}
			if mask.CIDR != tt.cidr {
				t.Errorf("CIDR: got %d, want %d", mask.CIDR, tt.cidr)
			}
		})
	}
}

func TestParseMaskWildcard(t *testing.T) {
	tests := []struct {
		input string
		cidr  int
	}{
		{"0.0.0.255", 24},
		{"0.0.255.255", 16},
		{"0.0.0.0", 32},
	}
	for _, tt := range tests {
		t.Run("wildcard_"+tt.input, func(t *testing.T) {
			res, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			mask, ok := res.(*MaskResult)
			if !ok {
				t.Fatalf("expected MaskResult, got %T", res)
			}
			if mask.CIDR != tt.cidr {
				t.Errorf("CIDR: got %d, want %d", mask.CIDR, tt.cidr)
			}
		})
	}
}

func TestParseInvalidInput(t *testing.T) {
	tests := []string{
		"",
		"abc",
		"24",
		"33",
		"0",
		"-1",
		"256.0.0.0",
		"999.999.999.999",
		"<script>",
		"192.168.1.0/24/extra",
	}
	for _, input := range tests {
		t.Run("invalid_"+input, func(t *testing.T) {
			_, err := ParseInput(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestParseNetwork(t *testing.T) {
	tests := []struct {
		input     string
		address   string
		cidr      int
		network   string
		broadcast string
		first     string
		last      string
		usable    int
	}{
		{"192.168.1.100/24", "192.168.1.100", 24, "192.168.1.0", "192.168.1.255", "192.168.1.1", "192.168.1.254", 254},
		{"10.0.0.1/8", "10.0.0.1", 8, "10.0.0.0", "10.255.255.255", "10.0.0.1", "10.255.255.254", 16777214},
		{"172.16.0.1/255.255.255.0", "172.16.0.1", 24, "172.16.0.0", "172.16.0.255", "172.16.0.1", "172.16.0.254", 254},
		{"192.168.1.0/0xffffff00", "192.168.1.0", 24, "192.168.1.0", "192.168.1.255", "192.168.1.1", "192.168.1.254", 254},
		{"192.168.1.0/0.0.0.255", "192.168.1.0", 24, "192.168.1.0", "192.168.1.255", "192.168.1.1", "192.168.1.254", 254},
		{"10.10.10.10/32", "10.10.10.10", 32, "10.10.10.10", "10.10.10.10", "10.10.10.10", "10.10.10.10", 1},
	}
	for _, tt := range tests {
		t.Run("network_"+tt.input, func(t *testing.T) {
			res, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			net, ok := res.(*NetworkResult)
			if !ok {
				t.Fatalf("expected NetworkResult, got %T", res)
			}
			if net.Address != tt.address {
				t.Errorf("Address: got %s, want %s", net.Address, tt.address)
			}
			if net.CIDR != tt.cidr {
				t.Errorf("CIDR: got %d, want %d", net.CIDR, tt.cidr)
			}
			if net.Network != tt.network {
				t.Errorf("Network: got %s, want %s", net.Network, tt.network)
			}
			if net.Broadcast != tt.broadcast {
				t.Errorf("Broadcast: got %s, want %s", net.Broadcast, tt.broadcast)
			}
			if net.First != tt.first {
				t.Errorf("First: got %s, want %s", net.First, tt.first)
			}
			if net.Last != tt.last {
				t.Errorf("Last: got %s, want %s", net.Last, tt.last)
			}
			if net.Usable != tt.usable {
				t.Errorf("Usable: got %d, want %d", net.Usable, tt.usable)
			}
		})
	}
}
