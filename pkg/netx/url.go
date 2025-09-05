package netx

import (
	"net/url"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

var (
	errInvalidURL = errors.New("invalid URL format")
)

// URI represents a parsed non-standard URI that supports multiple addresses
// and query parameters.
//
// The URI struct supports the following format:
// schema://[user:password@]addr_list/path[?query]
//
// Where:
// - schema: Protocol scheme (e.g., http, https)
// - user: Optional username for authentication
// - password: Optional password for authentication
// - addr_list: Comma-separated list of addresses (may contain multiple)
// - path: URL path (unescaped)
// - query: Optional query string (unescaped)
type URI struct {
	Scheme    string        // Protocol scheme (e.g., http, https)
	User      *url.Userinfo // User information with username and password
	Addresses []string      // List of addresses (may contain multiple comma-separated addresses)
	Path      string        // URL path (unescaped)
	RawPath   string        // Raw URL path (escaped)
	RawQuery  string        // Raw query string
}

// Parse parses a non-standard URI format: schema://[user:password@]addr_list/path
// Supports URL-encoded components for user, password, and path
func ParseURI(rawURL string) (*URI, error) {
	result := &URI{}
	if rawURL == "" {
		return nil, errors.WithMessagef(errInvalidURL, "empty URL")
	}

	parts := strings.SplitN(rawURL, "://", 2)
	if len(parts) != 2 {
		return nil, errors.WithMessagef(errInvalidURL, "missing scheme")
	}
	result.Scheme = parts[0]
	remainder := parts[1]

	// Process user authentication info (if exists)
	var addrListStart int
	if userInfoEnd := strings.Index(remainder, "@"); userInfoEnd != -1 {
		userInfo := remainder[:userInfoEnd]
		addrListStart = userInfoEnd + 1

		var err error
		// Split username and password (handle encoded components)
		if passStart := strings.Index(userInfo, ":"); passStart != -1 {
			username := userInfo[:passStart]
			password := userInfo[passStart+1:]

			// Unescape username and password if they are encoded
			if username, err = url.PathUnescape(username); err != nil {
				return nil, errors.WithMessagef(errInvalidURL, "invalid username encoding: %v", err)
			}
			if password, err = url.PathUnescape(password); err != nil {
				return nil, errors.WithMessagef(errInvalidURL, "invalid password encoding: %v", err)
			}
			result.User = url.UserPassword(username, password)
		} else {
			username := userInfo
			if username, err = url.PathUnescape(username); err != nil {
				return nil, errors.WithMessagef(errInvalidURL, "invalid username encoding: %v", err)
			}
			result.User = url.User(username)
		}
	} else {
		addrListStart = 0
	}

	// Find the start of path
	pathStart := strings.IndexAny(remainder[addrListStart:], "/")
	if pathStart == -1 {
		// No path part
		addrList := remainder[addrListStart:]
		result.Addresses = parseAddressList(addrList)
		result.Path = "/"
		result.RawPath = "/"
	} else {
		// Has path part
		pathStart += addrListStart
		addrList := remainder[addrListStart:pathStart]
		result.Addresses = parseAddressList(addrList)

		// Process path and query parameters
		pathPart := remainder[pathStart:]
		queryStart := strings.Index(pathPart, "?")
		if queryStart == -1 {
			// No query parameters
			result.RawPath = pathPart
			// Unescape path if it's encoded
			if unescapedPath, err := url.PathUnescape(pathPart); err == nil {
				result.Path = unescapedPath
			} else {
				return nil, errors.WithMessagef(errInvalidURL, "invalid path encoding: %v", err)
			}
		} else {
			// With query parameters
			result.RawPath = pathPart[:queryStart]
			result.RawQuery = pathPart[queryStart+1:]

			// Unescape path if it's encoded
			if unescapedPath, err := url.PathUnescape(result.RawPath); err == nil {
				result.Path = unescapedPath
			} else {
				return nil, errors.WithMessagef(errInvalidURL, "invalid path encoding: %v", err)
			}
		}
	}

	if len(result.Addresses) == 0 || slices.Contains(result.Addresses, "") {
		return nil, errors.WithMessagef(errInvalidURL, "find empty address in URI")
	}

	return result, nil
}

// parseAddressList splits and trims comma-separated address list
func parseAddressList(addrList string) []string {
	addresses := strings.Split(addrList, ",")
	for i, addr := range addresses {
		addresses[i] = strings.TrimSpace(addr)
	}
	return addresses
}

// String reassembles the URI into a valid URL string
func (u *URI) String() string {
	var result strings.Builder
	result.WriteString(u.Scheme)
	result.WriteString("://")

	if u.User != nil {
		result.WriteString(u.User.String())
		result.WriteString("@")
	}

	result.WriteString(strings.Join(u.Addresses, ","))
	result.WriteString(u.RawPath)

	if u.RawQuery != "" {
		result.WriteString("?")
		result.WriteString(u.RawQuery)
	}

	return result.String()
}
