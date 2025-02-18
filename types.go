package efashevdsapigo

import (
	"context"
	"errors"
)

const (
	APIV2BaseURL = "https://sb-api.efashe.com/rw/v2/"

	// Known balances id
	CommissionBalanceId = "commission"
	MainBalanceId       = "main"
	FeesbalanceId       = "fees"

	// Known verticals id
	AirtimeVerticalId     = "airtime"
	PayTvVerticalId       = "paytv"
	ElectricityVerticalId = "electricity"

	// Transaction state
	TransactionFailedState    = "failed"
	TransactionSuccessedState = "successful"
	TransactionTimeoutState   = "timedout"
	TransactionPendingState   = "pending"
	TransactionInitiatedtate  = "initiated"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrAccountBlocked      = errors.New("account blocked")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrProductOutOfStock   = errors.New("product is out of stock")
	ErrInsufficientBalance = errors.New("insufficient wallet balance")
	ErrAPIDown             = errors.New("API is down")
)

type Client interface {
	// This method is called to preset access token and verify api secret and keys.
	// It can be called multiple times to renew tokens.
	InitAuth(ctx context.Context, opts ...Option) error
	// Check if API gateway is up.
	Status(ctx context.Context, opts ...Option) (*StatusResp, error)
	// Calls auth endpoint and returns the auth details.
	// It does not update the client tokens.
	Auth(ctx context.Context, opts ...Option) (*AuthResp, error)
	// Calls refresh token endpoint by using the client's refresh token obtained by calling InitAuth.
	RefreshToken(ctx context.Context, opts ...Option) (*RefreshTokenResp, error)
	// Validate if the currently non-expired access token is valid.
	// To avoid updating the current access token of the client use WithDisableAutoUpdatingTokenOption(true) option.
	ValidateSession(ctx context.Context, opts ...Option) (bool, error)
	// Check balances information.
	Balance(ctx context.Context, opts ...Option) (*BalanceResp, error)
	// List available services.
	ListVerticals(ctx context.Context, opts ...Option) (*ListVerticalsResp, error)
	// Validate vend operation. It is called before carrying a transaction to get the transaction Id.
	VendValidate(ctx context.Context, body VendValidateBody, opts ...Option) (*VendValidateResp, error)
	// Execute a transaction.
	VendExecute(ctx context.Context, body VendExecuteBody, opts ...Option) (*VendExecuteResp, error)
	// Reports the status of a vend transaction.
	VendTransactionStatus(ctx context.Context, transactionId string, opts ...Option) (*VendTransactionStatusResp, error)
	// Get latest tokens of the meter number.
	ElectricityTokens(ctx context.Context, meterNo string, tokensCount int, opts ...Option) (*ElectricityTokenResp, error)
}

type Option interface {
	value() any
}

type Debugger interface {
	Debug(msg string, args ...any)
}

type ValidationError string

func (v ValidationError) Error() string {
	return string(v)
}

type VendTransactionStatusResp struct {
	Data struct {
		// transaction ID
		TransactionId string `json:"trxId"`
		// successful┃failed┃initiated┃pending┃timedout
		TransactionStatusId string `json:"trxStatusId"`
		// The customer account number i.e topup mobile number, electricity meterno, paytv decoder number etc.
		CustomerAccountNumber string `json:"customerAccountNumber"`
		// Where the Service Provider platform returns the customer account name, this field will bear it. Else, it will be empty
		CustomerAccountName string `json:"customerAccountName"`
		// The trx record creation timestamp on the Efashe platform
		CreatedAt string `json:"createdAt"`
		// The timestamp indicating when the trx record was last updated on the Efashe platforms.
		UpdatedAt string `json:"updatedAt"`
		// The amount tendered for the trx
		Amount float64 `json:"amount"`
		// The relevant currency code
		Currency string `json:"currency"`

		// The Service Provider Vend Information.
		// SP means MTN, Airtel, EUCL e.t.c...
		SpVendInfo struct {
			//  The Service Provider's Tax Identification Number (TIN)
			TIN string `json:"tin"`
			// The Service Provider's VAT number if provided.
			VatNo string `json:"vatNo"`
			// The timestamp returned by the Service Provider
			Tstamp string `json:"tstamp"`
			// The name of the Service Provider
			SpName string `json:"spName"`
			// This is the Service Provider's receipt number for products such as Electricity.
			ReceiptNo string `json:"receiptNo"`
			// Where the product is voucher based e.g Prepaid Electricity or Airtime Vouchers; this param bears the returned voucher.
			Voucher string `json:"voucher"`
			// Where relevant, this will bear the units of purchase.
			// E.g for prepaid electricity, these will be the actual electricity units (like 2.4 kwh), while for airtime, this will typically indicate the qty of the denomination bought.
			Units      string  `json:"units"`
			UnitsWorth float64 `json:"unitsWorth"`
			//  The amount posted to the Service Provider's platform for the trx
			TransactionAmount float64      `json:"trxAmount"`
			Deductions        []Deductions `json:"deductions"`
		} `json:"spVendInfo"`

		// May be our information. Agency might be referring to us.
		OurVendInfo struct {
			// The agency Tax Identification number. Where not available, this field will be blank.
			OurTIN        string       `json:"ourTIN"`
			OurDeductions []Deductions `json:"ourDeductions"`
			// The business name associated with the agency account
			AgencyName string `json:"agencyName"`
			// The Point of Presence (Branch) user friendly name. The default Branch is automatically named Main/HQ.
			BranchName string `json:"branchName"`
			// The Branch shortcode
			BranchShortCode string `json:"branchShortCode"`
			// The name of the staff who triggered the trx
			StaffName string `json:"staffName"`
			Narrative string `json:"narrative"`
		} `json:"ourVendInfo"`
	} `json:"data"`
}

