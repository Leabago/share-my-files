package main

import "time"

const pi = 3.14
const folderPath = "/tmp/share-my-file/"

// const configFolderPath = "/tmp/share-my-file-config/"
const configFolderPath = "./ui/static/js/"
const folderBegin = "share-my-file-"
const zipName = ".zip"
const available = "available:"

const smallTime = 1 * time.Minute
const mediumTime = 1 * time.Hour
const bigTime = 12 * time.Hour
const maxFileSize = 104857600
const maxFileSizeFileName = "max-file-size.js"
const maxFileSizeRegex = `^var maxFileSize = (\d*);$`
