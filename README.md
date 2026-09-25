# whatmask

A subnet calculator for IPv4 and IPv6. Give it a mask or a network and it tells you everything about it:

```
$ whatmask 192.168.1.0/24
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

It runs in your terminal, or as a small web server with a browser UI and a JSON API.

**Features:**
- Accepts masks in any common format: `/24`, `255.255.255.0`, `0xffffff00` or a wildcard (`0.0.0.255`)
- IPv6 support with compressed and expanded notation
- IPv6 address type classification (Global Unicast, Link-Local, Unique Local, etc.)
- Web UI with live results as you type
- JSON REST API for programmatic use
- Single binary with no dependencies, plus a 2.5MB Docker image

## Install

**Prerequisite:** [Go](https://go.dev/doc/install) 1.24 or newer. Run `go version` to check.

### 1. Install whatmask

```bash
go install github.com/whud/whatmask/cmd/whatmask@latest
```

This downloads whatmask, builds it, and saves the program in Go's `bin` folder, which is `~/go/bin` on macOS and Linux or `%USERPROFILE%\go\bin` on Windows.

### 2. Make sure `whatmask` runs

Try it:

```bash
whatmask /24
```

If you see subnet output, you're done. 🎉

If you get `command not found` (or `'whatmask' is not recognized` on Windows), your terminal doesn't know about Go's `bin` folder yet. Add it to your `PATH` using the commands for your shell:

**macOS (zsh, the default):**

```bash
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
source ~/.zshrc
```

**Linux (bash, the default):**

```bash
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc
```

**fish:**

```fish
fish_add_path ~/go/bin
```

**Windows:** the Go installer normally sets this up for you, so open a new terminal window and try again.

Not sure which shell you use? Run `echo $SHELL`.

Then run `whatmask /24` again.

### Updating and uninstalling

Run `whatmask --version` to see which version you have. To update, run the install command again. To uninstall, delete the binary with `rm ~/go/bin/whatmask`, or delete `%USERPROFILE%\go\bin\whatmask.exe` on Windows.

## Usage

Pass a subnet mask or network as an argument:

```bash
whatmask /24                  # what does a /24 mask look like?
whatmask 255.255.255.0        # same, from a dotted netmask
whatmask 0xffffff00           # same, from hex
whatmask 192.168.1.0/24       # full details for a network
whatmask 10.0.0.1/255.0.0.0   # network with a dotted netmask
whatmask 2001:db8::1/48       # IPv6 network
```

A mask on its own (like `/24`) gives the mask details and host count. A mask with an address (like `192.168.1.0/24`) also gives the network, broadcast and usable address range.

### Supported inputs

- **IPv4 mask:** `/24`, `255.255.255.0`, `0xffffff00`, or `0.0.0.255` (wildcard)
- **IPv4 network:** `192.168.1.0/24`, `10.0.0.1/255.255.255.0`
- **IPv6 network:** `2001:db8::1/48`, `fe80::1/10`, `::1/128`

IPv4 and IPv6 are auto-detected — if the input contains a colon, it's treated as IPv6.

Run `whatmask --help` for a summary.

## Web interface

whatmask can also run as a web server, which gives you a browser UI with live results as you type:

```bash
whatmask --serve               # http://localhost:8080
whatmask --serve --port 3000   # http://localhost:3000
```

The port can also be set with the `PORT` environment variable. The default is `8080`.

### Running with Docker

To host the web interface without installing Go, build and run the Docker image from a clone of this repo:

```bash
git clone https://github.com/whud/whatmask.git
cd whatmask
docker build -t whatmask .
docker run -d --name whatmask -p 8080:8080 whatmask
```

To use a different port, change the first number: `-p 3000:8080` serves on port 3000.

To update to the latest version:

```bash
git pull
docker build -t whatmask .
docker stop whatmask && docker rm whatmask
docker run -d --name whatmask -p 8080:8080 whatmask
```

### JSON API

While the server is running, you can also query it from scripts:

```bash
curl "localhost:8080/api/calc?input=/24"
curl "localhost:8080/api/calc?input=192.168.1.0/24"
curl "localhost:8080/api/calc?input=2001:db8::1/48"
```

#### `GET /api/calc?input=<value>`

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

## Development

```bash
git clone https://github.com/whud/whatmask.git
cd whatmask
go test ./...
go build ./cmd/whatmask/
./whatmask /24
```

## History

This is a complete rewrite of the original [whatmask](http://www.laffeycomputer.com/whatmask.html) C command-line tool by Joe Laffey. 

The [Ruby rewrite](https://github.com/geezyx/whatmask) by Joe Topjian served as the starting point. This version is a single Go binary that works both as a command-line tool and as a web service.

This project was built with [Claude Code](https://claude.ai/claude-code) by Anthropic.

## License

GPL-3.0
