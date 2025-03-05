package main

import (
	"archive/zip"
	"bytes"
	crypto "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

func getEnv(name string, logger *AppLogger) string {
	value := os.Getenv(name)
	if value == "" {
		logger.errorLog.Fatal(fmt.Errorf("empty environment variable %s", name))
	}

	logger.infoLog.Printf("get environment variable: %s=%s", name, value)
	return value
}

func createFolderForFiles(folderPath string, logger *AppLogger) {
	if err := os.Mkdir(folderPath, os.ModePerm); err != nil {
		if errors.Is(err, os.ErrExist) {
			// file exist
		} else {
			logger.errorLog.Fatal(err)
		}
	}
}

func (app *application) fileExist(code string) bool {
	fileName := folderPath + folderBegin + code + zipName
	_, err := os.Open(fileName) // For read access.
	if err != nil {
		return false
	} else {
		return true
	}
}

// removeExpiredFiles get keys from redis and delete expired files
func (app *application) removeExpiredFiles() {
	f, err := os.Open(folderPath)
	if err != nil {
		app.logger.errorLog.Fatal(err)
		return
	}
	folders, err := f.Readdir(0)
	if err != nil {
		app.logger.errorLog.Fatal(err)
		return
	}

	foldersForDelete := make(map[string]bool) // New empty set

	for _, v := range folders {
		foldersForDelete[v.Name()] = true // Add
	}

	// get folders name for saving
	var availableFolders = app.getAvailableFolders()

	fmt.Println("availableFolders: ", availableFolders)

	// delete Available folders from deleting list
	for _, folder := range availableFolders {
		delete(foldersForDelete, folder)
	}

	app.logger.infoLog.Println("files to be deleted: ", foldersForDelete)

	// delete folders
	for k := range foldersForDelete {
		e := os.RemoveAll(folderPath + k)
		if e != nil {
			log.Fatal(e)
		}
	}
}

func (app *application) deleteFileEveryNsec(second time.Duration) {
	ticker := time.NewTicker(time.Second * second)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		app.removeExpiredFiles()
	}
}

// getAvailableFolders return all folders(key) wich are not expire
func (app *application) getAvailableFolders() []string {
	// get all keys from redis
	available, _ := app.redisClient.Keys(availablePath + "*").Result()
	session, _ := app.redisClient.Keys(sessionPath + "*").Result()
	var availableFiles []string

	for _, s := range available {
		split := strings.Split(s, ":")
		filename := folderBegin + split[1] + zipName
		availableFiles = append(availableFiles, filename)
	}

	for _, s := range session {
		split := strings.Split(s, ":")
		filename := split[1]
		availableFiles = append(availableFiles, filename)
	}

	// app.logger.infoLog.Println("all available folders:", availableFiles)
	return availableFiles
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// app.errorLog.Println(trace)
	app.logger.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) fileTooLarge(w http.ResponseWriter, err error) {
	app.logger.infoLog.Output(1, err.Error())
	http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", name))
		return
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our serverError helper and then
	// return.

	// Execute the template set, passing in any dynamic data.
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write the contents of the buffer to the http.ResponseWriter. Again, this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.
	buf.WriteTo(w)
}

// createUserCode create random code
func createUserCode() string {
	chars := []rune(
		"abcdefghijklmnopqrstuvwxyz" +
			"0123456789")
	length := 6
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // E.g. "ExcbsVQs"

	return str
}

func createZip() {
	file, err := os.Create("../new_zip_file.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new zip writer
	wr := zip.NewWriter(file)
	defer wr.Close()
}

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	return td
}

// ParseMediaType parse files from request and add it to zip arhive
func ParseMediaType(r *http.Request, zipFileName string, maxFileSize int) ([]string, error) {
	var fileNameList []string

	file, err := os.Create(zipFileName)
	if err != nil {
		fmt.Println("ParseMediaType: os.Create(zipFileName):", err)
		return nil, err
	}
	defer file.Close()

	// Create a new zip writer
	wr := zip.NewWriter(file)

	defer wr.Close()

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("ParseMediaType:  Content-Type missing: ", err)
		return nil, err
	}
	if mediaType == "multipart/form-data" {
		mr := multipart.NewReader(r.Body, params["boundary"])
		sumSize := 0

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				fmt.Println("ParseMediaType: ok: ", err)
				return fileNameList, nil
			}
			if err != nil {
				fmt.Println("ParseMediaType: error1: ", err)
				return nil, err
			}
			slurp, err := io.ReadAll(p)
			if err != nil {
				fmt.Println("ParseMediaType: error2: ", err)
				return nil, err
			}

			// Add a file to the zip file
			f, err := wr.Create(p.FileName())
			if err != nil {
				fmt.Println("ParseMediaType: error3: ", err)
				return nil, err
			}

			// Write data to the file
			sizeOfslurp, err := f.Write(slurp)
			if err != nil {
				fmt.Println("ParseMediaType: error4: ", err)
				return nil, err
			}

			sumSize += sizeOfslurp
			fileNameList = append(fileNameList, p.FileName())

			// if szie of files bigger then
			if sumSize > maxFileSize {
				var err error = errors.New(fmt.Sprintf(bigFileMessage, maxFileSize))
				fmt.Println("ParseMediaType: error5: ", err)
				return nil, err
			}

		}
	}

	err = errors.New("mus be multipart")
	fmt.Println("ParseMediaType: error: ", err)
	return nil, err

}

