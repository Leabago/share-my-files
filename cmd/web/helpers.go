package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func getEnv(name string, logger AppLogger) string {
	varEnv := os.Getenv(name)
	if varEnv == "" {
		ErrDuplicateEmail := fmt.Errorf("empty environment variable %s", name)
		logger.errorLog.Fatal(ErrDuplicateEmail)
	}
	return varEnv
}

func createFolderForFiles(folderPath string, logger AppLogger) {
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

// removeNotAvaliableFiles get keys from redis and delete expired files
func (app *application) removeNotAvaliableFiles() {
	f, err := os.Open(folderPath)
	if err != nil {
		app.logger.errorLog.Fatal(err)
		return
	}
	files, err := f.Readdir(0)
	if err != nil {
		app.logger.errorLog.Fatal(err)
		return
	}

	fileNamesForDelete := make(map[string]bool) // New empty set

	for _, v := range files {
		fileNamesForDelete[v.Name()] = true // Add
	}

	// get file name for saving
	var availableFiles = app.getAvailableFiles()

	// delete Available file from deleting list
	for _, name := range availableFiles {
		delete(fileNamesForDelete, name)
	}

	// delete file
	for k := range fileNamesForDelete { // Loop
		e := os.Remove(folderPath + k)
		if e != nil {
			log.Fatal(e)
		}
	}

	// app.logger.infoLog.Println("files to be deleted: ", fileNamesForDelete)
}

func (app *application) deleteFileEveryNsec(second time.Duration) {
	ticker := time.NewTicker(time.Second * second)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		app.removeNotAvaliableFiles()
	}
}

// getAvailableFiles returen all folders(key) wich are not expire
func (app *application) getAvailableFiles() []string {
	// get all keys from redis
	result, _ := app.redisClient.Keys("available:*").Result()

	var availableFiles []string
	for _, s := range result {
		split := strings.Split(s, ":")
		filename := folderBegin + split[1] + zipName
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
	fmt.Println("ParseMediaType: error6: ", err)
	return nil, err

}

func writeFileSize(logger AppLogger) int {
	fileName := configFolderPath + maxFileSizeFileName
	//...................................
	//Writing struct type to a JSON file
	//...................................

	// if file not exist, then create it with default values
	if _, err := os.Stat(fileName); err != nil {
		// fill deafault maxFileSize
		var data []byte = []byte("var maxFileSize = " + strconv.Itoa(maxFileSize) + ";")

		err = os.WriteFile(fileName, data, 0644)
		if err != nil {
			logger.errorLog.Fatal(err)
		}

		return maxFileSize
	} else {
		// read from existing file

		// read json file
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
		return i
	}
}
