package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"path/filepath"
	"regexp"
	"strings"
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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9\. ]+`)

func (web *Web) uploadFileGeneric(w http.ResponseWriter, r *http.Request) {
	if !slices.Contains(web.methods, r.Method) {
		fmt.Printf("Upload endpoint requested using a unknown method '%s'\n", r.Method)
		return
	}

	fmt.Println("File upload requested")

	fmt.Println("Headers:")
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Printf("%v: %v\n", name, h)
		}
	}
	fmt.Println("----------------------")

	if r.Header.Get("content-type") == "application/x-www-form-urlencoded" {
		fmt.Println("Detected binary stream")
		web.uploadFileBinary(w, r)
	} else if strings.HasPrefix(r.Header.Get("content-type"), "Content-Type: multipart/form-data") {
		fmt.Println("Detected multipart upload")
		web.uploadFileMultipart(w, r)
	} else {
		fmt.Println("Detected pure binary upload")
		web.uploadFileBinary(w, r)
	}
	fmt.Println("File upload request closed")
}

func (web *Web) uploadFileBinary(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(web.directory, fmt.Sprintf("%d.blob", time.Now().Unix()))
	// Create a  file within our directory
	tFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x13\n")
		return
	}
	defer tFile.Close()

	// read all of the contents of our uploaded file
	_, err = io.Copy(tFile, r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x14\n")
		return
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "OK\n")
}

func (web *Web) uploadFileMultipart(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(web.maxSize << 20)
	file, handler, err := r.FormFile(web.name)
	if err != nil {
		fmt.Printf("0x01 - Error retrieving the multipart parameter '%s'\n", web.name)
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x01\n")
		return
	}
	defer file.Close()

	cleaned_filename := nonAlphanumericRegex.ReplaceAllString(handler.Filename, "_")

	fmt.Printf("Uploaded file: %+v\n", cleaned_filename)
	fmt.Printf("File size: %+v\n", handler.Size)
	fmt.Printf("MIME header: %+v\n", handler.Header)

	path := filepath.Join(web.directory, fmt.Sprintf("%d.%s", time.Now().Unix(), cleaned_filename))

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
	_, err = io.Copy(tFile, file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x04\n")
		return
	}
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "OK\n")
}

func main() {
	parser := argparse.NewParser("deaddrop", "Starts a HTTP server, which can recieve file uploads from curl")
	port := parser.Int("i", "port", &argparse.Options{Default: 5050, Help: "Port used to recieve HTTP traffic"})
	directory := parser.String("d", "directory", &argparse.Options{Default: "/var/deaddrop/", Help: "Location where the uploaded files are stored"})
	name := parser.String("n", "name", &argparse.Options{Default: "file", Help: "Multipart parameter name for the file"})
	methods := parser.StringList("m", "methods", &argparse.Options{Default: []string{"POST", "PUT"}, Help: "Sets the accepted HTTP method(s)"})
	size := parser.Int("s", "size", &argparse.Options{Default: 2048, Help: "Maximal file size allowed to be uploaded in Mb"})
	err := parser.Parse(os.Args)

	web := &Web{directory: *directory, name: *name, maxSize: int64(*size), methods: *methods}

	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	if *size <= 0 {
		fmt.Println("Maximal file size can not be 0 or smaller")
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
	http.HandleFunc("/", web.uploadFileGeneric)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
