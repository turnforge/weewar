package utils

import (
	"fmt"
	"net"
	"strings"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBold   = "\033[1m"
)

// makeClickableLink creates a clickable terminal link using OSC 8 escape sequences
// Supported by iTerm2, VS Code terminal, Windows Terminal, and many modern terminals
func makeClickableLink(url string, color string) string {
	// OSC 8 format: \033]8;;URL\033\\TEXT\033]8;;\033\\
	return fmt.Sprintf("\033]8;;%s\033\\%s%s%s\033]8;;\033\\", url, color, url, ColorReset)
}

// GetLocalIP returns the local network IP address (preferring ethernet/wifi over localhost)
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// Check if it's an IP network address
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// Only return IPv4 addresses
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no local network IP found")
}

// PrintStartupMessage prints a colorized startup message with clickable links
func PrintStartupMessage(address string) {
	port := address
	if strings.HasPrefix(port, ":") {
		port = port[1:]
	} else {
		// Extract port from address like "localhost:8080"
		parts := strings.Split(address, ":")
		if len(parts) > 0 {
			port = parts[len(parts)-1]
		}
	}

	fmt.Println()
	fmt.Printf("%s%s╔════════════════════════════════════════════════════════════╗%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s%s║  Server started! Open in your browser:                     ║%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s%s╠════════════════════════════════════════════════════════════╣%s\n", ColorBold, ColorCyan, ColorReset)

	// Localhost link
	localhostURL := fmt.Sprintf("http://localhost:%s", port)
	clickableLocalhost := makeClickableLink(localhostURL, ColorGreen)
	spaces := max(0, 60-len(localhostURL)) - 2
	fmt.Printf("%s%s║  %s%s%s%s║%s\n",
		ColorBold, ColorCyan, clickableLocalhost, ColorCyan, ColorBold, strings.Repeat(" ", spaces), ColorReset)

	// Local network IP link
	if localIP, err := GetLocalIP(); err == nil {
		networkURL := fmt.Sprintf("http://%s:%s", localIP, port)
		clickableNetwork := makeClickableLink(networkURL, ColorYellow)
		spaces := max(0, 60-len(networkURL)) - 2
		fmt.Printf("%s%s║  %s%s%s%s║%s\n",
			ColorBold, ColorCyan, clickableNetwork, ColorCyan, ColorBold, strings.Repeat(" ", spaces), ColorReset)
	}

	fmt.Printf("%s%s╚════════════════════════════════════════════════════════════╝%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Println()
}
