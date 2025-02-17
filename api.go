package efashevdsapigo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type client struct {
	apiKey                string
	apiSecret             string
	accessToken           string
	refreshToken          string
	accessTokenExpiresAt  time.Time
	refreshTokenExpiresAt time.Time

	baseURL         *url.URL
	client          *http.Client
	debugger        Debugger
	autoUpdateToken bool
}

func NewClient(ctx context.Context, apiKey, apiSecret string, opts ...Option) (Client, error) {

	if apiKey == "" {
		return nil, ValidationError("api key not provided")
	}
	if apiSecret == "" {
		return nil, ValidationError("api secret not provided")
	}

	c := &client{
		apiKey:          apiKey,
		apiSecret:       apiSecret,
		autoUpdateToken: true,
		client:          http.DefaultClient,
	}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case baseUrlOption:
			c.baseURL = opt.v
		case disableAutoUpdatingTokenOption:
			c.autoUpdateToken = !bool(opt)
		case customClientOption:
			c.client = opt.v
		case debugOption:
			c.debugger = opt.v
		}
	}

	if c.baseURL == nil {
		u, err := url.Parse(APIV2BaseURL)
		if err != nil {
			return nil, err
		}
		c.baseURL = u
	}

	err := c.InitAuth(ctx)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *client) InitAuth(ctx context.Context, opts ...Option) error {

	if c.accessTokenExpiresAt.After(time.Now()) {
		return nil
	}
	c.debug("[efashevdsapigo] access token expired.")

	if c.refreshTokenExpiresAt.After(time.Now()) {
		res, err := c.RefreshToken(ctx, opts...)
		if err != nil {
			return err
		}
		c.accessToken = res.Data.AccessToken
		c.accessTokenExpiresAt, err = parseTokenTstamp(res.Data.AccessToken)
		if err != nil {
			return err
		}
		c.debug("[efashevdsapigo] access token renewed with refresh token.")
		return nil
	}

	c.debug("[efashevdsapigo] refresh token has expired.")
	res, err := c.Auth(ctx, opts...)
	if err != nil {
		return err
	}
	t, err := parseTokenTstamp(res.Data.AccessToken)
	if err != nil {
		return err
	}
	c.accessToken = res.Data.AccessToken
	c.accessTokenExpiresAt = t
	t, err = parseTokenTstamp(res.Data.RefreshToken)
	if err != nil {
		return err
	}
	c.refreshToken = res.Data.RefreshToken
	c.refreshTokenExpiresAt = t
	c.debug("[efashevdsapigo] fresh authentication was successful.")
	return nil
}

func (c *client) Status(ctx context.Context, opts ...Option) (*StatusResp, error) {

	cl, req, err := c.setRequestParams(ctx, nil, http.MethodGet, "/status", false, opts...)
	if err != nil {
		return nil, err
	}

	var res StatusResp
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK:
		return &res, nil
	case http.StatusBadGateway:
		return nil, ErrAPIDown
	default:
		return nil, errors.New(status)
	}
}

