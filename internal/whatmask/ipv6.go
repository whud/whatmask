// IPv6 subnet calculation for whatmask.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package whatmask

import (
	"fmt"
	"math/big"
	"net/netip"
	"strconv"
	"strings"
)

// IPv6Result holds IPv6 address+prefix output.
type IPv6Result struct {
	Address     string `json:"address"`
	AddressFull string `json:"address_full"`
	CIDR        int    `json:"cidr"`
	Network     string `json:"network"`
	NetworkFull string `json:"network_full"`
	Last        string `json:"last"`
	LastFull    string `json:"last_full"`
	Total       string `json:"total"`
	Type        string `json:"type"`
}

func (r *IPv6Result) Mode() string { return "network6" }

func parseIPv6(input string) (*IPv6Result, error) {
	parts := strings.SplitN(input, "/", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidInput
	}

	addr, err := netip.ParseAddr(parts[0])
	if err != nil || !addr.Is6() {
		return nil, ErrInvalidInput
	}

	cidr, err := strconv.Atoi(parts[1])
	if err != nil || cidr < 0 || cidr > 128 {
		return nil, ErrInvalidInput
	}

	prefix := netip.PrefixFrom(addr, cidr).Masked()
	networkAddr := prefix.Addr()

	// Calculate last address: network OR (all-ones in host bits)
	networkBytes := networkAddr.As16()
	var lastBytes [16]byte
	copy(lastBytes[:], networkBytes[:])
	for i := cidr; i < 128; i++ {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)
		lastBytes[byteIdx] |= 1 << bitIdx
	}
	lastAddr := netip.AddrFrom16(lastBytes)

	// Total addresses: 2^(128-cidr)
	total := new(big.Int).Lsh(big.NewInt(1), uint(128-cidr))

	return &IPv6Result{
		Address:     addr.String(),
		AddressFull: expandIPv6(addr),
		CIDR:        cidr,
		Network:     networkAddr.String(),
		NetworkFull: expandIPv6(networkAddr),
		Last:        lastAddr.String(),
		LastFull:    expandIPv6(lastAddr),
		Total:       total.String(),
		Type:        classifyIPv6(addr),
	}, nil
}

func expandIPv6(addr netip.Addr) string {
	b := addr.As16()
	return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
		b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15])
}

func classifyIPv6(addr netip.Addr) string {
	b := addr.As16()

	if addr == netip.MustParseAddr("::1") {
		return "Loopback"
	}
	if addr == netip.MustParseAddr("::") {
		return "Unspecified"
	}
	if b[0] == 0xff {
		return "Multicast"
	}
	if b[0] == 0xfe && (b[1]&0xc0) == 0x80 {
		return "Link-Local"
	}
	if b[0]&0xfe == 0xfc {
		return "Unique Local"
	}
	if (b[0] & 0xe0) == 0x20 {
		return "Global Unicast"
	}
	return "Reserved"
}
