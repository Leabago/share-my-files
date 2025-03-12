package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"share-my-files/pkg/forms"
	"share-my-files/pkg/models"
	"strconv"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok"))
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {

	// Generate a new session ID
	sessionID, err := generateSessionID()
	if err != nil {
		app.logger.errorLog.Fatal(err)
		return
	}

	// create unique user code assosiated with files
	userCode := createUserCode()

	// Store the session in-memory store Redis
	app.redisClient.Set(app.getRedisPath(sessionPath, sessionID), userCode, sessionIdTime)

	// Create a secure cookie
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true, // Prevent access via JavaScript
		// Secure:   true, // Use HTTPS in production
		// for developnet, DELETE in production
		Secure:   false, // Use HTTPS in production
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // Prevent CSRF attacks
	}

	// Set the cookie in the response
	http.SetCookie(w, &cookie)

	app.render(w, r, "create.page.tmpl.html", &templateData{
		Form:        forms.New(nil),
		SessionCode: userCode,
	})
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get(":id")

	fileInfo := app.redisClient.HGet(app.getRedisPath(availablePath, code), fileInfoTitle).Val()
	file := &models.File{}
	json.Unmarshal([]byte(fileInfo), file)
	file.Exist = app.fileExist(code) && (fileInfo != "")
	file.FileCode = code

	app.render(w, r, "show.page.tmpl.html", &templateData{
		File: file,
	})
}

func (app *application) getSnippet(w http.ResponseWriter, r *http.Request) {
	fileCode := r.URL.Query().Get(":id")

	fileInfo := app.redisClient.HGet(app.getRedisPath(availablePath, fileCode), fileInfoTitle).Val()
	file := &models.File{}
	json.Unmarshal([]byte(fileInfo), file)

	if app.fileExist(fileCode) {
		filePath := folderPath + file.Name
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(file.Name))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, filePath)

		// delete file if it is for one download
		if file.OneTimeDownload {
			app.redisClient.HDel(app.getRedisPath(availablePath, fileCode), fileInfoTitle)
		}
	} else {
		http.Redirect(w, r, fmt.Sprintf("/archive/%s", fileCode), http.StatusSeeOther)
	}
}

// homeGetFiles upload files to zip
func (app *application) homeGetFiles(w http.ResponseWriter, r *http.Request) {

	sessionID, userCode, err := app.getSessionValue(r)
	if err != nil {
		app.serverError(w, err)
	}

	var folderPathFull = filepath.Join(folderPath, sessionID)
	app.logger.infoLog.Printf("create new folder %s", folderPathFull)

	_, err = saveFilesToFolder(r, folderPathFull, app.maxFileSize)
	if err != nil {
		if errors.Is(err, errFileTooLarge) {
			app.fileTooLarge(w, err)
		} else {
			app.serverError(w, err)
		}
		return
	}

	w.Write([]byte(userCode))
}

func (app *application) redirectToArchive(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	title := form.Values.Get("title")
	app.logger.infoLog.Printf("redirectToArchive %s", title)

}

func (app *application) createDownloadForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "download.page.tmpl.html", &templateData{})
}

func (application *application) redirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/upload", http.StatusSeeOther)
}

// archive upload files to zip
func (app *application) archive(w http.ResponseWriter, r *http.Request) {
	sessionID, fileCode, err := app.getSessionValue(r)
	if err != nil {
		app.serverError(w, err)
	}

	// create zip archive with files
	fileNameList, err := createArhive(sessionID, fileCode)
	if err != nil {
		app.serverError(w, err)
	}

	// select file lifetime by selected radio value
	oneTimeDownload, lifeTime := selectLifeTime(r.FormValue("storageDuration"))

	fmt.Println("archive lifeTime: ", lifeTime)

	// collect file information
	fullURL := getFullURL(r, fileCode)

	base64ImageData, err := createBase64ImageData(fullURL)
	if err != nil {
		app.serverError(w, err)
	}

	file := &models.File{
		Name:            folderBegin + fileCode + zipName,
		FileCode:        fileCode,
		FileNameList:    fileNameList,
		OneTimeDownload: oneTimeDownload,
		Exist:           true,
		URL:             fullURL,
		QRcodeBase64:    base64ImageData,
		LifeTime:        lifeTime.String(),
	}

	fileJson, err := json.Marshal(file)
	if err != nil {
		app.serverError(w, err)
	}

	app.redisClient.HSet((app.getRedisPath(availablePath, fileCode)), fileInfoTitle, string(fileJson))
	app.redisClient.Expire(app.getRedisPath(availablePath, fileCode), lifeTime).Result()

	w.Write([]byte(fileCode))
}

// deleteOneFile delete only one file during session
func (app *application) deleteOneFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get(":name")
	sessionID, _, err := app.getSessionValue(r)
	if err != nil {
		app.serverError(w, err)
	}
	fullPath := filepath.Join(folderPath, sessionID, fileName)
	err = os.Remove(fullPath)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) getUserCode(w http.ResponseWriter, r *http.Request) {
	_, fileCode, err := app.getSessionValue(r)
	if err != nil {
		app.serverError(w, err)
	}

	w.Write([]byte(fileCode))
}
