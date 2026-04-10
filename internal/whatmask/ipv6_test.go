package whatmask

import "testing"

func TestParseIPv6(t *testing.T) {
	tests := []struct {
		input       string
		address     string
		addressFull string
		cidr        int
		network     string
		networkFull string
		last        string
		lastFull    string
		total       string
		addrType    string
	}{
		{
			"2001:db8::1/48",
			"2001:db8::1",
			"2001:0db8:0000:0000:0000:0000:0000:0001",
			48,
			"2001:db8::",
			"2001:0db8:0000:0000:0000:0000:0000:0000",
			"2001:db8:0:ffff:ffff:ffff:ffff:ffff",
			"2001:0db8:0000:ffff:ffff:ffff:ffff:ffff",
			"1208925819614629174706176",
			"Global Unicast",
		},
		{
			"fe80::1/10",
			"fe80::1",
			"fe80:0000:0000:0000:0000:0000:0000:0001",
			10,
			"fe80::",
			"fe80:0000:0000:0000:0000:0000:0000:0000",
			"febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			"febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			"332306998946228968225951765070086144",
			"Link-Local",
		},
		{
			"::1/128",
			"::1",
			"0000:0000:0000:0000:0000:0000:0000:0001",
			128,
			"::1",
			"0000:0000:0000:0000:0000:0000:0000:0001",
			"::1",
			"0000:0000:0000:0000:0000:0000:0000:0001",
			"1",
			"Loopback",
		},
		{
			"::/0",
			"::",
			"0000:0000:0000:0000:0000:0000:0000:0000",
			0,
			"::",
			"0000:0000:0000:0000:0000:0000:0000:0000",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			"340282366920938463463374607431768211456",
			"Unspecified",
		},
		{
			"fd12:3456:789a::1/64",
			"fd12:3456:789a::1",
			"fd12:3456:789a:0000:0000:0000:0000:0001",
			64,
			"fd12:3456:789a::",
			"fd12:3456:789a:0000:0000:0000:0000:0000",
			"fd12:3456:789a:0:ffff:ffff:ffff:ffff",
			"fd12:3456:789a:0000:ffff:ffff:ffff:ffff",
			"18446744073709551616",
			"Unique Local",
		},
		{
			"ff02::1/8",
			"ff02::1",
			"ff02:0000:0000:0000:0000:0000:0000:0001",
			8,
			"ff00::",
			"ff00:0000:0000:0000:0000:0000:0000:0000",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			"1329227995784915872903807060280344576",
			"Multicast",
		},
	}
	for _, tt := range tests {
		t.Run("ipv6_"+tt.input, func(t *testing.T) {
			res, err := ParseInput(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			v6, ok := res.(*IPv6Result)
			if !ok {
				t.Fatalf("expected IPv6Result, got %T", res)
			}
			if v6.Address != tt.address {
				t.Errorf("Address: got %q, want %q", v6.Address, tt.address)
			}
			if v6.AddressFull != tt.addressFull {
				t.Errorf("AddressFull: got %q, want %q", v6.AddressFull, tt.addressFull)
			}
			if v6.CIDR != tt.cidr {
				t.Errorf("CIDR: got %d, want %d", v6.CIDR, tt.cidr)
			}
			if v6.Network != tt.network {
				t.Errorf("Network: got %q, want %q", v6.Network, tt.network)
			}
			if v6.NetworkFull != tt.networkFull {
				t.Errorf("NetworkFull: got %q, want %q", v6.NetworkFull, tt.networkFull)
			}
			if v6.Last != tt.last {
				t.Errorf("Last: got %q, want %q", v6.Last, tt.last)
			}
			if v6.LastFull != tt.lastFull {
				t.Errorf("LastFull: got %q, want %q", v6.LastFull, tt.lastFull)
			}
			if v6.Total != tt.total {
				t.Errorf("Total: got %q, want %q", v6.Total, tt.total)
			}
			if v6.Type != tt.addrType {
				t.Errorf("Type: got %q, want %q", v6.Type, tt.addrType)
			}
		})
	}
}

func TestParseIPv6Invalid(t *testing.T) {
	tests := []string{
		"2001:db8::1",
		"::1",
		"2001:db8::1/129",
		"2001:db8::1/-1",
		"notanaddress::/64",
		"2001:db8::gggg/64",
	}
	for _, input := range tests {
		t.Run("invalid_ipv6_"+input, func(t *testing.T) {
			_, err := ParseInput(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}