func saveFilesToFolder(r *http.Request, folderPath string, maxFileSize int) ([]string, error) {
	var fileNameList []string

	// Create the folder if it doesn't exist
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		fmt.Println("ParseMediaType: Error creating folder:", err)
		return nil, err
	}

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("ParseMediaType:  Content-Type missing: ", err)
		return nil, err
	}
	if mediaType == "multipart/form-data" {
		mr := multipart.NewReader(r.Body, params["boundary"])
		sumSize := 0

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				fmt.Println("ParseMediaType: ok: ", err)
				return fileNameList, nil
			}
			if err != nil {
				fmt.Println("ParseMediaType: error1: ", err)
				return nil, err
			}
			slurp, err := io.ReadAll(p)
			if err != nil {
				fmt.Println("ParseMediaType: error2: ", err)
				return nil, err
			}

			// Add a file to the folder

			err = os.WriteFile(filepath.Join(folderPath, p.FileName()), []byte(slurp), 0666)
			if err != nil {
				fmt.Println("Error writing file:", err)
				return nil, err
			}

			// Write data to the file
			// sizeOfslurp, err := f.Write(slurp)
			// if err != nil {
			// 	fmt.Println("ParseMediaType: error4: ", err)
			// 	return nil, err
			// }
			sizeOfslurp := 1

			sumSize += sizeOfslurp

			// if szie of files bigger then
			if sumSize > maxFileSize {
				fmt.Println("ParseMediaType: error5: ", fileTooLarge)
				return nil, fileTooLarge
			}

		}
	}

	err = errors.New("mus be multipart")
	fmt.Println("ParseMediaType: error: ", err)
	return nil, err
}

// writeFileSize create config file with max file size with size from constant, if file already existe then use value from existing file
func writeFileSize(logger *AppLogger) int {
	// get current directory
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	fileDir := filepath.Join(path, configFolderPath)
	fileName := filepath.Join(path, configFolderPath, maxFileSizeFileName)

	// if file not exist, then create it with default values
	if _, err := os.Stat(fileName); err != nil {
		logger.infoLog.Printf("create file '%s' with max file size %d bytes", fileName, maxFileSize)

		// fill deafault maxFileSize
		var data []byte = []byte("var maxFileSize = " + strconv.Itoa(maxFileSize) + ";")

		createFolderForFiles(fileDir, logger)

		err = os.WriteFile(fileName, data, 0644)
		if err != nil {
			logger.errorLog.Fatal(err)
		}

		return maxFileSize
	} else {
		// read from existing file

		// read javascript file
		file, err := os.Open(fileName)
		// if we os.Open returns an error then handle it
		if err != nil {
			logger.errorLog.Fatal(err)
		}
		fmt.Println("Successfully Opened ", fileName)
		// defer the closing of our jsonFile so that we can parse it later on
		defer file.Close()

		// read our opened jsonFile as a byte array.
		byteValue, _ := io.ReadAll(file)
		var stringFromFile = string(byteValue)
		re, _ := regexp.Compile(maxFileSizeRegex)
		// find size
		match := re.FindStringSubmatch(stringFromFile)

		i, err := strconv.Atoi(match[1])

		if err != nil {
			logger.errorLog.Fatal(err)
		}

		logger.infoLog.Printf("use existing file '%s' with max file size %d bytes", fileName, i)
		return i
	}
}

// generateSessionID generate a secure random session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 32 bytes for a secure session ID
	_, err := crypto.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// getSessionValue return session_id and code from redis
func (app *application) getSessionValue(r *http.Request) (string, string, error) {
	cookie, err := r.Cookie(session_id)
	if err != nil {
		if err == http.ErrNoCookie {
			// No cookie was found
			err := fmt.Errorf("no session cookie found: %v", err)
			app.logger.errorLog.Fatal(err)
			return "", "", err
		}
		// Other errors
		err := fmt.Errorf("error retrieving cookie: %v", err)
		app.logger.errorLog.Fatal(err)
		return "", "", err
	}

	// Get the session ID from the cookie
	sessionIdValue := cookie.Value

	result, err := app.redisClient.Keys(app.getRedisPath(sessionPath, sessionIdValue)).Result()
	if err != nil {
		app.logger.errorLog.Fatal(err)
		return "", "", err
	}

	if len(result) == 0 {
		err := fmt.Errorf("can`t find session id: %v", sessionIdValue)
		app.logger.infoLog.Println(err.Error())
		return "", "", err
	}
	// code - folder name
	sessionID := result[0]
	userCode := app.redisClient.Get(result[0]).Val()

	split := strings.Split(sessionID, ":")
	if len(split) != 2 {
		err = fmt.Errorf("can`t split session_id '%s'", sessionID)
		app.logger.errorLog.Fatal(err)
		return "", "", err
	}

	return split[1], userCode, nil
}

