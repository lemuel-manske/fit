package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeRejectsUnsupportedVersion(t *testing.T) {
	data := []byte(`{"protocolVersion": 999}`)

	_, err := Decode(data)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported protocol version")
}
