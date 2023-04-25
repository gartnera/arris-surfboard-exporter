package lib

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/cmconnectionstatus.html
var exampleConnectionStatus []byte

func TestParser(t *testing.T) {
	parser := &ConnectionStatusParser{}
	err := parser.ParseBytes(exampleConnectionStatus)
	require.NoError(t, err)

	results := parser.results

	require.Len(t, results.UpstreamBondedChannel, 5)

	for _, res := range results.UpstreamBondedChannel {
		require.NotZero(t, res.Frequency, "frequency")
		require.NotZero(t, res.Power, "power")
	}

	require.Len(t, results.DownstreamBondedChannel, 34)

	for _, res := range results.DownstreamBondedChannel {
		require.NotZero(t, res.Frequency, "frequency")
		if res.ChannelID != "160" {
			require.NotZero(t, res.Power, "power")
		}
		require.NotZero(t, res.SNR, "snr")
		require.NotZero(t, res.Corrected, "corrected")
		require.NotZero(t, res.Uncorrectables, "uncorrectables")
	}
}
