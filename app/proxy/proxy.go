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

type Candidate struct {
	Address       string
	Host          string
	Port          int
	SourceCountry string
	SourceCity    string
	Raw           string
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

func ExtractCandidates(raw string) ([]Candidate, []error) {
	if IsSkippableLine(raw) {
		return nil, nil
	}
	fields := candidateFields(raw)
	if len(fields) == 0 {
		return nil, []error{errors.New("proxy line is empty")}
	}
	candidates := make([]Candidate, 0, len(fields))
	errs := make([]error, 0)
	for _, field := range fields {
		candidate, err := ParseCandidate(field)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		candidates = append(candidates, candidate)
	}
	return candidates, errs
}

func ParseCandidate(raw string) (Candidate, error) {
	token, sourceCountry, sourceCity := splitCandidateMetadata(raw)
	parsed, err := Parse(token)
	if err != nil {
		return Candidate{}, err
	}
	return Candidate{
		Address:       parsed.String(),
		Host:          parsed.Host,
		Port:          parsed.Port,
		SourceCountry: sourceCountry,
		SourceCity:    sourceCity,
		Raw:           strings.TrimSpace(raw),
	}, nil
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

func candidateFields(raw string) []string {
	if IsSkippableLine(raw) {
		return nil
	}
	trimmed := strings.TrimSpace(strings.TrimPrefix(raw, "\uFEFF"))
	return strings.Fields(trimmed)
}

// IsSkippableLine reports whether a fetched source line is structural text,
// not a proxy candidate.
func IsSkippableLine(raw string) bool {
	trimmed := strings.TrimSpace(strings.TrimPrefix(raw, "\uFEFF"))
	return trimmed == "" || strings.HasPrefix(trimmed, "#")
}

func splitCandidateMetadata(raw string) (token string, sourceCountry string, sourceCity string) {
	parts := strings.Split(strings.TrimSpace(strings.TrimRight(raw, ";")), ",")
	if len(parts) > 0 {
		token = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		sourceCountry = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		sourceCity = strings.TrimSpace(parts[2])
	}
	return token, sourceCountry, sourceCity
}

func splitHostPort(token string) (string, string, error) {
	if strings.Contains(token, "://") {
		parsed, err := url.Parse(token)
		if err != nil {
			return "", "", fmt.Errorf("invalid proxy URL %q: %w", token, err)
		}
		if parsed.Scheme != "socks5" {
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
