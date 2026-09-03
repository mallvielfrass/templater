package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	//prepare
	st := testStorage{
		files: make(map[string]testFile),
	}
	us := testUserStorage{
		users: make(map[string]bool),
	}
	router := NewRouter(&st, &us, nil, "khjnyjgkbjujkg6t7565ec5svdx", "", "", "")
	router.Mount()
	path := "/api/user"

	var reqBody bytes.Buffer

	req, err := http.NewRequest("POST", path, &reqBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userJWT))
	rr := httptest.NewRecorder()

	router.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	type RequestData struct {
		Jwt  string `json:"jwt"`
		User string `json:"user"`
	}
	var result RequestData
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, us.IsUserExist(result.User), true)
	//fmt.Printf("body:%v", result)

}

func TestForceSaveTaskDocument(t *testing.T) {
	st := testStorage{
		files:  make(map[string]testFile),
		taskOO: make(map[string]string),
	}
	st.SaveTaskOOKey("task-123", "key-456")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/command", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":0}`))
	}))
	defer ts.Close()

	router := NewRouter(&st, nil, nil, "jwt-secret", "http://base", "oo-secret", ts.URL)

	// Simulate callback arriving concurrently
	go func() {
		// Wait a small bit, then notify waiter
		time.Sleep(10 * time.Millisecond)
		router.notifySaveWaiters("key-456")
	}()

	err := router.forceSaveTaskDocument("task-123")
	require.NoError(t, err)
}
