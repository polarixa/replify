package sysx

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// localCIDRs holds the parsed private and unique-local IP ranges used by
// IsLocalIP. Parsing is performed exactly once on first use via getLocalCIDRs.
var (
	_localCIDRsOnce sync.Once
	_localCIDRs     []*net.IPNet
)

// IsIPv4 reports whether the given string is a valid IPv4 address.
//
// The check uses net.ParseIP: the string must be a dotted-decimal notation
// such as "192.168.1.1". IPv4-in-IPv6 forms (e.g. "::ffff:192.168.1.1") are
// NOT recognized as IPv4 by this function.
//
// Parameters:
//   - `ip`: the IP address string to validate.
//
// Returns:
//
//	A boolean value:
//	 - true  when ip is a valid, pure IPv4 address;
//	 - false otherwise.
//
// Example:
//
//	sysx.IsIPv4("192.168.1.1") // true
//	sysx.IsIPv4("::1")         // false
func IsIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil && strings.Contains(ip, ".")
}

// IsIPv6 reports whether the given string is a valid IPv6 address.
//
// Addresses that are valid IPv4 dotted-decimal strings (e.g. "192.168.1.1")
// are NOT considered IPv6 by this function, even though Go internally
// represents IPv4 addresses using a 16-byte form.
//
// Parameters:
//   - `ip`: the IP address string to validate.
//
// Returns:
//
//	A boolean value:
//	 - true  when ip is a valid IPv6 address (and not a pure IPv4 string);
//	 - false otherwise.
//
// Example:
//
//	sysx.IsIPv6("::1")           // true
//	sysx.IsIPv6("2001:db8::1")  // true
//	sysx.IsIPv6("192.168.1.1")  // false
func IsIPv6(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && !strings.Contains(ip, ".")
}

// IsLocalIP reports whether the given IP address is a loopback or
// private (RFC 1918 / RFC 4193) address.
//
// Recognized local ranges:
//   - IPv4 loopback:   127.0.0.0/8
//   - IPv4 link-local: 169.254.0.0/16
//   - IPv4 private:    10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
//   - IPv6 loopback:   ::1
//   - IPv6 link-local: fe80::/10
//   - IPv6 unique-local: fc00::/7
//
// Parameters:
//   - `ip`: the IP address string to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when the address belongs to a local/private range;
//	 - false when the address is public or cannot be parsed.
//
// Example:
//
//	sysx.IsLocalIP("127.0.0.1")    // true
//	sysx.IsLocalIP("10.0.0.1")     // true
//	sysx.IsLocalIP("8.8.8.8")      // false
func IsLocalIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// Loopback covers 127.0.0.1/8 and ::1
	if parsed.IsLoopback() {
		return true
	}
	// Link-local (169.254.x.x / fe80::/10)
	if parsed.IsLinkLocalUnicast() {
		return true
	}

	// Private IPv4 ranges (RFC 1918) and IPv6 unique-local (fc00::/7)
	for _, cidr := range getLocalCIDRs() {
		if cidr.Contains(parsed) {
			return true
		}
	}
	return false
}

