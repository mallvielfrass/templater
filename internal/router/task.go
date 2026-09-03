package router

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	docreader "github.com/mallvielfrass/templater/internal/docReader"
	exdocconverter "github.com/mallvielfrass/templater/internal/exdocConverter"
	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func taskIDFrom(req *http.Request) string {
	id := req.URL.Query().Get("task_id")
	if id == "" {
		id = req.Header.Get("task_id")
	}
	return id
}

func (root *Router) ownedTask(req *http.Request, taskID string) (models.Task, bool) {
	if taskID == "" {
		return models.Task{}, false
	}
	task, err := root.taskStorage.GetTask(taskID)
	if err != nil || task.UserID != userFrom(req) {
		return models.Task{}, false
	}
	return task, true
}

func (root *Router) CreateTask(w http.ResponseWriter, req *http.Request) {
	userID := userFrom(req)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	err := req.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	exelFile, exelHandler, err := req.FormFile("exel_file")
	if err != nil {
		http.Error(w, "Error retrieving the exel file", http.StatusBadRequest)
		return
	}
	defer exelFile.Close()
	docFile, docHandler, err := req.FormFile("doc_file")
	var docData []byte
	docName := "document.docx"
	if errors.Is(err, http.ErrMissingFile) {
		docData, err = docreader.EmptyBytes()
		if err != nil {
			http.Error(w, "Error creating the doc file", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Error retrieving the doc file", http.StatusBadRequest)
		return
	} else {
		defer docFile.Close()
		docData, err = io.ReadAll(docFile)
		if err != nil {
			http.Error(w, "Error reading the doc file", http.StatusInternalServerError)
			return
		}
		if docHandler.Filename != "" {
			docName = docHandler.Filename
		}
	}

	task, err := root.taskStorage.CreateTask(userID)
	if err != nil {
		http.Error(w, "Error creating the task", http.StatusInternalServerError)
		return
	}
	exelData, err := io.ReadAll(exelFile)
	if err != nil {
		http.Error(w, "Error reading the exel file", http.StatusInternalServerError)
		return
	}
	hash, err := root.fileStorage.SaveExelFile(exelHandler.Filename, exelData)
	if err != nil {
		log.Printf("SaveExelFile: %v", err)
		http.Error(w, "Error saving the exel file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := root.fileStorage.BindOwner(hash, userID); err != nil {
		http.Error(w, "Error binding the exel file", http.StatusInternalServerError)
		return
	}
	docHash, err := root.fileStorage.SaveDocFile(docName, docData)
	if err != nil {
		http.Error(w, "Error saving the doc file", http.StatusInternalServerError)
		return
	}
	if err := root.fileStorage.BindOwner(docHash, userID); err != nil {
		http.Error(w, "Error binding the doc file", http.StatusInternalServerError)
		return
	}
	err = root.taskStorage.AddExelHash(task.ID, hash)
	if err != nil {
		http.Error(w, "Error adding the exel file hash to the task", http.StatusInternalServerError)
		return
	}
	err = root.taskStorage.AddDocHash(task.ID, docHash)
	if err != nil {
		http.Error(w, "Error adding the doc file hash to the task", http.StatusInternalServerError)
		return
	}
	exelInfo, err := root.fileStorage.GetExelFileInfo(hash)
	if err != nil {
		http.Error(w, "Error getting the exel file info", http.StatusInternalServerError)
		return
	}
	docInfo, err := root.fileStorage.GetDocFileInfo(docHash)
	if err != nil {
		http.Error(w, "Error getting the doc file info", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task_id":   task.ID,
		"exel_info": exelInfo,
		"doc_info":  docInfo,
		"doc_hash":  docHash,
	})
}

func (root *Router) RunTask(w http.ResponseWriter, req *http.Request) {
	taskID := taskIDFrom(req)
	task, ok := root.ownedTask(req, taskID)
	if !ok {
		http.NotFound(w, req)
		return
	}
	sheetName := req.URL.Query().Get("sheet_name")
	if sheetName == "" {
		http.Error(w, "Sheet name is required", http.StatusBadRequest)
		return
	}
	useFirstRowAsColumns := req.URL.Query().Get("use_first_row_as_columns")
	if useFirstRowAsColumns == "" {
		http.Error(w, "Use first row as columns is required", http.StatusBadRequest)
		return
	}
	useFirstRowAsColumnsBool, err := strconv.ParseBool(useFirstRowAsColumns)
	if err != nil {
		http.Error(w, "Invalid use first row as columns", http.StatusBadRequest)
		return
	}
	minRow := req.URL.Query().Get("min_row")
	maxRow := req.URL.Query().Get("max_row")
	if minRow == "" || maxRow == "" {
		http.Error(w, "Min and max row are required", http.StatusBadRequest)
		return
	}
	minRowInt, err := strconv.Atoi(minRow)
	if err != nil {
		http.Error(w, "Invalid min row", http.StatusBadRequest)
		return
	}
	maxRowInt, err := strconv.Atoi(maxRow)
	if err != nil {
		http.Error(w, "Invalid max row", http.StatusBadRequest)
		return
	}
	if minRowInt > maxRowInt {
		http.Error(w, "Min row is greater than max row", http.StatusBadRequest)
		return
	}
	if minRowInt < 1 || maxRowInt < 1 {
		http.Error(w, "Min and max row must be greater than 0", http.StatusBadRequest)
		return
	}

	exelData, err := root.fileStorage.GetExelFileData(task.ExelHash)
	if err != nil {
		http.Error(w, "Error getting the exel file", http.StatusInternalServerError)
		return
	}
	openExel, err := exelreader.ReadBuffer(task.ExelHash, exelData)
	if err != nil {
		http.Error(w, "Error reading the exel file", http.StatusInternalServerError)
		return
	}

	_ = root.forceSaveTaskDocument(task.ID)
	if updatedTask, err := root.taskStorage.GetTask(task.ID); err == nil {
		task = updatedTask
	}

	docData, err := root.fileStorage.GetDocFileData(task.DocHash)
	if err != nil {
		http.Error(w, "Error getting the doc file", http.StatusInternalServerError)
		return
	}

	converter, err := exdocconverter.NewExDocConverter(root.fileStorage, &openExel, docData)
	if err != nil {
		http.Error(w, "Error creating converter", http.StatusInternalServerError)
		return
	}

	options, err := converter.CreateConvertOptions(sheetName, useFirstRowAsColumnsBool)
	if err != nil {
		http.Error(w, "Error creating convert options", http.StatusInternalServerError)
		return
	}

	docHashes, err := converter.Convert(options, minRowInt, maxRowInt)
	if err != nil {
		http.Error(w, "Error converting documents", http.StatusInternalServerError)
		return
	}

	userID := userFrom(req)
	for _, h := range docHashes {
		if bindErr := root.fileStorage.BindOwner(h, userID); bindErr != nil {
			http.Error(w, "Error binding generated file", http.StatusInternalServerError)
			return
		}
	}
	_ = root.taskStorage.UpdateTaskStatus(task.ID, models.StatusCompleted)

	writeJSON(w, http.StatusOK, map[string]any{
		"task_id":    taskID,
		"doc_hashes": docHashes,
		"total_docs": len(docHashes),
	})
}

func (root *Router) Columns(w http.ResponseWriter, req *http.Request) {
	task, ok := root.ownedTask(req, taskIDFrom(req))
	if !ok {
		http.NotFound(w, req)
		return
	}
	sheetName := req.URL.Query().Get("sheet")
	if sheetName == "" {
		sheetName = req.URL.Query().Get("sheet_name")
	}
	if sheetName == "" {
		http.Error(w, "Sheet name is required", http.StatusBadRequest)
		return
	}
	exelData, err := root.fileStorage.GetExelFileData(task.ExelHash)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	exelReader, err := exelreader.ReadBuffer(task.ExelHash, exelData)
	if err != nil {
		http.Error(w, "Error reading the exel file", http.StatusInternalServerError)
		return
	}
	columns, err := exelReader.ReadFirstRow(sheetName)
	if err != nil {
		http.Error(w, "Error reading columns", http.StatusInternalServerError)
		return
	}
	out := make([]string, 0, len(columns))
	for _, c := range columns {
		name := strings.TrimSpace(c)
		if name != "" {
			out = append(out, name)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"columns": out, "sheet": sheetName})
}

func (root *Router) DownloadZip(w http.ResponseWriter, req *http.Request) {
	user := userFrom(req)
	if user == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Hashes []string `json:"hashes"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(body.Hashes) == 0 {
		http.Error(w, "No hashes provided", http.StatusBadRequest)
		return
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	added := 0
	for i, hash := range body.Hashes {
		owner, err := root.fileStorage.OwnerOf(hash)
		if err != nil || owner != user {
			continue
		}
		data, err := root.fileStorage.GetAnyDocData(hash)
		if err != nil {
			continue
		}
		shortHash := hash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}
		fileName := fmt.Sprintf("doc_%d_%s.docx", i+1, shortHash)
		f, err := zw.Create(fileName)
		if err != nil {
			http.Error(w, "Error creating zip entry", http.StatusInternalServerError)
			return
		}
		if _, err := f.Write(data); err != nil {
			http.Error(w, "Error writing zip entry", http.StatusInternalServerError)
			return
		}
		added++
	}

	if added == 0 {
		http.Error(w, "No accessible files found", http.StatusNotFound)
		return
	}

	if err := zw.Close(); err != nil {
		http.Error(w, "Error closing zip", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"documents.zip\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
