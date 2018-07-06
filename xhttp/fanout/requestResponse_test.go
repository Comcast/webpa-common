package fanout

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/Comcast/webpa-common/xhttp/xhttptest"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDefaultDecoderNilBody(t *testing.T) {
	assert := assert.New(t)
	assert.Panics(func() {
		DefaultDecoder(context.Background(), new(http.Request), make(http.Header))
	})
}

func testDefaultDecoderBodyError(t *testing.T) {
	var (
		assert = assert.New(t)

		expectedErr = errors.New("expected")
		body        = new(xhttptest.MockBody)
		fanout      = make(http.Header)
	)

	body.OnReadError(expectedErr).Once()
	actualCtx, actualBody, actualErr := DefaultDecoder(context.Background(), &http.Request{Body: body}, fanout)
	assert.Equal(context.Background(), actualCtx)
	assert.Empty(actualBody)
	assert.Equal(expectedErr, actualErr)
	assert.Empty(fanout)

	body.AssertExpectations(t)
}

func testDefaultDecoderBody(t *testing.T) {
	testData := []struct {
		expectedBody   []byte
		originalHeader http.Header
		expectedFanout http.Header
	}{
		{[]byte{}, http.Header{}, http.Header{}},
		{[]byte{}, http.Header{"X-Something": []string{"foo"}}, http.Header{}},
		{[]byte("here is some lovely content"), http.Header{"Content-Type": []string{"text/plain"}}, http.Header{"Content-Type": []string{"text/plain"}}},
		{[]byte("here is some lovely content"), http.Header{"X-Something": []string{"bar"}, "Content-Type": []string{"text/plain"}}, http.Header{"Content-Type": []string{"text/plain"}}},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert       = assert.New(t)
				original     = httptest.NewRequest("GET", "/foo", bytes.NewReader(record.expectedBody))
				actualFanout = make(http.Header)
			)

			original.Header = record.originalHeader
			actualCtx, actualBody, err := DefaultDecoder(context.Background(), original, actualFanout)
			assert.Equal(context.Background(), actualCtx)
			assert.Equal(record.expectedBody, actualBody)
			assert.NoError(err)
			assert.Equal(record.expectedFanout, actualFanout)
		})
	}
}

func TestDefaultDecoder(t *testing.T) {
	t.Run("NilBody", testDefaultDecoderNilBody)
	t.Run("BodyError", testDefaultDecoderBodyError)
	t.Run("Body", testDefaultDecoderBody)
}

func TestDefaultEncoder(t *testing.T) {
	testData := []struct {
		fanoutResponse   *http.Response
		expectedBody     []byte
		expectedOriginal http.Header
	}{
		{
			&http.Response{
				Header: http.Header{},
			},
			nil,
			http.Header{},
		},
		{
			&http.Response{
				Header: http.Header{},
			},
			[]byte{},
			http.Header{},
		},
		{
			&http.Response{
				Header: http.Header{"X-Something": []string{"foo"}},
			},
			[]byte("here is a lovely body"),
			http.Header{},
		},
		{
			&http.Response{
				Header: http.Header{"Content-Type": []string{"text/plain"}, "X-Something": []string{"foo"}},
			},
			[]byte("here is a lovely body"),
			http.Header{"Content-Type": []string{"text/plain"}},
		},
		{
			&http.Response{
				Header: http.Header{"Content-Type": []string{"text/plain"}},
			},
			[]byte("here is a lovely body"),
			http.Header{"Content-Type": []string{"text/plain"}},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)

				actualOriginal  = make(http.Header)
				actualBody, err = DefaultEncoder(context.Background(), Result{Body: record.expectedBody, Response: record.fanoutResponse}, actualOriginal)
			)

			assert.Equal(record.expectedBody, actualBody)
			assert.NoError(err)
			assert.Equal(record.expectedOriginal, actualOriginal)
		})
	}
}

func testForwardHeaders(t *testing.T, originalHeader http.Header, headersToCopy []string, expectedFanoutHeader http.Header) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		ctx     = context.WithValue(context.Background(), "foo", "bar")

		original = &http.Request{
			Header: originalHeader,
		}

		fanout = &http.Request{
			Header: make(http.Header),
		}

		rf = ForwardHeaders(headersToCopy...)
	)

	require.NotNil(rf)
	returnedCtx, err := rf(ctx, original, fanout)
	assert.Equal(ctx, returnedCtx)
	assert.NoError(err)
	assert.Equal(expectedFanoutHeader, fanout.Header)
}

