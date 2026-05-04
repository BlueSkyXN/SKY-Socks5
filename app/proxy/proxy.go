package proxy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type Proxy struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func Parse(raw string) (Proxy, error) {
	token := firstToken(raw)
	if token == "" {
		return Proxy{}, errors.New("proxy line is empty")
	}
	if strings.HasPrefix(token, "#") {
		return Proxy{}, errors.New("proxy line is a comment")
	}

	host, portText, err := splitHostPort(token)
	if err != nil {
		return Proxy{}, err
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid port %q: %w", portText, err)
	}
	parsed := Proxy{
		Host: strings.ToLower(strings.TrimSpace(host)),
		Port: port,
	}
	if err := parsed.Validate(); err != nil {
		return Proxy{}, err
	}
	return parsed, nil
}

func Normalize(raw string) (string, error) {
	parsed, err := Parse(raw)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func (p Proxy) String() string {
	return net.JoinHostPort(p.Host, strconv.Itoa(p.Port))
}

func (p Proxy) Validate() error {
	if strings.TrimSpace(p.Host) == "" {
		return errors.New("host is required")
	}
	if p.Port < 1 || p.Port > 65535 {
		return fmt.Errorf("port %d is out of range", p.Port)
	}
	return nil
}

func SortStrings(addresses []string) {
	sort.Slice(addresses, func(i, j int) bool {
		leftHost, leftPort := sortParts(addresses[i])
		rightHost, rightPort := sortParts(addresses[j])
		if leftHost == rightHost {
			return leftPort < rightPort
		}
		return leftHost < rightHost
	})
}

func firstToken(raw string) string {
	trimmed := strings.TrimSpace(strings.TrimPrefix(raw, "\uFEFF"))
	if trimmed == "" {
		return ""
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return ""
	}
	return strings.TrimRight(fields[0], ",;")
}

func splitHostPort(token string) (string, string, error) {
	if strings.Contains(token, "://") {
		parsed, err := url.Parse(token)
		if err != nil {
			return "", "", fmt.Errorf("invalid proxy URL %q: %w", token, err)
		}
		if parsed.Scheme != "socks5" && parsed.Scheme != "http" && parsed.Scheme != "https" {
			return "", "", fmt.Errorf("unsupported proxy scheme %q", parsed.Scheme)
		}
		host := parsed.Hostname()
		port := parsed.Port()
		if host == "" || port == "" {
			return "", "", fmt.Errorf("proxy %q must include host and port", token)
		}
		return host, port, nil
	}
	if host, port, err := net.SplitHostPort(token); err == nil {
		return host, port, nil
	}
	if strings.Count(token, ":") > 1 && !strings.HasPrefix(token, "[") {
		return "", "", fmt.Errorf("IPv6 proxy %q must use bracket notation", token)
	}
	lastColon := strings.LastIndex(token, ":")
	if lastColon <= 0 || lastColon == len(token)-1 {
		return "", "", fmt.Errorf("proxy %q must be in host:port format", token)
	}
	return token[:lastColon], token[lastColon+1:], nil
}

func sortParts(address string) (string, int) {
	parsed, err := Parse(address)
	if err != nil {
		return strings.ToLower(address), 0
	}
	return parsed.Host, parsed.Port
}
