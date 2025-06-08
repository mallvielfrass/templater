package router

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	exdocconverter "github.com/mallvielfrass/templater/internal/exdocConverter"
	exelreader "github.com/mallvielfrass/templater/internal/exelReader"
	"github.com/mallvielfrass/templater/internal/models"
)

func (root *Router) CreateTask(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	// Разбор формы multipart/form-data
	err := req.ParseMultipartForm(100 << 20) // 100 MB максимальный размер файла
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	//get exel file from multipart form
	exelFile, exelHandler, err := req.FormFile("exel_file")
	if err != nil {
		http.Error(w, "Error retrieving the exel file", http.StatusBadRequest)
		return
	}
	defer exelFile.Close()
	//get doc file info
	docFile, docHandler, err := req.FormFile("doc_file")
	if err != nil {
		http.Error(w, "Error retrieving the doc file", http.StatusBadRequest)
		return
	}
	defer docFile.Close()

	//create task
	taskID, err := root.taskStorage.CreateTask(userID)
	if err != nil {
		http.Error(w, "Error creating the task", http.StatusInternalServerError)
		return
	}
	//save exel file
	exelData, err := io.ReadAll(exelFile)
	if err != nil {
		http.Error(w, "Error reading the exel file", http.StatusInternalServerError)
		return
	}
	hash, err := root.fileStorage.SaveExelFile(exelHandler.Filename, exelData)
	if err != nil {
		http.Error(w, "Error saving the exel file", http.StatusInternalServerError)
		return
	}
	//save doc file
	docData, err := io.ReadAll(docFile)
	if err != nil {
		http.Error(w, "Error reading the doc file", http.StatusInternalServerError)
		return
	}
	docHash, err := root.fileStorage.SaveDocFile(docHandler.Filename, docData)
	if err != nil {
		http.Error(w, "Error saving the doc file", http.StatusInternalServerError)
		return
	}
	//add exel file hash to task
	err = root.taskStorage.AddExelHash(taskID.ID, hash)
	if err != nil {
		http.Error(w, "Error adding the exel file hash to the task", http.StatusInternalServerError)
		return
	}
	//add doc file hash to task
	err = root.taskStorage.AddDocHash(taskID.ID, docHash)
	if err != nil {
		http.Error(w, "Error adding the doc file hash to the task", http.StatusInternalServerError)
		return
	}
	//get list of tables from exel file
	exelInfo, err := root.fileStorage.GetExelFileInfo(hash)
	if err != nil {
		http.Error(w, "Error getting the exel file info", http.StatusInternalServerError)
		return
	}
	type Response struct {
		TaskID   string          `json:"task_id"`
		ExelInfo models.FileInfo `json:"exel_info"`
		DocInfo  models.FileInfo `json:"doc_info"`
	}

	response := Response{
		TaskID:   taskID.ID,
		ExelInfo: exelInfo,
		//DocInfo:  docInfo,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
	return
}

// Run task
func (root *Router) RunTask(w http.ResponseWriter, req *http.Request) {
	taskID := req.Header.Get("task_id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	//get sheet name from query params
	sheetName := req.URL.Query().Get("sheet_name")
	if sheetName == "" {
		http.Error(w, "Sheet name is required", http.StatusBadRequest)
		return
	}
	//get param "use_first_row_as_columns"
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
	//get min and max row from query params
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

	//get task
	task, err := root.taskStorage.GetTask(taskID)
	if err != nil {
		http.Error(w, "Error getting the task", http.StatusInternalServerError)
		return
	}
	//get exel file
	exelData, err := root.fileStorage.GetExelFileData(task.ExelHash)
	if err != nil {
		http.Error(w, "Error getting the exel file", http.StatusInternalServerError)
		return
	}
	//read exel file
	openExel, err := exelreader.ReadBuffer(task.ExelHash, exelData)
	if err != nil {
		http.Error(w, "Error reading the exel file", http.StatusInternalServerError)
		return
	}
	//get doc file
	docData, err := root.fileStorage.GetDocFileData(task.DocHash)
	if err != nil {
		http.Error(w, "Error getting the doc file", http.StatusInternalServerError)
		return
	}

	// Создаем ExDocConverter
	converter, err := exdocconverter.NewExDocConverter(root.fileStorage, &openExel, docData)
	if err != nil {
		http.Error(w, "Error creating converter", http.StatusInternalServerError)
		return
	}

	// Создаем опции конвертации
	options, err := converter.CreateConvertOptions(sheetName, useFirstRowAsColumnsBool)
	if err != nil {
		http.Error(w, "Error creating convert options", http.StatusInternalServerError)
		return
	}

	// Конвертируем документы
	docHashes, err := converter.Convert(options, minRowInt, maxRowInt)
	if err != nil {
		http.Error(w, "Error converting documents", http.StatusInternalServerError)
		return
	}

	// Возвращаем результат
	response := map[string]interface{}{
		"task_id":    taskID,
		"doc_hashes": docHashes,
		"total_docs": len(docHashes),
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
}
