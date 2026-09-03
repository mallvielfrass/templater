package router

import (
	"io"
	"net/http"

	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/models"
)

func (root *Router) XlsxInfo(w http.ResponseWriter, req *http.Request) {
	if userFrom(req) == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := req.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	exelFile, exelHandler, err := req.FormFile("exel_file")
	if err != nil {
		http.Error(w, "Error retrieving the exel file", http.StatusBadRequest)
		return
	}
	defer exelFile.Close()
	data, err := io.ReadAll(exelFile)
	if err != nil {
		http.Error(w, "Error reading the exel file", http.StatusInternalServerError)
		return
	}
	name := "book.xlsx"
	if exelHandler.Filename != "" {
		name = exelHandler.Filename
	}
	openExel, err := exelreader.ReadBuffer(name, data)
	if err != nil {
		http.Error(w, "Error reading the exel file: "+err.Error(), http.StatusBadRequest)
		return
	}
	info, err := openExel.FileInfo()
	if err != nil {
		http.Error(w, "Error getting the exel file info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sheets":    info.Sheets,
		"file_name": info.FileName,
	})
}

func (root *Router) SheetInfo(w http.ResponseWriter, req *http.Request) {
	taskID := taskIDFrom(req)
	task, ok := root.ownedTask(req, taskID)
	if !ok {
		http.NotFound(w, req)
		return
	}
	sheetName := req.URL.Query().Get("sheet_name")
	if sheetName == "" {
		sheetName = req.Header.Get("sheet_name")
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
	sheet, err := exelReader.SheetInfo(sheetName)
	if err != nil {
		http.Error(w, "Error getting the sheet", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Sheet models.Sheet `json:"sheet"`
	}{Sheet: sheet})
}
