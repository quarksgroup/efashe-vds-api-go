package efashevdsapigo

import (
	"net/http"
	"net/url"
)

type baseUrlOption struct {
	v *url.URL
}

func (opt baseUrlOption) value() any { return opt.v }

// Attach custom API base URL only during creation of a client.
// this includes API version, check APIV2BaseURL.
func WithBaseURLOption(u *url.URL) Option {
	return baseUrlOption{v: u}
}

type urlOption struct {
	v *url.URL
}

func (opt urlOption) value() any { return opt.v }

// Attach custom full API URL during calling certain API.
func WithURLOption(u *url.URL) Option {
	return urlOption{v: u}
}

type headersOption struct {
	v http.Header
}

func (opt headersOption) value() any { return opt.v }

// Attach custom headers to an api request.
func WithHeadersOption(headers http.Header) Option {
	return headersOption{v: headers}
}

type disableAutoUpdatingTokenOption bool

func (opt disableAutoUpdatingTokenOption) value() any { return opt }

// Disable automatically updating token once access token or refresh token has expired.
// usually once access token or refresh token has expired, token are first updated before initiating the actual request.
func WithDisableAutoUpdatingTokenOption(disableFlag bool) Option {
	return disableAutoUpdatingTokenOption(disableFlag)
}

type customClientOption struct {
	v *http.Client
}

func (opt customClientOption) value() any { return opt.v }

// custom client for an api request
func WithCustomClientOption(httpClient *http.Client) Option {
	return customClientOption{v: httpClient}
}

type debugOption struct {
	v Debugger
}

func (opt debugOption) value() any { return opt.v }

func WithDebuggerOption(debugger Debugger) Option {
	return debugOption{v: debugger}
}
