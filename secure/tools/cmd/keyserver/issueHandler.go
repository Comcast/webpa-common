package main

import (
	"errors"
	"fmt"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
	"github.com/gorilla/schema"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	KeyIDVariableName = "kid"
)

var (
	ErrorMissingKeyID = errors.New("A kid parameter is required")

	defaultSigningMethod crypto.SigningMethod = crypto.SigningMethodRS256

	supportedSigningMethods = map[string]crypto.SigningMethod{
		defaultSigningMethod.Alg():      defaultSigningMethod,
		crypto.SigningMethodRS384.Alg(): crypto.SigningMethodRS384,
		crypto.SigningMethodRS512.Alg(): crypto.SigningMethodRS512,
	}

	supportedNumericDateLayouts = []string{
		time.RFC3339,
		time.RFC822,
		time.RFC822Z,
	}
)

// NumericDate represents a JWT NumericDate as specified in:
// https://tools.ietf.org/html/rfc7519#section-2
//
// A number of formats for numeric dates are allowed, and each
// is converted appropriately:
//
// (1) An int64 value, which is interpreted as the exact value to use
// (2) A valid time.Duration, which is added to time.Now() to compute the value
// (3) An absolute date specified in RFC33399 or RFC822 formates.  See the time package for details.
type NumericDate struct {
	duration time.Duration
	absolute int64
}

func (nd *NumericDate) UnmarshalText(raw []byte) error {
	if len(raw) == 0 {
		*nd = NumericDate{}
		return nil
	}

	text := string(raw)

	if value, err := strconv.ParseInt(text, 10, 64); err == nil {
		*nd = NumericDate{duration: 0, absolute: value}
		return nil
	}

	if duration, err := time.ParseDuration(text); err == nil {
		*nd = NumericDate{duration: duration, absolute: 0}
		return nil
	}

	for _, layout := range supportedNumericDateLayouts {
		if value, err := time.Parse(layout, text); err == nil {
			*nd = NumericDate{duration: 0, absolute: value.Unix()}
			return nil
		}
	}

	return fmt.Errorf("Unparseable datetime: %s", text)
}

// Compute calculates the NumericDate value given a point in time
// assumed to be "now".  Use of this level of indirection allows a
// single time value to be used in all calculations when issuing JWTs.
func (nd *NumericDate) Compute(now time.Time) int64 {
	if nd.absolute != 0 {
		return nd.absolute
	}

	return now.Add(nd.duration).Unix()
}

// SigningMethod is a custom type which holds the alg value.
type SigningMethod struct {
	crypto.SigningMethod
}

func (s *SigningMethod) UnmarshalText(raw []byte) error {
	if len(raw) == 0 {
		*s = SigningMethod{defaultSigningMethod}
		return nil
	}

	text := string(raw)
	value, ok := supportedSigningMethods[text]
	if ok {
		*s = SigningMethod{value}
		return nil
	}

	return fmt.Errorf("Unsupported algorithm: %s", text)
}

// IssueRequest contains the information necessary for issuing a JWS.
// Any custom claims must be transmitted separately.
type IssueRequest struct {
	Now time.Time `schema:"-"`

	KeyID     string         `schema:"kid"`
	Algorithm *SigningMethod `schema:"alg"`

	Expires   *NumericDate `schema:"exp"`
	NotBefore NumericDate  `schema:"nbf"`

	JWTID    *string   `schema:"jti"`
	Subject  *string   `scheme:"sub"`
	Audience *[]string `scheme:"aud"`
}

func (ir *IssueRequest) SigningMethod() crypto.SigningMethod {
	if ir.Algorithm != nil {
		return ir.Algorithm.SigningMethod
	}

	return defaultSigningMethod
}

// AddToHeader adds the appropriate header information from this issue request
func (ir *IssueRequest) AddToHeader(header map[string]interface{}) error {
	// right now, we just add the kid
	header[KeyIDVariableName] = ir.KeyID
	return nil
}

// AddToClaims takes the various parts of this issue request and formats them
// appropriately into a supplied jwt.Claims object.
func (ir *IssueRequest) AddToClaims(claims jwt.Claims) error {
	claims.SetIssuedAt(ir.Now)
	return nil
}

func NewIssueRequest(decoder *schema.Decoder, source map[string][]string) (*IssueRequest, error) {
	issueRequest := &IssueRequest{}
	if err := decoder.Decode(issueRequest, source); err != nil {
		return nil, err
	}

	if len(issueRequest.KeyID) == 0 {
		return nil, ErrorMissingKeyID
	}

	issueRequest.Now = time.Now()
	return issueRequest, nil
}

// BasicHandler handles all keyserver requests
type IssueHandler struct {
	BasicHandler
	issuer  string
	decoder *schema.Decoder
}

// issue handles all the common logic for issuing a JWS token
func (handler *IssueHandler) issue(response http.ResponseWriter, issueRequest *IssueRequest, claims jwt.Claims) {
	issueKey, ok := handler.keyStore.PrivateKey(issueRequest.KeyID)
	if !ok {
		handler.httpError(response, http.StatusBadRequest, fmt.Sprintf("No such key: %s", issueRequest.KeyID))
		return
	}

	if claims == nil {
		claims = make(jwt.Claims)
	}

	issuedJWT := jws.NewJWT(jws.Claims(claims), issueRequest.SigningMethod())
	if err := issueRequest.AddToClaims(issuedJWT.Claims()); err != nil {
		handler.httpError(response, http.StatusInternalServerError, err.Error())
		return
	}

	issuedJWT.Claims().SetIssuer(handler.issuer)
	issuedJWS := issuedJWT.(jws.JWS)
	if err := issueRequest.AddToHeader(issuedJWS.Protected()); err != nil {
		handler.httpError(response, http.StatusInternalServerError, err.Error())
		return
	}

	compact, err := issuedJWS.Compact(issueKey)
	if err != nil {
		handler.httpError(response, http.StatusInternalServerError, err.Error())
		return
	}

	response.Header().Set("Content-Type", "application/jwt")
	response.Write(compact)
}

// SimpleIssue handles requests with no body, appropriate for simple use cases.
func (handler *IssueHandler) SimpleIssue(response http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		handler.httpError(response, http.StatusBadRequest, err.Error())
		return
	}

	issueRequest, err := NewIssueRequest(handler.decoder, request.Form)
	if err != nil {
		handler.httpError(response, http.StatusBadRequest, err.Error())
		return
	}

	handler.issue(response, issueRequest, nil)
}

// IssueUsingBody accepts a JSON claims document, to which it then adds all the standard
// claims mentioned in request parameters, e.g. exp.  It then uses the merged claims
// in an issued JWS.
func (handler *IssueHandler) IssueUsingBody(response http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		handler.httpError(response, http.StatusBadRequest, err.Error())
		return
	}

	issueRequest, err := NewIssueRequest(handler.decoder, request.Form)
	if err != nil {
		handler.httpError(response, http.StatusBadRequest, err.Error())
		return
	}

	// this variant reads the claims directly from the request body
	claims := make(jwt.Claims)
	if request.Body != nil {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			handler.httpError(response, http.StatusBadRequest, err.Error())
			return
		}

		if len(body) > 0 {
			if err := claims.UnmarshalJSON(body); err != nil {
				handler.httpError(response, http.StatusBadRequest, err.Error())
				return
			}
		}
	}

	handler.issue(response, issueRequest, claims)
}