// IsPortOpen reports whether the given TCP port is reachable on host.
//
// The check attempts a TCP connection with a 3-second timeout. The function
// returns true only when the connection is accepted; no data is exchanged.
//
// Parameters:
//   - `host`: the host name or IP address to connect to.
//   - `port`: the TCP port number (1-65535).
//
// Returns:
//
//	A boolean value:
//	 - true  when a TCP connection to host:port succeeds within 3 seconds;
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsPortOpen("localhost", 5432) {
//	    fmt.Println("PostgreSQL is reachable")
//	}
func IsPortOpen(host string, port int) bool {
	return CheckTCPConn(host, port, 3*time.Second) == nil
}

// IsPortAvailable reports whether the given TCP port is available for binding
// on the local machine (i.e. nothing is currently listening on that port).
//
// The check attempts to open a TCP listener on 0.0.0.0:<port>. If the listen
// succeeds the port is free and the listener is immediately closed.
//
// Parameters:
//   - `port`: the TCP port number to test (1-65535).
//
// Returns:
//
//	A boolean value:
//	 - true  when the port is not in use and can be bound;
//	 - false when the port is already in use, the OS rejects the bind, or
//	         the port number is outside the valid range 1-65535.
//
// Example:
//
//	if sysx.IsPortAvailable(8080) {
//	    fmt.Println("port 8080 is free")
//	}
func IsPortAvailable(port int) bool {
	// Security fix (CWE-20): validate port range before syscall.
	// port=0 triggers OS ephemeral-port assignment (always succeeds, misleading);
	// negative values produce a syscall error that could mask the real intent.
	if port < 1 || port > 65535 {
		return false
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	defer ln.Close()
	return true
}

// GetLocalIP returns the first non-loopback, non-link-local IPv4 address
// found on the machine's network interfaces.
//
// The function iterates through all network interfaces in the order returned
// by net.Interfaces() and returns the first eligible address. An error is
// returned when no suitable address is found or interface enumeration fails.
//
// Returns:
//
//	(string, error): the IPv4 address string and nil on success, or an empty
//	string and a non-nil error when no local IP can be determined.
//
// Example:
//
//	ip, err := sysx.GetLocalIP()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("local IP:", ip)
func GetLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("sysx: failed to list network interfaces: %w", err)
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ipStr string
			switch v := addr.(type) {
			case *net.IPNet:
				ipStr = v.IP.String()
			case *net.IPAddr:
				ipStr = v.IP.String()
			}
			if IsIPv4(ipStr) && !net.ParseIP(ipStr).IsLoopback() && !net.ParseIP(ipStr).IsLinkLocalUnicast() {
				return ipStr, nil
			}
		}
	}
	return "", errors.New("sysx: no suitable local IPv4 address found")
}

// GetPublicIP retrieves the public (externally visible) IP address of the
// machine by querying the well-known plain-text endpoint https://api.ipify.org.
//
// A 10-second timeout is applied to the HTTP request. The function requires
// outbound internet access; it will fail in air-gapped environments.
//
// Returns:
//
//	(string, error): the public IP address string and nil on success, or an
//	empty string and a non-nil error when the request fails or times out.
//
// Example:
//
//	ip, err := sysx.GetPublicIP()
//	if err != nil {
//	    log.Printf("cannot determine public IP: %v", err)
//	} else {
//	    fmt.Println("public IP:", ip)
//	}
func GetPublicIP() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "", fmt.Errorf("sysx: public IP request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sysx: public IP endpoint returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return "", fmt.Errorf("sysx: failed to read public IP response: %w", err)
	}
	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("sysx: public IP response is not a valid IP address: %q", ip)
	}
	return ip, nil
}

// GetInterfaceIPs returns all unicast IP addresses assigned to the machine's
// network interfaces, including both IPv4 and IPv6 addresses.
//
// Loopback and down interfaces are excluded. The returned strings are in
// their canonical form (e.g. "192.168.1.5" or "fe80::1").
//
// Returns:
//
//	([]string, error): a slice of IP address strings and nil on success, or
//	nil and a non-nil error when interface enumeration fails.
//
// Example:
//
//	ips, err := sysx.GetInterfaceIPs()
//	for _, ip := range ips {
//	    fmt.Println(ip)
//	}
func GetInterfaceIPs() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("sysx: failed to list network interfaces: %w", err)
	}
	var result []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				result = append(result, v.IP.String())
			case *net.IPAddr:
				result = append(result, v.IP.String())
			}
		}
	}
	return result, nil
}

// IsValidHost reports whether the given string is a resolvable host name or a
// valid IP address.
//
// The function first tries to parse the input as an IP address. If that fails,
// it attempts a DNS lookup. It returns true if either check succeeds.
//
// Parameters:
//   - `host`: the host name or IP address to validate.
//
// Returns:
//
//	A boolean value:
//	 - true  when the host is a valid IP or can be resolved via DNS;
//	 - false otherwise.
//
// Example:
//
//	sysx.IsValidHost("localhost")   // true (resolves to 127.0.0.1)
//	sysx.IsValidHost("8.8.8.8")    // true (valid IP)
//	sysx.IsValidHost("invalid..x") // false
func IsValidHost(host string) bool {
	if strutil.IsEmpty(host) {
		return false
	}
	if net.ParseIP(host) != nil {
		return true
	}
	_, err := net.LookupHost(host)
	return err == nil
}

