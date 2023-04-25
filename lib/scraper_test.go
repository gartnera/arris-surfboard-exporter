package lib

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScraper(t *testing.T) {
	scraper, err := NewScraper("https://192.168.100.1", os.Getenv("CREDS"))
	require.NoError(t, err)

	results, err := scraper.GetConnectionStatus()
	require.NoError(t, err)

	require.Greater(t, len(results.DownstreamBondedChannel), 1, "downstream results")
	require.Greater(t, len(results.UpstreamBondedChannel), 1, "upstream results")
}
