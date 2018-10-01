package logging

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"bytes"
	"fmt"
	"time"
	"github.com/go-kit/kit/log"
	"errors"
)

func reformatLoggerSetup() (*bytes.Buffer, log.Logger) {
	buf := &bytes.Buffer{}
	return buf, NewReformatLogger(buf, &TextFormatter{
		DisableLevelTruncation: false,
		DisableColors:          true,
	})
}

func TestReformatLogger(t *testing.T) {
	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()

	err := logger.Log("msg", "hello", "level", "warn", "ts", time.Now(), "isCool", "YES")
	assert.Nil(err)

	//test
	expected := "WARN[00000] hello                                       isCool=YES \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithNoMSG(t *testing.T) {
	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()

	err := logger.Log("level", "error", "key", "value")
	assert.Nil(err)

	expected := "ERRO[00000]                                             key=value \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithError(t *testing.T) {
	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()

	err := logger.Log("level", "error", "key", "value", "err", errors.New("unknown error"))
	assert.Nil(err)

	expected := "ERRO[00000]                                             ERR:unknown error key=value \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithStringError(t *testing.T) {
	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()

	err := logger.Log("level", "error", "key", "value", "err", "unknown error")
	assert.Nil(err)

	expected := "ERRO[00000]                                             ERR:unknown error key=value \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithStringColorError(t *testing.T) {
	assert := assert.New(t)

	buf := &bytes.Buffer{}
	logger := NewReformatLogger(buf, &TextFormatter{
		DisableLevelTruncation: false,
		DisableColors:          false,
	})

	err := logger.Log("level", "error", "key", "value", "err", "unknown error")
	assert.Nil(err)

	expected := "\x1b[31mERRO\x1b[39m[00000]                                             \x1b[41m\x1b[30mERR\x1b[49m\x1b[39m:unknown error \x1b[31mkey\x1b[39m=value \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithNoLevel(t *testing.T) {
	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()

	err := logger.Log("msg", "Calling Endpoint")
	assert.Nil(err)

	//test
	expected := "INFO[00000] Calling Endpoint                            \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithTime(t *testing.T) {
	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()

	err := logger.Log("msg", "hello", "ts", time.Now().Add(time.Second*5))
	assert.Nil(err)

	expected := "INFO[00005] hello                                       \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithComplexKey(t *testing.T) {
	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()

	key := errors.New("error_key")

	value := make(map[string]string)
	value["k"] = "v"

	err := logger.Log("msg", "complex keys with map value and array value", "ts", time.Now().Add(time.Second*5), key, value)
	assert.Nil(err)

	expected := "INFO[00005] complex keys with map value and array value error_key=map[string]string{\"k\":\"v\"} \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}

func TestReformatLoggerWithComplexStruct(t *testing.T) {
	type neat struct{
		A int
		B string
		C []string
	}
	neatA := neat{
		A: 42,
		B: "everything",
		C: []string{"is", "awesome"},
	}

	assert := assert.New(t)

	buf, logger := reformatLoggerSetup()
	err := logger.Log("msg", "complex keys with map value and array value", "ts", time.Now().Add(time.Second*25), "life", neatA)
	assert.Nil(err)

	expected := "INFO[00025] complex keys with map value and array value life=logging.neat{A:42, B:\"everything\", C:[]string{\"is\", \"awesome\"}} \n"
	actual := buf.String()
	assert.Equal(expected, actual, fmt.Sprintf("want %#v, have %#v", expected, actual))
}
