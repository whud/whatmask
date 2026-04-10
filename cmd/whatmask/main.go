// Whatmask web service — subnet calculator.
//
// Original C implementation by Joe Laffey (laffeycomputer.com/whatmask.html).
// Ruby rewrite by Joe Topjian (github.com/geezyx/whatmask).
// Go rewrite for web service.
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/whud/whatmask/internal/whatmask"
)

//go:embed static
var staticFiles embed.FS

func main() {
	serve := flag.Bool("serve", false, "start web server instead of CLI mode")
	port := flag.String("port", "", "port for web server (default: 8080, or PORT env)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: whatmask [options] [input]\n\n")
		fmt.Fprintf(os.Stderr, "Subnet calculator. Pass an input for CLI mode, or --serve for web server.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  whatmask /24                    Mask lookup\n")
		fmt.Fprintf(os.Stderr, "  whatmask 192.168.1.0/24         Network calculation\n")
		fmt.Fprintf(os.Stderr, "  whatmask 255.255.255.0          Mask from dotted quad\n")
		fmt.Fprintf(os.Stderr, "  whatmask 0xffffff00             Mask from hex\n")
		fmt.Fprintf(os.Stderr, "  whatmask 2001:db8::/32          IPv6 network\n")
		fmt.Fprintf(os.Stderr, "  whatmask --serve                Start web server\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *serve {
		startServer(*port)
		return
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	input := strings.Join(flag.Args(), "")
	result, err := whatmask.ParseInput(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printResult(result)
}

func printResult(result whatmask.Result) {
	switch v := result.(type) {
	case *whatmask.MaskResult:
		fmt.Printf("CIDR:             /%d\n", v.CIDR)
		fmt.Printf("Netmask:          %s\n", v.Netmask)
		fmt.Printf("Hex:              %s\n", v.Hex)
		fmt.Printf("Wildcard:         %s\n", v.Wildcard)
		fmt.Printf("Usable hosts:     %d\n", v.Usable)
	case *whatmask.NetworkResult:
		fmt.Printf("Address:          %s\n", v.Address)
		fmt.Printf("CIDR:             /%d\n", v.CIDR)
		fmt.Printf("Netmask:          %s\n", v.Netmask)
		fmt.Printf("Hex:              %s\n", v.Hex)
		fmt.Printf("Wildcard:         %s\n", v.Wildcard)
		fmt.Printf("Network:          %s\n", v.Network)
		fmt.Printf("Broadcast:        %s\n", v.Broadcast)
		fmt.Printf("First usable:     %s\n", v.First)
		fmt.Printf("Last usable:      %s\n", v.Last)
		fmt.Printf("Usable hosts:     %d\n", v.Usable)
	case *whatmask.IPv6Result:
		fmt.Printf("Address:          %s\n", v.Address)
		fmt.Printf("Address (full):   %s\n", v.AddressFull)
		fmt.Printf("CIDR:             /%d\n", v.CIDR)
		fmt.Printf("Network:          %s\n", v.Network)
		fmt.Printf("Network (full):   %s\n", v.NetworkFull)
		fmt.Printf("Last address:     %s\n", v.Last)
		fmt.Printf("Last (full):      %s\n", v.LastFull)
		fmt.Printf("Total addresses:  %s\n", v.Total)
		fmt.Printf("Type:             %s\n", v.Type)
	}
}

func startServer(portFlag string) {
	p := portFlag
	if p == "" {
		p = os.Getenv("PORT")
	}
	if p == "" {
		p = "8080"
	}

	mux := http.NewServeMux()

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/api/calc", handleCalc)

	handler := securityHeaders(rateLimiter(mux))

	srv := &http.Server{
		Addr:              ":" + p,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}

	log.Printf("whatmask listening on :%s", p)
	log.Fatal(srv.ListenAndServe())
}

func handleCalc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	input := r.URL.Query().Get("input")
	if input == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input"})
		return
	}

	result, err := whatmask.ParseInput(input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input"})
		return
	}

	switch v := result.(type) {
	case *whatmask.MaskResult:
		writeJSON(w, http.StatusOK, struct {
			Mode string `json:"mode"`
			*whatmask.MaskResult
		}{"mask", v})
	case *whatmask.NetworkResult:
		writeJSON(w, http.StatusOK, struct {
			Mode string `json:"mode"`
			*whatmask.NetworkResult
		}{"network", v})
	case *whatmask.IPv6Result:
		writeJSON(w, http.StatusOK, struct {
			Mode string `json:"mode"`
			*whatmask.IPv6Result
		}{"network6", v})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		next.ServeHTTP(w, r)
	})
}

const maxVisitors = 10000

type visitor struct {
	count   int
	resetAt time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

// clientIP extracts the real client IP, preferring Cf-Connecting-Ip
// (set by Cloudflare) and falling back to RemoteAddr with port stripped.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("Cf-Connecting-Ip"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func rateLimiter(next http.Handler) http.Handler {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			now := time.Now()
			for ip, v := range visitors {
				if now.After(v.resetAt) {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		mu.Lock()
		v, exists := visitors[ip]
		now := time.Now()
		if !exists || now.After(v.resetAt) {
			if len(visitors) >= maxVisitors && !exists {
				mu.Unlock()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			visitors[ip] = &visitor{count: 1, resetAt: now.Add(time.Minute)}
			mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		if v.count > 60 {
			mu.Unlock()
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