type AuthResp struct {
	Data struct {
		AgencyPOP struct {
			POPId        string `json:"popId"`
			POPName      string `json:"popName"`
			POPShortCode string `json:"popShortCode"`
			POPStatusId  string `json:"popStatusId"`
			// Allowed: fixed┃roaming┃virtual
			PresenceId string `json:"presenceId"`
			// Allowed: main┃branch
			ClassId       string `json:"classId"`
			Address1Id    string `json:"address1Id"`
			Address2Id    string `json:"address2Id"`
			StreetAddress string `json:"streetAddress"`
		} `json:"agencyPOP"`
		AgencyAccount struct {
			AgencyId        string `json:"agencyId"`
			AgencyName      string `json:"agencyName"`
			AgencyShortCode string `json:"agencyShortCode"`
			// Allowed: L1┃L2┃L3
			AgencyLevelId string `json:"agencyLevelId"`
			// Allowed: active┃inactive┃suspended┃blacklisted┃kyc_pending
			AgencyStatusId string `json:"agencyStatusId"`
		} `json:"agencyAccount"`
		// The JWT to use to access protected enpoints
		AccessToken string `json:"accessToken"`
		// The JWT Refresh Token used to generate a new accessToken
		RefreshToken string `json:"refreshToken"`
		// The date and time (expressed in UTC) when the access token will expire
		AccessTokenExpiresAt string `json:"accessTokenExpiresAt"`
		// The date and time (expressed in UTC) when the refresh token will expire
		RefreshTokenExpiresAt string `json:"refreshTokenExpiresAt "`
	} `json:"data"`
}

type RefreshTokenResp struct {
	Data struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresAt    string `json:"expiresAt"`
	} `json:"data"`
}

type BalanceResp struct {
	Data           []Balance `json:"data"`
	Total          float64   `json:"total"`
	TotalFormatted string    `json:"totalFormatted"`
}

type ListVerticalsResp struct {
	Data []Vertical `json:"data"`
}

type VendValidateResp struct {
	Data struct {
		// e.g: electricity-eucl-rw, airtime-mtn-rw...
		PdtId string `json:"pdtId"`
		// e.g: EUCL Prepaid Electricity
		PdtName string `json:"pdtName"`
		// may be among active┃inactive┃suspended...
		PdtStatusId string `json:"pdtStatusId"`
		SharedVendInfo
		CustomerAccountName string `json:"customerAccountName"`
		// e.g: EUCL
		ServiceProviderName string `json:"svcProviderName"`
		// 	Allowed: fixed┃flexible
		// This param defines the vending model of the product or service as below:
		//
		// fixed - this means that only fixed amounts set by the Service Provider (SP) can be accepted. This is typical of Airtime vouchers and subscription package based services. When set, the integrated frontend should disable arbitrary input of amounts and only allow the selection of fixed denominations.
		// flexible - this means that the customer can tender any amount between the defined vendMin and vendMax amount. When set, the integrated frontend can allow for arbitrary input of amounts within the accepted vending range or a selection of amount from the selectAmount list
		VendUnitId string `json:"vendUnitId"`
		// This is the minimum vend amount that can be accepted for the trx
		VendMin float64 `json:"vendMin"`
		// This is the upper limit of the vend transaction amount. All amounts greater than this value will be rejected by the API.
		VendMax float64 `json:"vendMax"`

		// This is the trxId to use when calling the /vend/execute endpoint. This is required for idempotent processing of the transaction.
		TransactionId string `json:"trxId"`
		//	Allowed: voucher┃direct_topup
		//
		// voucher - this means the vend transaction results or returns a voucher
		// direct_topup - this means the vend transaction results in the direct topup of the customer's service account
		TransactionResult string `json:"trxResult"`
		// 	This shows the available wallet balance for the transaction. It is usually the sum of:
		//
		// main.AvailBal + refund.AvailBal - where commission auto depletion is disabled
		// main.AvailBal + refund.AvailBal + commission.AvailBal - where business policy allows for automatic depletion of the commission wallet.
		AvailTransactionBalance float64                  `json:"availTrxBalance"`
		DeliveryMethods         []VerticalDeliveryMethod `json:"deliveryMethods"`
	} `json:"data"`
}

