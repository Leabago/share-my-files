package main

import (
	"fmt"
	"time"
)

const pi = 3.14

// port for aplication
const APP_PORT = ":8080"

// folderPath folder with user files
const folderPath = "/tmp/share-my-files/"

const configFolderPath = "/ui/static/js/"
const folderBegin = "share-my-files-"
const zipName = ".zip"

// availablePath path to redis keys where user keys is stored
const availablePath = "available:"

// sessionPath path to redis keys where sessions is stored
const sessionPath = "session:"

// session_id key for cookie
const session_id = "session_id"

const fileInfoTitle = "fileInfo"

const smallTime = 2 * time.Minute
const mediumTime = 1 * time.Hour
const bigTime = 12 * time.Hour

// sessionIdTime lifetime for session
const sessionIdTime = 30 * time.Minute

// 100 megabytes = = 104857600 bytes
// maxFileSize - maximum file size
// const maxFileSize = int64(10)

// file for keep maxFileSize variable
const maxFileSizeFileName = "max-file-size.js"

// file for keep ddns address
const ddnsAddressFileName = "ddns_address.js"

// errors
const bigFileMessage = "File size is too large, no more than %d megabytes allowed" // maxFileSize

var errFileTooLarge error = fmt.Errorf("file size limit exceeded")
