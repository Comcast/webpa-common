package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/natefinch/lumberjack.v2"
	"bytes"
)

func testOptionsLoggerFactory(t *testing.T) {
	assert := assert.New(t)

	for _, o := range []*Options{nil, new(Options), {JSON: true}, {JSON: false}} {
		assert.NotNil(o.loggerFactory())
	}
}

func testOptionsOutput(t *testing.T) {
	assert := assert.New(t)

	for _, o := range []*Options{nil, {File: StdoutFile}} {
		output := o.output()
		assert.NotNil(output)
		assert.NotPanics(func() {
			_, err := output.Write([]byte("expected output: this shouldn't panic\n"))
			assert.NoError(err)
		})
	}

	var (
		rolling = &Options{
			File:       "foobar.log",
			MaxSize:    689328,
			MaxAge:     9,
			MaxBackups: 454,
		}

		output               = rolling.output()
		lumberjackLogger, ok = output.(*lumberjack.Logger)
	)

	assert.True(ok)
	assert.Equal("foobar.log", lumberjackLogger.Filename)
	assert.Equal(689328, lumberjackLogger.MaxSize)
	assert.Equal(9, lumberjackLogger.MaxAge)
	assert.Equal(454, lumberjackLogger.MaxBackups)
}

func testOptionsLevel(t *testing.T) {
	assert := assert.New(t)

	for _, o := range []*Options{nil, new(Options)} {
		assert.Empty(o.level())
	}

	assert.Equal("info", (&Options{Level: "info"}).level())
}

func TestOptions(t *testing.T) {
	t.Run("LoggerFactory", testOptionsLoggerFactory)
	t.Run("Output", testOptionsOutput)
	t.Run("Level", testOptionsLevel)
}

func TestOptionsWithReformatLogger(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer

	o := &Options{
		File:       StdoutFile,
		FormatType: "term",
		TermOptions: TextFormatter{
			DisableColors: true,
			DisableLevelTruncation: false,
		},
	}
	logger := o.loggerFactory()(&buf)
	assert.NotNil(logger)
	logger.Log("msg", "testing")
	t.Log(buf.String())
	assert.NotNil(buf.String())
	assert.Equal("INFO[00000] testing                                     \n", buf.String())
}

func TestOptionsForOldFmt(t *testing.T){
	assert := assert.New(t)

	var buf bytes.Buffer

	o := &Options{
		File:       StdoutFile,
		FormatType: "fmt",
	}
	logger := o.loggerFactory()(&buf)
	assert.NotNil(logger)
	logger.Log("msg", "testing")
	t.Log(buf.String())
	assert.NotNil(buf.String())
	assert.Equal("msg=testing\n", buf.String())
}
