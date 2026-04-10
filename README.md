# whatmask

A subnet calculator with IPv4 and IPv6 support. Use it directly from the command line or run it as a web service with a browser UI and JSON API.

## About This Fork

This is a complete rewrite of the original [whatmask](http://www.laffeycomputer.com/whatmask.html) C CLI tool. The [Ruby rewrite](https://github.com/geezyx/whatmask) by Joe Topjian served as the starting point. This version is a single Go binary that works both as a CLI tool and a web service.

**What's new:**
- CLI mode for quick lookups directly in your terminal
- Web UI with live results as you type
- JSON REST API for programmatic use
- IPv6 support with compressed and expanded notation
- Address type classification (Global Unicast, Link-Local, Unique Local, etc.)
- Single binary deployment, 2.5MB Docker image

## Install

### Go Install

```bash
go install github.com/whud/whatmask/cmd/whatmask@latest
```

### From Source

Requires Go 1.24+.

```bash
go build ./cmd/whatmask/
```

### Docker (web server only)

```bash
docker build -t whatmask .
docker run -d -p 8080:8080 whatmask
```

To use a different port, change the first number: `-p 3000:8080` serves on port 3000.

To update to the latest version:

```bash
git pull
docker build -t whatmask .
docker stop whatmask && docker rm whatmask
docker run -d --name whatmask -p 8080:8080 whatmask
```

## CLI Mode (default)

Pass a subnet mask or network as an argument to get results directly in your terminal:

```bash
whatmask /24
whatmask 192.168.1.0/24
whatmask 255.255.255.0
whatmask 0xffffff00
whatmask 2001:db8::1/48
```

Example output for `whatmask 192.168.1.0/24`:

```
Address:          192.168.1.0
CIDR:             /24
Netmask:          255.255.255.0
Hex:              0xffffff00
Wildcard:         0.0.0.255
Network:          192.168.1.0
Broadcast:        192.168.1.255
First usable:     192.168.1.1
Last usable:      192.168.1.254
Usable hosts:     254
```

### Supported inputs

- **IPv4 mask:** `/24`, `255.255.255.0`, `0xffffff00`, or `0.0.0.255` (wildcard)
- **IPv4 network:** `192.168.1.0/24`, `10.0.0.1/255.255.255.0`
- **IPv6 network:** `2001:db8::1/48`, `fe80::1/10`, `::1/128`

IPv4 and IPv6 are auto-detected — if the input contains a colon, it's treated as IPv6.

## Web Server Mode

Start the web server with `--serve`:

```bash
whatmask --serve
whatmask --serve --port 3000
```

The port can also be set with the `PORT` environment variable. Default is `8080`.

Open [http://localhost:8080](http://localhost:8080) for the browser UI, or use the JSON API:

```bash
curl "localhost:8080/api/calc?input=/24"
curl "localhost:8080/api/calc?input=192.168.1.0/24"
curl "localhost:8080/api/calc?input=2001:db8::1/48"
```

## API

### `GET /api/calc?input=<value>`

**IPv4 mask-only** (e.g. `?input=/24`):

```json
{
  "mode": "mask",
  "cidr": 24,
  "netmask": "255.255.255.0",
  "hex": "0xffffff00",
  "wildcard": "0.0.0.255",
  "usable": 254
}
```

**IPv4 network** (e.g. `?input=192.168.1.100/24`):

```json
{
  "mode": "network",
  "address": "192.168.1.100",
  "cidr": 24,
  "netmask": "255.255.255.0",
  "hex": "0xffffff00",
  "wildcard": "0.0.0.255",
  "network": "192.168.1.0",
  "broadcast": "192.168.1.255",
  "first": "192.168.1.1",
  "last": "192.168.1.254",
  "usable": 254
}
```

**IPv6 network** (e.g. `?input=2001:db8::1/48`):

```json
{
  "mode": "network6",
  "address": "2001:db8::1",
  "address_full": "2001:0db8:0000:0000:0000:0000:0000:0001",
  "cidr": 48,
  "network": "2001:db8::",
  "network_full": "2001:0db8:0000:0000:0000:0000:0000:0000",
  "last": "2001:db8:0:ffff:ffff:ffff:ffff:ffff",
  "last_full": "2001:0db8:0000:ffff:ffff:ffff:ffff:ffff",
  "total": "1208925819614629174706176",
  "type": "Global Unicast"
}
```

**Error** (400):

```json
{
  "error": "invalid input"
}
```

## Attribution

- Original whatmask by Joe Laffey ([laffeycomputer.com/whatmask.html](http://www.laffeycomputer.com/whatmask.html))
- Ruby rewrite by Joe Topjian ([github.com/geezyx/whatmask](https://github.com/geezyx/whatmask))

## Built With

This project was built with [Claude Code](https://claude.ai/claude-code) by Anthropic.

## License

GPL-3.0
