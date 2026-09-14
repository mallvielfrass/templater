package router

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mallvielfrass/templater/internal/models"
	"github.com/mallvielfrass/templater/internal/sampledata"
	"github.com/stretchr/testify/require"
)

func testApp(t *testing.T) (*Router, *testStorage, *testUserStorage) {
	t.Helper()
	st := &testStorage{
		files:     map[string]testFile{},
		docs:      map[string][]byte{},
		generated: map[string][]byte{},
		owners:    map[string]string{},
	}
	us := &testUserStorage{users: map[string]bool{}}
	ts := &testTaskStorage{tasks: map[string]models.Task{}}
	r := NewRouter(st, us, ts, "test-jwt-secret", "http://base", "oo-secret", "")
	r.Mount()
	return r, st, us
}

func createSession(t *testing.T, r *Router) (jwt, user string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/user", nil)
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)
	var body struct {
		Jwt  string `json:"jwt"`
		User string `json:"user"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.NotEmpty(t, body.Jwt)
	return body.Jwt, body.User
}

func multipartBody(t *testing.T, fields map[string][]byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for name, data := range fields {
		part, err := w.CreateFormFile(name, name+".bin")
		require.NoError(t, err)
		_, err = part.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return body, w.FormDataContentType()
}

func authReq(t *testing.T, method, url, jwt string, body io.Reader, contentType string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Authorization", "Bearer "+jwt)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req
}

func TestCreateTaskUnauthorized(t *testing.T) {
	r, _, _ := testApp(t)
	xlsx, err := sampledata.XLSXBytes(2)
	require.NoError(t, err)
	body, ctype := multipartBody(t, map[string][]byte{"exel_file": xlsx})
	req := httptest.NewRequest(http.MethodPost, "/api/create_task", body)
	req.Header.Set("Content-Type", ctype)
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCreateTaskMissingExcel(t *testing.T) {
	r, _, _ := testApp(t)
	jwt, _ := createSession(t, r)
	docx, err := sampledata.DOCXBytes()
	require.NoError(t, err)
	body, ctype := multipartBody(t, map[string][]byte{"doc_file": docx})
	req := authReq(t, http.MethodPost, "/api/create_task", jwt, body, ctype)
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "exel file")
}

func TestXlsxInfoInvalidFile(t *testing.T) {
	r, _, _ := testApp(t)
	jwt, _ := createSession(t, r)
	body, ctype := multipartBody(t, map[string][]byte{"exel_file": []byte("not-an-xlsx")})
	req := authReq(t, http.MethodPost, "/api/xlsx_info", jwt, body, ctype)
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRunTaskErrors(t *testing.T) {
	r, _, _ := testApp(t)
	jwt, _ := createSession(t, r)

	t.Run("missing task", func(t *testing.T) {
		req := authReq(t, http.MethodPost, "/api/run_task?task_id=missing&sheet_name=Sheet1&use_first_row_as_columns=true&min_row=2&max_row=3", jwt, nil, "")
		rr := httptest.NewRecorder()
		r.router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	xlsx, err := sampledata.XLSXBytes(3)
	require.NoError(t, err)
	docx, err := sampledata.DOCXBytes()
	require.NoError(t, err)
	body, ctype := multipartBody(t, map[string][]byte{"exel_file": xlsx, "doc_file": docx})
	req := authReq(t, http.MethodPost, "/api/create_task", jwt, body, ctype)
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var created struct {
		TaskID string `json:"task_id"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))

	t.Run("missing sheet", func(t *testing.T) {
		req := authReq(t, http.MethodPost, "/api/run_task?task_id="+created.TaskID+"&use_first_row_as_columns=true&min_row=2&max_row=3", jwt, nil, "")
		rr := httptest.NewRecorder()
		r.router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "Sheet name")
	})

	t.Run("too many rows", func(t *testing.T) {
		q := fmt.Sprintf("/api/run_task?task_id=%s&sheet_name=Sheet1&use_first_row_as_columns=true&min_row=1&max_row=101", created.TaskID)
		req := authReq(t, http.MethodPost, q, jwt, nil, "")
		rr := httptest.NewRecorder()
		r.router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "max 100")
	})

	t.Run("min greater than max", func(t *testing.T) {
		q := fmt.Sprintf("/api/run_task?task_id=%s&sheet_name=Sheet1&use_first_row_as_columns=true&min_row=5&max_row=2", created.TaskID)
		req := authReq(t, http.MethodPost, q, jwt, nil, "")
		rr := httptest.NewRecorder()
		r.router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestGenerateAndDownloadZip(t *testing.T) {
	r, st, _ := testApp(t)
	jwt, _ := createSession(t, r)
	xlsx, err := sampledata.XLSXBytes(3)
	require.NoError(t, err)
	docx, err := sampledata.DOCXBytes()
	require.NoError(t, err)
	body, ctype := multipartBody(t, map[string][]byte{"exel_file": xlsx, "doc_file": docx})
	req := authReq(t, http.MethodPost, "/api/create_task", jwt, body, ctype)
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var created struct {
		TaskID string `json:"task_id"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))

	q := fmt.Sprintf("/api/run_task?task_id=%s&sheet_name=%s&use_first_row_as_columns=true&min_row=2&max_row=4", created.TaskID, sampledata.SheetName)
	req = authReq(t, http.MethodPost, q, jwt, nil, "")
	rr = httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	var run struct {
		DocHashes []string `json:"doc_hashes"`
		TotalDocs int      `json:"total_docs"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &run))
	require.Equal(t, 3, run.TotalDocs)
	require.Len(t, run.DocHashes, 3)

	data, err := st.GetAnyDocData(run.DocHashes[0])
	require.NoError(t, err)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	var xml string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			require.NoError(t, err)
			b, err := io.ReadAll(rc)
			rc.Close()
			require.NoError(t, err)
			xml = string(b)
		}
	}
	require.Contains(t, xml, "Demo-1")
	require.Contains(t, xml, "item-1.example")
	require.NotContains(t, xml, "{"+sampledata.ColName+"}")

	payload, err := json.Marshal(map[string][]string{"hashes": run.DocHashes})
	require.NoError(t, err)
	req = authReq(t, http.MethodPost, "/api/download_zip", jwt, bytes.NewReader(payload), "application/json")
	rr = httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "application/zip", rr.Header().Get("Content-Type"))
	require.Greater(t, rr.Body.Len(), 100)
}

func TestDownloadZipNoHashes(t *testing.T) {
	r, _, _ := testApp(t)
	jwt, _ := createSession(t, r)
	req := authReq(t, http.MethodPost, "/api/download_zip", jwt, strings.NewReader(`{"hashes":[]}`), "application/json")
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDownloadZipForeignHash(t *testing.T) {
	r, _, _ := testApp(t)
	jwt, _ := createSession(t, r)
	req := authReq(t, http.MethodPost, "/api/download_zip", jwt, strings.NewReader(`{"hashes":["deadbeef"]}`), "application/json")
	rr := httptest.NewRecorder()
	r.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}
