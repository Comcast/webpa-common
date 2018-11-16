package bookkeeping

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func testHeaders(t *testing.T, originalHeader http.Header, headersToCopy []string, expectedKeyValues []interface{}) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		request = &http.Request{
			Header: originalHeader,
		}

		rf = RequestHeaders(headersToCopy...)
	)

	require.NotNil(rf)
	returnedKeyValuePair := rf(request)
	assert.Equal(expectedKeyValues, returnedKeyValuePair)

}

func TestBookkeepingHeaders(t *testing.T) {
	testData := []struct {
		originalHeader   http.Header
		headersToCopy    []string
		expectedResponse []interface{}
	}{
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			nil,
			[]interface{}{},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Does-Not-Exist"},
			[]interface{}{},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Does-Not-Exist", "X-Test-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Does-Not-Exist", "x-test-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Test-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Test-3", "X-Test-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"x-TeST-3", "X-tESt-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Test-3", "X-Test-1", "X-Test-2"},
			[]interface{}{"X-Test-1", []string{"foo"}, "X-Test-2", []string{"foo", "bar"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-TEST-3", "x-TEsT-1", "x-TesT-2"},
			[]interface{}{"X-Test-1", []string{"foo"}, "X-Test-2", []string{"foo", "bar"}},
		},
	}

	for i, record := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Logf("%#v", record)
			testHeaders(t, record.originalHeader, record.headersToCopy, record.expectedResponse)
		})
	}
}

func testReturnHeadersWithPrefix(t *testing.T, request *http.Request, headerPrefixToCopy []string, expectedKV []interface{}) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rf = RequestHeadersWithPrefix(headerPrefixToCopy...)
	)

	require.NotNil(rf)
	kv := rf(request)
	assert.Equal(expectedKV, kv)
}

func TestReturnHeadersWithPrefix(t *testing.T) {
	testData := []struct {
		request    *http.Request
		prefixs    []string
		expectedKV []interface{}
	}{
		{
			nil,
			nil,
			[]interface{}{},
		},
		{
			&http.Request{},
			nil,
			[]interface{}{},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}}},
			nil,
			[]interface{}{},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Does-Not-Exist"},
			[]interface{}{},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Does-Not-Exist", "X-Test-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Does-Not-Exist", "x-TeSt-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Test-3", "X-Test-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"x-TeST-3", "X-tESt-1"},
			[]interface{}{"X-Test-1", []string{"foo"}},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Test-3", "X-Test-1", "X-Test-2"},
			[]interface{}{"X-Test-1", []string{"foo"}, "X-Test-2", []string{"foo", "bar"}},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-TEST-3", "x-TEsT-1", "x-TesT-2"},
			[]interface{}{"X-Test-1", []string{"foo"}, "X-Test-2", []string{"foo", "bar"}},
		},
		{
			&http.Request{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-TEST"},
			[]interface{}{"X-Test-1", []string{"foo"}, "X-Test-2", []string{"foo", "bar"}},
		},
	}

	for i, record := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Logf("%#v", record)
			testReturnHeadersWithPrefix(t, record.request, record.prefixs, record.expectedKV)
		})
	}
}

func testUsePath(t *testing.T, request *http.Request, expectedKeyValuePair []interface{}) {
	assert := assert.New(t)

	kv := Path()(request)
	assert.Equal(expectedKeyValuePair, kv)
}

func TestUsePath(t *testing.T) {

	testData := []struct {
		request  *http.Request
		expected []interface{}
	}{
		{httptest.NewRequest("POST", "http://foobar.com:8080/", nil), []interface{}{"path", "/"}},
		{httptest.NewRequest("POST", "http://foobar.com:8080/neat", nil), []interface{}{"path", "/neat"}},
		{httptest.NewRequest("POST", "http://foobar.com:8080/neat/stuff", nil), []interface{}{"path", "/neat/stuff"}},
		{httptest.NewRequest("POST", "http://foobar.com:8080/neat/stuff?world=awesome", nil), []interface{}{"path", "/neat/stuff"}},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Logf("%#v", record)
			testUsePath(t, record.request, record.expected)
		})
	}
}