// createArhive creates archive puts all files into archive and delete all files except archive
func createArhive(sessionID, sessionValue string) ([]string, error) {

	var fileNameList []string

	// pathe to session folders
	var folderPathFull = filepath.Join(folderPath, sessionID)

	// Define the name of the ZIP archive
	var zipFileName = folderBegin + sessionValue + zipName
	zipFileNameFull := filepath.Join(folderPath, zipFileName)

	// Create the ZIP file
	zipFile, err := os.Create(zipFileNameFull)
	if err != nil {
		return fileNameList, fmt.Errorf("error creating ZIP file: %v", err)
	}
	defer zipFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Get the list of files in the current directory
	files, err := os.ReadDir(folderPathFull)
	fmt.Println("Get the list of files in the current directory: ", files)
	if err != nil {
		return fileNameList, fmt.Errorf("error reading directory: %s", err)
	}

	// Add each file to the ZIP archive
	for _, file := range files {
		// Skip directories and the ZIP file itself
		if file.IsDir() || file.Name() == zipFileName {
			continue
		}

		err := addFileToZip(zipWriter, folderPathFull, file.Name())
		if err != nil {
			return fileNameList, fmt.Errorf("error adding file to ZIP: %v", err)
		}

		fileNameList = append(fileNameList, file.Name())
	}

	// Close the ZIP writer to finalize the archive
	zipWriter.Close()
	zipFile.Close()

	// Delete all files except the ZIP archive

	err = os.RemoveAll(folderPathFull)
	if err != nil {
		return fileNameList, fmt.Errorf("error deleting folder '%s': %v", folderPathFull, err)
	}

	// for _, file := range files {

	// 	fmt.Println("1 delete file : ", file)
	// 	if file.Name() == zipFileName {
	// 		fmt.Println("2 delete zipFileName continue : ", file)
	// 		continue
	// 	}
	// 	err := os.Remove(filepath.Join(folderPathFull, file.Name()))
	// 	fmt.Println("err : ", err)
	// 	fmt.Println("2 delete file : ", filepath.Join(folderPathFull, file.Name()))
	// 	if err != nil {
	// 		return fileNameList, fmt.Errorf("error deleting file '%s': %v", file.Name(), err)
	// 	}
	// }

	fmt.Println("Backup completed. Archive saved as", zipFileNameFull)
	return fileNameList, nil
}

// Function to add a file to a ZIP archive
func addFileToZip(zipWriter *zip.Writer, folderPathFull, fileName string) error {
	// Open the file
	file, err := os.Open(filepath.Join(folderPathFull, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a ZIP file entry
	zipFileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	// Copy file content into ZIP
	_, err = io.Copy(zipFileWriter, file)
	if err != nil {
		return err
	}

	return nil
}

// selectLifeTime choose life time for file
func selectLifeTime(selectedOption string) time.Duration {
	switch selectedOption {
	case "1":
		return mediumTime
	case "2":
		return smallTime
	case "3":
		return mediumTime
	case "4":
		return bigTime
	}

	return smallTime
}

// isOneDownload return true if life time is one download
func isOneDownload(selectedOption string) bool {
	switch selectedOption {
	case "1":
		return true
	default:
		return false
	}
}

// return QR-code string
func createBase64ImageData(url string) (string, error) {
	var png []byte
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("cant create QRcode: %v", err)
	}

	base64ImageData := base64.StdEncoding.EncodeToString(png)

	return base64ImageData, nil
}

// getRedisPath return path for redis key
func (app *application) getRedisPath(path, key string) string {
	return path + key
}

// Generate full URL with HTTPS
func getFullURL(r *http.Request, fileCode string) string {

	// Check if the request is HTTPS
	if r.TLS != nil {
		return "https://" + filepath.Join(r.Host, r.RequestURI, fileCode)
	} else {
		return "http://" + filepath.Join(r.Host, r.RequestURI, fileCode)
	}
}

// writeDdnsAddress create config file with ddns address
func writeDdnsAddress(ddnsAddress string, logger *AppLogger) {
	// get current directory
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	fileDir := filepath.Join(path, configFolderPath)
	fileName := filepath.Join(path, configFolderPath, ddnsAddressFileName)

	// if file not exist, then create it with default values
	if _, err := os.Stat(fileName); err != nil {
		logger.infoLog.Printf("create file '%s' with ddns address: %s", fileName, ddnsAddress)

		// fill ddns address
		var data []byte = []byte("var ddnsAddress = '" + ddnsAddress + "';")

		createFolderForFiles(fileDir, logger)

		err = os.WriteFile(fileName, data, 0644)
		if err != nil {
			logger.errorLog.Fatal(err)
		}
	}
}
