package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSecrets_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "test-token", "secret")
	require.NoError(t, err)

	paths, err := c.ListSecrets(context.Background(), "myapp")
	require.NoError(t, err)
	assert.Empty(t, paths)
}

func TestListSecrets_FlatKeys(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Header.Get("X-Vault-Request") == "" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"keys":["alpha","beta"]}}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "test-token", "secret")
	require.NoError(t, err)

	paths, err := c.ListSecrets(context.Background(), "myapp")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"myapp/alpha", "myapp/beta"}, paths)
}

func TestListSecrets_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":["internal server error"]}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "test-token", "secret")
	require.NoError(t, err)

	_, err = c.ListSecrets(context.Background(), "myapp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing")
}
