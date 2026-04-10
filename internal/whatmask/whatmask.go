// Whatmask — subnet calculator.
//
// Original C implementation by Joe Laffey (laffeycomputer.com/whatmask.html).
// Ruby rewrite by Joe Topjian (github.com/geezyx/whatmask).
// Go rewrite for web service.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package whatmask

import (
	"errors"
	"fmt"
	"math/bits"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	inputAllowlist  = regexp.MustCompile(`^[0-9a-fA-Fx./:]{1,50}$`)
	dottedQuad      = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	hexMask         = regexp.MustCompile(`^0x[0-9a-fA-F]{1,8}$`)
)

// Result is the interface returned by ParseInput.
type Result interface {
	Mode() string
}

// MaskResult holds mask-only output.
type MaskResult struct {
	CIDR     int    `json:"cidr"`
	Netmask  string `json:"netmask"`
	Hex      string `json:"hex"`
	Wildcard string `json:"wildcard"`
	Usable   int    `json:"usable"`
}

func (m *MaskResult) Mode() string { return "mask" }

// NetworkResult holds IP+mask output.
type NetworkResult struct {
	Address   string `json:"address"`
	CIDR      int    `json:"cidr"`
	Netmask   string `json:"netmask"`
	Hex       string `json:"hex"`
	Wildcard  string `json:"wildcard"`
	Network   string `json:"network"`
	Broadcast string `json:"broadcast"`
	First     string `json:"first"`
	Last      string `json:"last"`
	Usable    int    `json:"usable"`
}

func (n *NetworkResult) Mode() string { return "network" }

// ParseInput parses a mask or IP/mask string and returns a Result.
func ParseInput(input string) (Result, error) {
	if !inputAllowlist.MatchString(input) {
		return nil, ErrInvalidInput
	}

	// IPv6 detection: any colon means IPv6
	if strings.Contains(input, ":") {
		return parseIPv6(input)
	}

	parts := strings.SplitN(input, "/", 2)
	if len(parts) == 2 {
		if parts[0] == "" {
			// "/24" style — mask-only with CIDR prefix
			return parseMask(parts[1])
		}
		return parseNetwork(parts[0], parts[1])
	}

	// Bare numbers (CIDR without /) are no longer accepted;
	// use /24 syntax instead. Dotted quads and hex masks still work.
	if _, err := strconv.Atoi(input); err == nil {
		return nil, ErrInvalidInput
	}

	return parseMask(input)
}

func parseMask(s string) (*MaskResult, error) {
	maskBits, err := parseMaskValue(s)
	if err != nil {
		return nil, err
	}
	return maskResult(maskBits)
}

func parseNetwork(ipStr, maskStr string) (*NetworkResult, error) {
	ip, err := parseIPv4(ipStr)
	if err != nil {
		return nil, ErrInvalidInput
	}
	maskBits, err := parseMaskValue(maskStr)
	if err != nil {
		return nil, err
	}

	cidr := bits.OnesCount32(maskBits)
	wildcard := maskBits ^ 0xFFFFFFFF
	network := ip & maskBits
	broadcast := network | wildcard

	first := network + 1
	last := broadcast - 1
	if cidr >= 31 {
		first = ip
		last = ip
	}

	return &NetworkResult{
		Address:   uint32ToIP(ip),
		CIDR:      cidr,
		Netmask:   uint32ToIP(maskBits),
		Hex:       uint32ToHex(maskBits),
		Wildcard:  uint32ToIP(wildcard),
		Network:   uint32ToIP(network),
		Broadcast: uint32ToIP(broadcast),
		First:     uint32ToIP(first),
		Last:      uint32ToIP(last),
		Usable:    usableHosts(cidr),
	}, nil
}

func parseMaskValue(s string) (uint32, error) {
	// Hex mask: 0xffffff00
	if hexMask.MatchString(s) {
		v, err := strconv.ParseUint(s[2:], 16, 32)
		if err != nil {
			return 0, ErrInvalidInput
		}
		mask := uint32(v)
		if !isValidMask(mask) {
			return 0, ErrInvalidInput
		}
		return mask, nil
	}

	// Dotted quad: could be netmask or wildcard.
	if dottedQuad.MatchString(s) {
		v, err := parseIPv4(s)
		if err != nil {
			return 0, ErrInvalidInput
		}
		inverted := v ^ 0xFFFFFFFF
		// If only the wildcard interpretation is valid (not a valid netmask itself),
		// treat as wildcard. If both or only netmask is valid, treat as netmask.
		// Special case: 0.0.0.0 is valid both ways; prefer wildcard (/32) over /0.
		if v == 0 {
			// 0.0.0.0 as wildcard → mask = 0xFFFFFFFF (/32)
			return inverted, nil
		}
		if isValidMask(v) {
			return v, nil
		}
		if isValidMask(inverted) {
			return inverted, nil
		}
		return 0, ErrInvalidInput
	}

	// CIDR notation: 0-32
	cidr, err := strconv.Atoi(s)
	if err != nil || cidr < 0 || cidr > 32 {
		return 0, ErrInvalidInput
	}
	if cidr == 0 {
		return 0, nil
	}
	return uint32(0xFFFFFFFF) << (32 - cidr), nil
}

func parseIPv4(s string) (uint32, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return 0, ErrInvalidInput
	}
	var result uint32
	for _, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 || v > 255 {
			return 0, ErrInvalidInput
		}
		result = (result << 8) | uint32(v)
	}
	return result, nil
}

func isValidMask(v uint32) bool {
	// A valid netmask is all 1s followed by all 0s.
	// Inverting gives all 0s followed by all 1s.
	// Adding 1 to that should be a power of 2 (single bit set), or zero (for /32).
	inv := ^v
	return (inv & (inv + 1)) == 0
}

func uint32ToIP(v uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(v>>24)&0xFF, (v>>16)&0xFF, (v>>8)&0xFF, v&0xFF)
}

func uint32ToHex(v uint32) string {
	return fmt.Sprintf("0x%08x", v)
}

func maskResult(maskBits uint32) (*MaskResult, error) {
	cidr := bits.OnesCount32(maskBits)
	wildcard := maskBits ^ 0xFFFFFFFF
	return &MaskResult{
		CIDR:     cidr,
		Netmask:  uint32ToIP(maskBits),
		Hex:      uint32ToHex(maskBits),
		Wildcard: uint32ToIP(wildcard),
		Usable:   usableHosts(cidr),
	}, nil
}

func usableHosts(cidr int) int {
	switch {
	case cidr <= 1:
		return 1
	case cidr <= 30:
		return (1 << (32 - cidr)) - 2
	default:
		return 1
	}
}
