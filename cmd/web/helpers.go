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
	"runtime/debug"
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

func createFolderForFiles(logger AppLogger) {
	if err := os.Mkdir(folderPath, os.ModePerm); err != nil {
		if errors.Is(err, os.ErrExist) {
			// file exist
		} else {
			logger.errorLog.Fatal(err)
		}
	}
}

func (app *application) createFoldeWithCoderForFiles(code string) string {
	var path = folderPath + folderBegin + code + zipName
	// if err := os.Mkdir(path, os.ModePerm); err != nil {
	// 	if errors.Is(err, os.ErrExist) {
	// 		// file exist
	// 	} else {
	// 		app.logger.errorLog.Fatal(err)
	// 	}
	// }

	return path
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

func (app *application) getAllFilesFromFolder() {
	f, err := os.Open(folderPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	files, err := f.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return
	}

	var asd string
	asd = "asd"

	fileNamesForDelete := make(map[string]bool) // New empty set

	// for k := range set {         // Loop
	// 	fmt.Println(k)
	// }
	// delete(set, "Foo")    // Delete
	// size := len(set)      // Size
	// exists := set["Foo"]  // Membership

	fmt.Print(asd)

	for _, v := range files {
		fmt.Println(v.Name(), v.IsDir())
		fileNamesForDelete[v.Name()] = true // Add
	}

	// get file name for saving
	var availableFiles = app.getAvailableFiles()

	// delete Available file from deleting list
	for i, name := range availableFiles {
		fmt.Println(i, name)
		delete(fileNamesForDelete, name)
	}

	fmt.Println("fileNamesForDelete after:", fileNamesForDelete)

	// delete file
	for k := range fileNamesForDelete { // Loop
		fmt.Println(k)
		e := os.Remove(folderPath + k)
		if e != nil {
			log.Fatal(e)
		}
	}

}

func (app *application) getAvailableFiles() []string {

	result, _ := app.redisClient.Keys("available:*").Result()

	var availableFiles []string

	fmt.Println("availableFiles:")
	for i, s := range result {
		// fmt.Println(i, s)

		split := strings.Split(s, ":")
		fmt.Println(i, s)
		fmt.Println(split)
		filename := folderBegin + split[1] + zipName
		availableFiles = append(availableFiles, filename)
	}

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

func createUserCode() string {
	rand.Seed(time.Now().UnixNano())
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

func ParseMediaType(r *http.Request, zipFileName string) error {

	file, err := os.Create(zipFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new zip writer
	wr := zip.NewWriter(file)

	defer wr.Close()

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}
	if mediaType == "multipart/form-data" {
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				fmt.Println(err)
				log.Fatal(err)

			}
			slurp, err := io.ReadAll(p)
			if err != nil {
				log.Fatal(err)
			}
			// fmt.Printf("Part %q: %q\n", p.Header.Get("Foo"), slurp)

			// fmt.Printf("FileName %s: \n", p.FileName())

			// path := folder + p.FileName()
			// err = os.WriteFile(path, slurp, 0644)

			// Add a file to the zip file
			f, err := wr.Create(p.FileName())
			if err != nil {
				log.Fatal(err)
			}

			// Write data to the file
			_, err = f.Write(slurp)
			if err != nil {
				log.Fatal(err)
			}

			check(err)

		}
	}

	return nil
}
