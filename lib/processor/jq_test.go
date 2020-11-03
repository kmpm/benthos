package processor

import (
	"strconv"
	"testing"

	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/message"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/gabs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJQAllParts(t *testing.T) {
	conf := NewConfig()
	conf.JQ.Query = ".foo.bar"

	jSet, err := NewJQ(conf, nil, log.Noop(), metrics.Noop())
	require.NoError(t, err)

	msgIn := message.New([][]byte{
		[]byte(`{"foo":{"bar":0}}`),
		[]byte(`{"foo":{"bar":1}}`),
		[]byte(`{"foo":{"bar":2}}`),
	})
	msgs, res := jSet.ProcessMessage(msgIn)
	require.Nil(t, res)
	require.Len(t, msgs, 1)
	for i, part := range message.GetAllBytes(msgs[0]) {
		assert.Equal(t, strconv.Itoa(i), string(part))
	}
}

func TestJQValidation(t *testing.T) {
	conf := NewConfig()
	conf.JQ.Query = ".foo.bar"

	jSet, err := NewJQ(conf, nil, log.Noop(), metrics.Noop())
	require.NoError(t, err)

	msgIn := message.New([][]byte{[]byte("this is bad json")})
	msgs, res := jSet.ProcessMessage(msgIn)

	require.Nil(t, res)
	require.Len(t, msgs, 1)

	assert.Equal(t, "this is bad json", string(message.GetAllBytes(msgs[0])[0]))
}

func TestJQMutation(t *testing.T) {
	conf := NewConfig()
	conf.JQ.Query = `{foo: .foo} | .foo.bar = "baz"`

	jSet, err := NewJQ(conf, nil, log.Noop(), metrics.Noop())
	require.NoError(t, err)

	ogObj := gabs.New()
	ogObj.Set("is this", "foo", "original", "content")
	ogObj.Set("remove this", "bar")
	ogExp := ogObj.String()

	msgIn := message.New(make([][]byte, 1))
	msgIn.Get(0).SetJSON(ogObj.Data())
	msgs, res := jSet.ProcessMessage(msgIn)
	require.Nil(t, res)
	require.Len(t, msgs, 1)

	assert.Equal(t, `{"foo":{"bar":"baz","original":{"content":"is this"}}}`, string(message.GetAllBytes(msgs[0])[0]))
	assert.Equal(t, ogExp, ogObj.String())
}

func TestJQ(t *testing.T) {
	type jTest struct {
		name   string
		path   string
		input  string
		output string
	}

	tests := []jTest{
		{
			name:   "select obj",
			path:   ".foo.bar",
			input:  `{"foo":{"bar":{"baz":1}}}`,
			output: `{"baz":1}`,
		},
		{
			name:   "select array",
			path:   ".foo.bar",
			input:  `{"foo":{"bar":["baz","qux"]}}`,
			output: `["baz","qux"]`,
		},
		{
			name:   "select obj as str",
			path:   ".foo.bar",
			input:  `{"foo":{"bar":"{\"baz\":1}"}}`,
			output: `"{\"baz\":1}"`,
		},
		{
			name:   "select str",
			path:   ".foo.bar",
			input:  `{"foo":{"bar":"hello world"}}`,
			output: `"hello world"`,
		},
		{
			name:   "select float",
			path:   ".foo.bar",
			input:  `{"foo":{"bar":0.123}}`,
			output: `0.123`,
		},
		{
			name:   "select int",
			path:   ".foo.bar",
			input:  `{"foo":{"bar":123}}`,
			output: `123`,
		},
		{
			name:   "select bool",
			path:   ".foo.bar",
			input:  `{"foo":{"bar":true}}`,
			output: `true`,
		},
	}

	for _, test := range tests {
		conf := NewConfig()
		conf.JQ.Query = test.path

		jSet, err := NewJQ(conf, nil, log.Noop(), metrics.Noop())
		require.NoError(t, err)

		inMsg := message.New(
			[][]byte{
				[]byte(test.input),
			},
		)
		msgs, _ := jSet.ProcessMessage(inMsg)
		require.Len(t, msgs, 1)
		assert.Equal(t, test.output, string(message.GetAllBytes(msgs[0])[0]))
	}
}
