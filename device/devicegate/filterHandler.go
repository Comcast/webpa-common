package devicegate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/webpa-common/xhttp"
)

type FilterHandler struct {
	Gate DeviceGate
}

func (fh *FilterHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {

	logger := logging.GetLogger(request.Context())

	method := request.Method
	if method == "GET" {
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(response, `{"filters": %s}`, filtersToString(fh.Gate))
	} else if method == "POST" || method == "PUT" || method == "DELETE" {
		var message filterRequest
		msgBytes, err := ioutil.ReadAll(request.Body)
		request.Body.Close()

		if err != nil {
			logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "could not read request body", logging.ErrorKey(), err)
			xhttp.WriteError(response, http.StatusBadRequest, err)
			return
		}

		err = json.Unmarshal(msgBytes, &message)
		if err != nil {
			logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "could not unmarshal request body", logging.ErrorKey(), err)
			xhttp.WriteError(response, http.StatusBadRequest, err)
			return
		}

		if len(message.Key) == 0 {
			logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "no filter key found")
			xhttp.WriteErrorf(response, http.StatusBadRequest, "missing filter key")
			return
		}

		if method == "POST" || method == "PUT" {

			if len(message.Values) == 0 {
				logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "no filter values found")
				xhttp.WriteErrorf(response, http.StatusBadRequest, "missing filter values")
				return
			}

			allowedFilters, allowedFiltersFound := fh.Gate.GetAllowedFilters()

			if allowedFiltersFound {
				if !allowedFilters.Has(message.Key) {
					logger.Log(level.Key(), level.ErrorValue(), logging.MessageKey(), "filter key is not allowed", "key: ", message.Key)
					xhttp.WriteErrorf(response, http.StatusBadRequest, "filter key %s is not allowed. Allowed filters: %v", message.Key, allowedFilters.String())
					return
				}
			}

			fh.Gate.SetFilter(message.Key, message.Values)

		} else if method == "DELETE" {
			fh.Gate.DeleteFilter(message.Key)
		}

		filters := filtersToString(fh.Gate)
		logger.Log(level.Key(), level.InfoValue(), logging.MessageKey(), "gate filters updated", "filters", filters)

		response.WriteHeader(http.StatusOK)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(response, `{"filters": %s}`, filters)
	}
}

func writeFilters(b *strings.Builder) func(string, Set) {
	var needsComma bool

	return func(key string, val Set) {
		if needsComma {
			b.WriteString(",\n")
			needsComma = false
		}

		fmt.Fprintf(b, `"%s": `, key)
		fmt.Fprintf(b, "%s", val.String())
		needsComma = true
	}
}

// manual construction of JSON string
// func writeFilters(b *strings.Builder) func(string, Set) {
// 	var needsComma bool
// 	var currentKey string

// 	return func(key string, val interface{}) {
// 		if currentKey != key {
// 			if len(currentKey) > 0 {
// 				b.WriteString("]\n")
// 				needsComma = false
// 			}

// 			currentKey = key
// 			fmt.Fprintf(b, `"%s": [`, currentKey)
// 		}

// 		if needsComma {
// 			b.WriteString(", ")
// 			needsComma = false
// 		}

// 		fmt.Fprintf(b, `"%v"`, val)
// 		needsComma = true
// 	}
// }

func filtersToString(g DeviceGate) string {
	var b strings.Builder
	var filtersBuilder strings.Builder
	b.WriteString("{")
	g.VisitAll(writeFilters(&filtersBuilder))

	if filtersBuilder.Len() > 0 {
		filtersBuilder.WriteString("\n")
		b.WriteString(filtersBuilder.String())
		filtersBuilder.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}