// IsValidURL reports whether the given string is a syntactically valid URL
// with a non-empty scheme and host component.
//
// The function uses url.Parse for parsing. It does not perform DNS resolution
// or check whether the URL is reachable.
//
// Parameters:
//   - `rawURL`: the URL string to validate.
//
// Returns:
//
//	A boolean value:
//	 - true  when rawURL has a valid scheme and non-empty host;
//	 - false otherwise.
//
// Example:
//
//	sysx.IsValidURL("https://example.com/path") // true
//	sysx.IsValidURL("not-a-url")                // false
func IsValidURL(rawURL string) bool {
	if strutil.IsEmpty(rawURL) {
		return false
	}
	u, err := url.Parse(rawURL)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// ParseHostPort splits a network address of the form "host:port" into its
// constituent parts, returning the host string and port integer.
//
// The function delegates to net.SplitHostPort and additionally parses the
// port string to an integer. IPv6 addresses must be enclosed in square
// brackets (e.g. "[::1]:8080").
//
// Parameters:
//   - `addr`: the "host:port" address string to parse.
//
// Returns:
//
//	(host string, port int, err error): the host name, port number, and nil
//	on success, or empty values and a non-nil error when the address is
//	malformed.
//
// Example:
//
//	host, port, err := sysx.ParseHostPort("localhost:8080")
//	// host == "localhost", port == 8080, err == nil
//
//	host, port, err = sysx.ParseHostPort("[::1]:443")
//	// host == "::1", port == 443, err == nil
func ParseHostPort(addr string) (host string, port int, err error) {
	h, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("sysx: ParseHostPort(%q): %w", addr, err)
	}
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("sysx: ParseHostPort(%q): port is not an integer: %w", addr, err)
	}
	return h, p, nil
}

// PingHost checks whether the given host is reachable by attempting a TCP
// connection to port 80 with a 5-second timeout.
//
// This is a connectivity probe suitable for production backend services. It
// does not send ICMP packets (which require elevated privileges on most
// systems); instead it verifies TCP-layer reachability to the standard HTTP
// port.
//
// For probing arbitrary ports, use CheckTCPConn.
//
// Parameters:
//   - `host`: the host name or IP address to probe.
//
// Returns:
//
//	An error if the host cannot be reached via TCP port 80 within 5 seconds;
//	nil when the connection succeeds.
//
// Example:
//
//	if err := sysx.PingHost("google.com"); err != nil {
//	    fmt.Println("host unreachable:", err)
//	}
func PingHost(host string) error {
	return CheckTCPConn(host, 80, 5*time.Second)
}

// CheckTCPConn verifies that a TCP connection can be established to
// host:port within the specified timeout duration.
//
// The connection is closed immediately after being established; no data is
// exchanged. An error is returned on connection failure, timeout, or when
// an invalid port is provided.
//
// Parameters:
//   - `host`:    the host name or IP address to connect to.
//   - `port`:    the TCP port number (1-65535).
//   - `timeout`: the maximum duration to wait for the connection.
//
// Returns:
//
//	An error if the connection could not be established within timeout; nil on success.
//
// Example:
//
//	err := sysx.CheckTCPConn("db.internal", 5432, 3*time.Second)
//	if err != nil {
//	    log.Printf("database unreachable: %v", err)
//	}
func CheckTCPConn(host string, port int, timeout time.Duration) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("sysx: CheckTCPConn: port %d is out of range (1-65535)", port)
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Errorf("sysx: TCP connection to %s failed: %w", addr, err)
	}
	defer conn.Close()
	return nil
}

// getLocalCIDRs returns the package-level private/unique-local CIDR list,
// initializing it on first call via sync.Once. The returned slice must not be
// modified by callers.
// Panics if any hardcoded CIDR block fails to parse, which would indicate a
// programming error.
func getLocalCIDRs() []*net.IPNet {
	_localCIDRsOnce.Do(func() {
		blocks := []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"fc00::/7",
		}
		nets := make([]*net.IPNet, 0, len(blocks))
		for _, b := range blocks {
			_, ipNet, err := net.ParseCIDR(b)
			if err != nil {
				panic("sysx: failed to parse hardcoded CIDR block " + b + ": " + err.Error())
			}
			nets = append(nets, ipNet)
		}
		_localCIDRs = nets
	})
	return _localCIDRs
}
