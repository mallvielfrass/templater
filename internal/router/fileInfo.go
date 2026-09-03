package router

import (
	"encoding/json"
	"net/http"

	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/models"
)

func (root *Router) SheetInfo(w http.ResponseWriter, req *http.Request) {
	taskID := req.Header.Get("task_id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	sheetName := req.Header.Get("sheet_name")
	if sheetName == "" {
		http.Error(w, "Sheet name is required", http.StatusBadRequest)
		return
	}
	// get task
	task, err := root.taskStorage.GetTask(taskID)
	if err != nil {
		http.Error(w, "Error getting the task", http.StatusInternalServerError)
		return
	}
	// get exel file
	exelData, err := root.fileStorage.GetExelFileData(task.ExelHash)
	if err != nil {
		http.Error(w, "Error getting the exel file", http.StatusInternalServerError)
		return
	}
	// read exel file
	exelReader, err := exelreader.ReadBuffer(task.ExelHash, exelData)
	if err != nil {
		http.Error(w, "Error reading the exel file", http.StatusInternalServerError)
		return
	}
	//get sheet
	sheet, err := exelReader.SheetInfo(sheetName)
	if err != nil {
		http.Error(w, "Error getting the sheet", http.StatusInternalServerError)
		return
	}
	type Response struct {
		Sheet models.Sheet `json:"sheet"`
	}
	response := Response{
		Sheet: sheet,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding the response", http.StatusInternalServerError)
		return
	}
	return
}