type VendExecuteResp struct {
	Data struct {
		// The endpoint to poll for the trx status
		PollEndpoint string `json:"pollEndpoint"`
		// An estimate of when processing will complete in seconds.
		// This is designed to prevent polling clients from overwhelming the back-end with retries.
		RetryAfterSecs float64
	}
}

type ElectricityTokenResp struct {
	Data []ElectricityToken `json:"data"`
}

type StatusResp struct {
	// Allowed: operational┃degraded┃partial_outage┃major_outage┃maintenance Status of the API.
	Status string `json:"status"`
}

type VendValidateBody struct {
	SharedVendInfo
}

type VendExecuteBody struct {
	SharedVendInfo
	// transaction amount
	Amount float64 `json:"amount"`
	// The trxId returned in the /vend/validate response
	TransactionId string `json:"trxId"`
	// Allowed: print┃email┃sms┃direct_topup
	DeliveryMethodId string `json:"deliveryMethodId"`
	// This is the delivery destination of the trx receipt depending on the deliveryMethodId selected.
	// This can only be empty if the deliveryMethodId is set to print or direct_topup.
	// Else, an appropriate error will be returned.
	DeliverTo string `json:"deliverTo"`
	// This parameter defines the trx callback for asynchronous trx processing.
	// This field is only relevant to async systems/platforms.
	CallBack string `json:"callBack"`
}

type Balance struct {
	Id               string  `json:"id"`
	Name             string  `json:"name"`
	BalanceFormatted string  `json:"balanceFormatted"`
	Balance          float64 `json:"balance"`
}

type Deductions struct {
	DeductionName string `json:"deductionName"`
	// Allowed: percentage┃flat
	RateType string `json:"rateType"`
	// The deduction rate
	Rate string `json:"rate"`
	// the amount deducted
	AmountDeducted float64 `json:"amountDeducted"`
}

type Vertical struct {
	GenericInfo
	// Allowed: active┃inactive
	Status          string                   `json:"status"`
	CountryId       string                   `json:"countryId"`
	Input           []VerticalInput          `json:"input"`
	DeliveryMethods []VerticalDeliveryMethod `json:"deliveryMethods"`
}

type VerticalInput struct {
	GenericInfo
	Instruction string `json:"instruction"`
	// Allowed: selection┃text┃integer┃msisdn
	Type string `json:"type"`
}

type VerticalDeliveryMethod struct {
	GenericInfo
}

type SharedVendInfo struct {
	// This value should be obtained from the list verticals endpoint
	VerticalId string `json:"verticalId"`
	// This is the customer supplied service account number in the context of the product or service the customer is paying for:
	// Electricity Meter Number (if the service verticalId = electricity)
	// Decoder Number (if the service verticalId = paytv)
	// Mobile Phone Number to be topped up (if the service verticalId = airtime).
	// Note that this value should be the same during vend validate and vend execute.
	CustomerAccountNumber string `json:"customerAccountNumber"`
}

type GenericInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ElectricityToken struct {
	// The Number of Units of the token
	Units          float64 `json:"units"`
	Token          string  `json:"token"`
	Toke2          string  `json:"token2"`
	Toke3          string  `json:"token3"`
	MeterNo        string  `json:"meterno"`
	ReceiptNo      string  `json:"receipt_no"`
	Tstamp         string  `json:"tstamp"`
	RegulatoryFees float64 `json:"regulatory_fees"`
	Amount         float64 `json:"amount"`
	Vat            float64 `json:"vat"`
	CustomerName   string  `json:"customer_name"`
}