func TestForwardHeaders(t *testing.T) {
	testData := []struct {
		originalHeader       http.Header
		headersToCopy        []string
		expectedFanoutHeader http.Header
	}{
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			nil,
			http.Header{},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Does-Not-Exist"},
			http.Header{},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Does-Not-Exist", "X-Test-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Does-Not-Exist", "x-test-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Test-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Test-3", "X-Test-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"x-TeST-3", "X-tESt-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-Test-3", "X-Test-1", "X-Test-2"},
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}},
		},
		{
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}},
			[]string{"X-TEST-3", "x-TEsT-1", "x-TesT-2"},
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}},
		},
	}

	for i, record := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Logf("%#v", record)
			testForwardHeaders(t, record.originalHeader, record.headersToCopy, record.expectedFanoutHeader)
		})
	}
}

func testUsePathPanics(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rf = UsePath("/foo")
	)

	require.NotNil(rf)
	assert.Panics(func() {
		rf(context.Background(), httptest.NewRequest("GET", "/", nil), new(http.Request))
	})
}

func testUsePath(t *testing.T, fanout http.Request) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rf = UsePath("/api/v1/device/foo/bar")
	)

	require.NotNil(rf)

	rf(context.Background(), httptest.NewRequest("GET", "/", nil), &fanout)
	assert.Equal("/api/v1/device/foo/bar", fanout.URL.Path)
	assert.Empty(fanout.URL.RawPath)
}

func TestUsePath(t *testing.T) {
	t.Run("Panics", testUsePathPanics)

	testData := []http.Request{
		{URL: new(url.URL)},
		{URL: &url.URL{Host: "foobar.com:8080", Path: "/original"}},
		{URL: &url.URL{Host: "foobar.com:8080", Path: "/something", RawPath: "this is a raw path"}},
		{URL: &url.URL{Host: "foobar.com:8080", RawPath: "this is a raw path"}},
	}

	for i, fanout := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testUsePath(t, fanout)
		})
	}
}

func testForwardVariableAsHeaderMissing(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		ctx      = context.WithValue(context.Background(), "foo", "bar")
		original = httptest.NewRequest("GET", "/", nil)
		fanout   = httptest.NewRequest("GET", "/", nil)
		rf       = ForwardVariableAsHeader("test", "X-Test")
	)

	require.NotNil(rf)
	returnedCtx, err := rf(ctx, original, fanout)
	assert.Equal(ctx, returnedCtx)
	assert.NoError(err)
	assert.Equal("", fanout.Header.Get("X-Test"))
}

func testForwardVariableAsHeaderValue(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		ctx       = context.WithValue(context.Background(), "foo", "bar")
		variables = map[string]string{
			"test": "foobar",
		}

		original = mux.SetURLVars(
			httptest.NewRequest("GET", "/", nil),
			variables,
		)

		fanout = httptest.NewRequest("GET", "/", nil)
		rf     = ForwardVariableAsHeader("test", "X-Test")
	)

	require.NotNil(rf)
	returnedCtx, err := rf(ctx, original, fanout)
	assert.Equal(ctx, returnedCtx)
	assert.NoError(err)
	assert.Equal("foobar", fanout.Header.Get("X-Test"))
}

func TestForwardVariableAsHeader(t *testing.T) {
	t.Run("Missing", testForwardVariableAsHeaderMissing)
	t.Run("Value", testForwardVariableAsHeaderValue)
}

func testReturnHeaders(t *testing.T, fanoutResponse *http.Response, headersToCopy []string, expectedResponseHeader http.Header) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		ctx     = context.WithValue(context.Background(), "foo", "bar")

		response = httptest.NewRecorder()
		rf       = ReturnHeaders(headersToCopy...)
	)

	require.NotNil(rf)
	assert.Equal(ctx, rf(ctx, Result{Response: fanoutResponse}, response.Header()))
	assert.Equal(expectedResponseHeader, response.Header())
}

func TestReturnHeaders(t *testing.T) {
	testData := []struct {
		fanoutResponse         *http.Response
		headersToCopy          []string
		expectedResponseHeader http.Header
	}{
		{
			nil,
			nil,
			http.Header{},
		},
		{
			&http.Response{},
			nil,
			http.Header{},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}}},
			nil,
			http.Header{},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Does-Not-Exist"},
			http.Header{},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Does-Not-Exist", "X-Test-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Does-Not-Exist", "x-TeSt-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Test-3", "X-Test-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"x-TeST-3", "X-tESt-1"},
			http.Header{"X-Test-1": []string{"foo"}},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-Test-3", "X-Test-1", "X-Test-2"},
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}},
		},
		{
			&http.Response{Header: http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}, "X-Test-3": []string{}}},
			[]string{"X-TEST-3", "x-TEsT-1", "x-TesT-2"},
			http.Header{"X-Test-1": []string{"foo"}, "X-Test-2": []string{"foo", "bar"}},
		},
	}

	for i, record := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Logf("%#v", record)
			testReturnHeaders(t, record.fanoutResponse, record.headersToCopy, record.expectedResponseHeader)
		})
	}
}
