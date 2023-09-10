package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"path/filepath"
	"time"

	"github.com/akamensky/argparse"
	"golang.org/x/exp/slices"
)

type Web struct {
	directory string
	name      string
	maxSize   int64
	methods   []string
}

func (web *Web) uploadFile(w http.ResponseWriter, r *http.Request) {
	if !slices.Contains(web.methods, r.Method) {
		fmt.Printf("Upload endpoint requested using a unknown method '%s'\n", r.Method)
		return
	}
	fmt.Println("File upload requested")
	r.ParseMultipartForm(web.maxSize << 20)
	file, handler, err := r.FormFile(web.name)
	if err != nil {
		fmt.Printf("0x01 - Error Retrieving the multipart parameter '%s'\n", web.name)
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x01\n")
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	path := filepath.Join(web.directory, fmt.Sprintf("%d.%s", time.Now().Unix(), handler.Filename))

	//Check for path traversal
	if filepath.Dir(path) != filepath.Dir(web.directory) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(filepath.Dir(path))
		fmt.Println(filepath.Dir(web.directory))
		fmt.Fprintf(w, "ERR - 0x02\n")
		return
	}

	// Create a  file within our directory
	tFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x03\n")
		return
	}
	defer tFile.Close()

	// read all of the contents of our uploaded file into a byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x04\n")
		return
	}
	// write this byte array to our file
	tFile.Write(fileBytes)
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "OK\n")
}

func main() {
	parser := argparse.NewParser("deaddrop", "Starts a http server, which can recieve file uploads from wget and curl")
	port := parser.Int("i", "port", &argparse.Options{Default: 5050, Help: "Port used to recieved HTTP traffic"})
	directory := parser.String("d", "directory", &argparse.Options{Default: "/var/deaddrop/", Help: "Location where the files are stored"})
	name := parser.String("n", "name", &argparse.Options{Default: "file", Help: "multipart parametername for the file"})
	methods := parser.StringList("m", "methods", &argparse.Options{Default: []string{"POST"}, Help: "Sets the accepted HTTP method"})
	size := parser.Int("s", "size", &argparse.Options{Default: 100, Help: "Maximal file size allowed to be uploaded in Mb"})
	err := parser.Parse(os.Args)

	web := &Web{directory: *directory, name: *name, maxSize: int64(*size), methods: *methods}

	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	if *size <= 0 {
		fmt.Println("maximal file size can not be 0 or smaller")
		return
	}

	//Create upload folder if not exists
	if _, err := os.Stat(*directory); os.IsNotExist(err) {
		err = os.Mkdir(*directory, 0766)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Printf("Starting HTTP server on: %d\n", *port)
	http.HandleFunc("/", web.uploadFile)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
