package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newScoreMockServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}))
}

func newScoreClient(addr string) *Client {
	c, _ := New(addr, "token")
	return c
}

func TestNewScorer_MissingClient(t *testing.T) {
	logger, _ := newTestLogger()
	_, err := NewScorer(nil, logger, 0, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client is required")
}

func TestNewScorer_MissingLogger(t *testing.T) {
	srv := newScoreMockServer(map[string]interface{}{})
	defer srv.Close()
	c := newScoreClient(srv.URL)
	_, err := NewScorer(c, nil, 0, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "logger is required")
}

func TestScore_PerfectScore(t *testing.T) {
	srv := newScoreMockServer(map[string]interface{}{
		"username": "admin",
		"password": "s3cr3t",
	})
	defer srv.Close()
	c := newScoreClient(srv.URL)
	logger, _ := newTestLogger()
	scorer, err := NewScorer(c, logger, 0, []string{"username", "password"})
	require.NoError(t, err)

	result, err := scorer.Score("secret/data/app")
	require.NoError(t, err)
	assert.Equal(t, 100, result.Score)
	assert.Empty(t, result.Issues)
}

func TestScore_MissingRequiredKey(t *testing.T) {
	srv := newScoreMockServer(map[string]interface{}{
		"username": "admin",
	})
	defer srv.Close()
	c := newScoreClient(srv.URL)
	logger, _ := newTestLogger()
	scorer, err := NewScorer(c, logger, 0, []string{"username", "password"})
	require.NoError(t, err)

	result, err := scorer.Score("secret/data/app")
	require.NoError(t, err)
	assert.Equal(t, 80, result.Score)
	assert.Len(t, result.Issues, 1)
	assert.Contains(t, result.Issues[0], "password")
}

func TestScore_EmptyValue_DeductsPoints(t *testing.T) {
	srv := newScoreMockServer(map[string]interface{}{
		"username": "",
		"password": "ok",
	})
	defer srv.Close()
	c := newScoreClient(srv.URL)
	logger, _ := newTestLogger()
	scorer, err := NewScorer(c, logger, 0, nil)
	require.NoError(t, err)

	result, err := scorer.Score("secret/data/app")
	require.NoError(t, err)
	assert.Equal(t, 90, result.Score)
	assert.Len(t, result.Issues, 1)
}

func TestScore_ScoreNeverBelowZero(t *testing.T) {
	srv := newScoreMockServer(map[string]interface{}{
		"a": "",
		"b": "",
	})
	defer srv.Close()
	c := newScoreClient(srv.URL)
	logger, _ := newTestLogger()
	scorer, err := NewScorer(c, logger, 0, []string{"x", "y", "z", "w", "v", "u"})
	require.NoError(t, err)

	result, err := scorer.Score("secret/data/app")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Score, 0)
}
