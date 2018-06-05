package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Comcast/webpa-common/logging"
	"github.com/Comcast/webpa-common/secure"
	"github.com/SermoDigital/jose/jws"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	// The Content-Type value for JSON
	JsonContentType string = "application/json; charset=UTF-8"

	// The Content-Type header
	ContentTypeHeader string = "Content-Type"

	// The X-Content-Type-Options header
	ContentTypeOptionsHeader string = "X-Content-Type-Options"

	// NoSniff is the value used for content options for errors written by this package
	NoSniff string = "nosniff"
)

// WriteJsonError writes a standard JSON error to the response
func WriteJsonError(response http.ResponseWriter, code int, message string) error {
	response.Header().Set(ContentTypeHeader, JsonContentType)
	response.Header().Set(ContentTypeOptionsHeader, NoSniff)

	response.WriteHeader(code)
	_, err := fmt.Fprintf(response, `{"message": "%s"}`, message)
	return err
}

// AuthorizationHandler provides decoration for http.Handler instances and will
// ensure that requests pass the validator.  Note that secure.Validators is a Validator
// implementation that allows chaining validators together via logical OR.
type AuthorizationHandler struct {
	HeaderName          string
	ForbiddenStatusCode int
	Validator           secure.Validator
	Logger              log.Logger
	measures            *secure.JWTValidationMeasures
}

// headerName returns the authorization header to use, either a.HeaderName
// or secure.AuthorizationHeader if no header is supplied
func (a AuthorizationHandler) headerName() string {
	if len(a.HeaderName) > 0 {
		return a.HeaderName
	}

	return secure.AuthorizationHeader
}

// forbiddenStatusCode returns a.ForbiddenStatusCode if supplied, otherwise
// http.StatusForbidden is returned
func (a AuthorizationHandler) forbiddenStatusCode() int {
	if a.ForbiddenStatusCode > 0 {
		return a.ForbiddenStatusCode
	}

	return http.StatusForbidden
}

func (a AuthorizationHandler) logger() log.Logger {
	if a.Logger != nil {
		return a.Logger
	}

	return logging.DefaultLogger()
}

// Decorate provides an Alice-compatible constructor that validates requests
// using the configuration specified.
func (a AuthorizationHandler) Decorate(delegate http.Handler) http.Handler {
	// if there is no validator, there's no point in decorating anything
	if a.Validator == nil {
		return delegate
	}

	var (
		headerName          = a.headerName()
		forbiddenStatusCode = a.forbiddenStatusCode()
		logger              = a.logger()
		errorLog            = logging.Error(logger)
	)

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		headerValue := request.Header.Get(headerName)
		if len(headerValue) == 0 {
			errorLog.Log(logging.MessageKey(), "missing header", "name", headerName)
			WriteJsonError(response, forbiddenStatusCode, fmt.Sprintf("missing header: %s", headerName))

			if a.measures != nil {
				a.measures.ValidationReason.With("reason", "missing_header").Add(1)
			}
			return
		}

		token, err := secure.ParseAuthorization(headerValue)
		if err != nil {
			errorLog.Log(logging.MessageKey(), "invalid authorization header", "name", headerName, "token", headerValue, logging.ErrorKey(), err)
			WriteJsonError(response, forbiddenStatusCode, fmt.Sprintf("Invalid authorization header [%s]: %s", headerName, err.Error()))

			if a.measures != nil {
				a.measures.ValidationReason.With("reason", "invalid_header").Add(1)
			}
			return
		}

		contextValues := &ContextValues{
			Method: request.Method,
			Path:   request.URL.Path,
		}

		sharedContext := NewContextWithValue(request.Context(), contextValues)

		valid, err := a.Validator.Validate(sharedContext, token)
		if err == nil && valid {
			if err := populateContextValues(token, contextValues); err != nil {
				logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "unable to populate context", logging.ErrorKey(), err)
			}

			request = request.WithContext(sharedContext)
			delegate.ServeHTTP(response, request)
			return
		}

		errorLog.Log(
			logging.MessageKey(), "request denied",
			"validator-response", valid,
			"validator-error", err,
			"sat-client-id", contextValues.SatClientID,
			"token", headerValue,
			"method", request.Method,
			"url", request.URL,
			"user-agent", request.Header.Get("User-Agent"),
			"content-length", request.ContentLength,
			"remoteAddress", request.RemoteAddr,
		)

		WriteJsonError(response, forbiddenStatusCode, "request denied")
	})
}

//DefineMeasures facilitates clients to define authHandler metrics tools
func (a *AuthorizationHandler) DefineMeasures(m *secure.JWTValidationMeasures) {
	a.measures = m
}

func populateContextValues(token *secure.Token, values *ContextValues) error {
	values.SatClientID = "N/A"

	if token.Type() != secure.Bearer {
		return nil
	}

	jwsToken, err := secure.DefaultJWSParser.ParseJWS(token)
	if err != nil {
		return err
	}

	claims, ok := jwsToken.Payload().(jws.Claims)
	if !ok {
		return errors.New("no claims")
	}

	if sub, ok := claims.Get("sub").(string); ok {
		values.SatClientID = sub
	}

	if allowedResources, ok := claims.Get("allowedResources").(map[string]interface{}); ok {
		if allowedPartners, ok := allowedResources["allowedPartners"].([]string); ok {
			values.PartnerIDs = allowedPartners
		}
	}

	return nil
}