func (c *client) ValidateSession(ctx context.Context, opts ...Option) (bool, error) {

	cl, req, err := c.setRequestParams(ctx, nil, http.MethodGet, "/validate/session", true, opts...)
	if err != nil {
		return false, err
	}

	var res struct {
		Msg string `json:"string"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return false, err
	}
	switch statusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized:
		c.debug("[efashevdsapigo] /validate/session", "status", status, "message", res.Msg)
		return false, nil
	default:
		c.debug("[efashevdsapigo] /validate/session", "status", status, "message", res.Msg)
		return false, errors.New(res.Msg)
	}
}

func (c *client) Auth(ctx context.Context, opts ...Option) (*AuthResp, error) {

	body := strings.NewReader(fmt.Sprintf(`{"api_key": %q, "api_secret": %q}`, c.apiKey, c.apiSecret))
	cl, req, err := c.setRequestParams(ctx, body, http.MethodPost, "/auth", false, opts...)
	if err != nil {
		return nil, err
	}

	var res struct {
		AuthResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK:
		v := res.AuthResp
		return &v, nil
	case http.StatusBadRequest:
		return nil, ValidationError(res.Msg)
	case http.StatusUnauthorized:
		c.debug("[efashevdsapigo] /auth", "status", status, "message", res.Msg)
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		c.debug("[efashevdsapigo] /auth", "status", status, "message", res.Msg)
		return nil, ErrAccountBlocked
	case http.StatusNotFound:
		c.debug("[efashevdsapigo] /auth", "status", status, "message", res.Msg)
		return nil, ErrAccountNotFound
	default:
		c.debug("[efashevdsapigo] /auth", "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) RefreshToken(ctx context.Context, opts ...Option) (*RefreshTokenResp, error) {

	body := strings.NewReader(fmt.Sprintf(`{"data": {"refreshToken": %q} }`, c.refreshToken))
	cl, req, err := c.setRequestParams(ctx, body, http.MethodPost, "/refresh-token", false, opts...)
	if err != nil {
		return nil, err
	}

	var res struct {
		RefreshTokenResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK:
		v := res.RefreshTokenResp
		return &v, nil
	default:
		c.debug("[efashevdsapigo] /refresh-token", "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) Balance(ctx context.Context, opts ...Option) (*BalanceResp, error) {

	cl, req, err := c.setRequestParams(ctx, nil, http.MethodGet, "/balance", true, opts...)
	if err != nil {
		return nil, err
	}
	req.URL, _ = url.Parse(fmt.Sprintf("%s?format=list", req.URL.String()))

	var res struct {
		BalanceResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK:
		v := res.BalanceResp
		return &v, nil
	default:
		c.debug("[efashevdsapigo] /balance", "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) ListVerticals(ctx context.Context, opts ...Option) (*ListVerticalsResp, error) {

	cl, req, err := c.setRequestParams(ctx, nil, http.MethodGet, "/verticals", true, opts...)
	if err != nil {
		return nil, err
	}

	var res struct {
		ListVerticalsResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK:
		v := res.ListVerticalsResp
		return &v, nil
	case http.StatusBadRequest:
		return nil, ValidationError(res.Msg)
	case http.StatusUnauthorized:
		c.debug("[efashevdsapigo] /verticals", "status", status, "message", res.Msg)
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		c.debug("[efashevdsapigo] /verticals", "status", status, "message", res.Msg)
		return nil, ErrAccountBlocked
	case http.StatusNotFound:
		c.debug("[efashevdsapigo] /verticals", "status", status, "message", res.Msg)
		return nil, ErrAccountNotFound
	default:
		c.debug("[efashevdsapigo] /verticals", "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) VendValidate(ctx context.Context, body VendValidateBody, opts ...Option) (*VendValidateResp, error) {

	bodyRaw, _ := json.Marshal(body)
	cl, req, err := c.setRequestParams(ctx, bytes.NewReader(bodyRaw), http.MethodPost, "/vend/validate", true, opts...)
	if err != nil {
		return nil, err
	}

	var res struct {
		VendValidateResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK:
		v := res.VendValidateResp
		return &v, nil
	case http.StatusBadRequest:
		return nil, ValidationError(res.Msg)
	case http.StatusUnauthorized:
		c.debug("[efashevdsapigo] /vend/validate", "status", status, "message", res.Msg)
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		c.debug("[efashevdsapigo] /vend/validate", "status", status, "message", res.Msg)
		return nil, ErrAccountBlocked
	case http.StatusNotFound:
		c.debug("[efashevdsapigo] /vend/validate", "status", status, "message", res.Msg)
		return nil, ErrAccountNotFound
	default:
		c.debug("[efashevdsapigo] /vend/validate", "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) VendExecute(ctx context.Context, body VendExecuteBody, opts ...Option) (*VendExecuteResp, error) {

	bodyRaw, _ := json.Marshal(body)
	cl, req, err := c.setRequestParams(ctx, bytes.NewReader(bodyRaw), http.MethodPost, "/vend/execute", true, opts...)
	if err != nil {
		return nil, err
	}

	var res struct {
		VendExecuteResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK, http.StatusAccepted:
		v := res.VendExecuteResp
		return &v, nil
	case http.StatusPreconditionFailed:
		c.debug("[efashevdsapigo] /vend/execute", "status", status, "message", res.Msg)
		return nil, ErrProductOutOfStock
	case http.StatusFailedDependency:
		c.debug("[efashevdsapigo] /vend/execute", "status", status, "message", res.Msg)
		return nil, ErrInsufficientBalance
	default:
		c.debug("[efashevdsapigo] /vend/execute", "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) VendTransactionStatus(ctx context.Context, transactionId string, opts ...Option) (*VendTransactionStatusResp, error) {

	path := fmt.Sprintf("/vend/%s/status", transactionId)
	cl, req, err := c.setRequestParams(ctx, nil, http.MethodGet, path, true, opts...)
	if err != nil {
		return nil, err
	}

	var res struct {
		VendTransactionStatusResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK, http.StatusAccepted:
		v := res.VendTransactionStatusResp
		return &v, nil
	case http.StatusNotFound:
		c.debug(fmt.Sprintf("[efashevdsapigo] %s", path), "status", status, "message", res.Msg)
		return nil, ErrTransactionNotFound
	default:
		c.debug(fmt.Sprintf("[efashevdsapigo] %s", path), "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) ElectricityTokens(ctx context.Context, meterNo string, tokensCount int, opts ...Option) (*ElectricityTokenResp, error) {

	meterNo = strings.TrimSpace(meterNo)
	if meterNo == "" {
		return nil, ValidationError("meter number is required")
	}
	if tokensCount <= 0 || tokensCount > 10 {
		tokensCount = 10
	}

	cl, req, err := c.setRequestParams(ctx, nil, http.MethodGet, "/electricity/tokens", true, opts...)
	if err != nil {
		return nil, err
	}
	req.URL, _ = url.Parse(fmt.Sprintf("%s?meterNo=%s&numTokens=%d", req.URL, meterNo, tokensCount))

	var res struct {
		ElectricityTokenResp
		Msg string `json:"msg"`
	}
	statusCode, status, err := httpDo(cl, req, &res, true)
	if err != nil {
		return nil, err
	}
	switch statusCode {
	case http.StatusOK:
		v := res.ElectricityTokenResp
		return &v, nil
	default:
		c.debug(fmt.Sprintf("[efashevdsapigo] %s", req.URL.RawPath), "status", status, "message", res.Msg)
		return nil, errors.New(res.Msg)
	}
}

func (c *client) setRequestParams(ctx context.Context, body io.Reader, method, path string, shouldAuth bool, opts ...Option) (*http.Client, *http.Request, error) {

	var (
		u           = c.baseURL.JoinPath(path)
		cl          = c.client
		customHd    http.Header
		updateToken = c.autoUpdateToken
	)

	for _, opt := range opts {
		switch opt := opt.(type) {
		case urlOption:
			u = opt.v
		case customClientOption:
			cl = opt.v
		case headersOption:
			customHd = opt.v
		case disableAutoUpdatingTokenOption:
			updateToken = !bool(opt)
		}
	}

	if updateToken && shouldAuth {
		err := c.InitAuth(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, nil, err
	}
	if shouldAuth {
		addBearerToken(req.Header, c.accessToken)
	}
	setHeaders(req.Header, customHd)
	return cl, req, nil
}

func (c *client) debug(msg string, args ...any) {

	if c.debugger != nil {
		c.debugger.Debug(msg, args...)
	}
}
