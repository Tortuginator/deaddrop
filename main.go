package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/charmbracelet/log"
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
		log.Infof("Upload endpoint '%s' requested using a unknown method '%s'", r.URL, r.Method)
		return
	}

	log.Infof("Recieved potential upload request [%s] %s", r.Method, r.URL)
	log.Debug("Headers:")
	for name, headers := range r.Header {
		for _, h := range headers {
			log.Debugf("%v: %v\n", name, h)
		}
	}
	log.Debug("----------------------")
	var err error = nil
	if r.Header.Get("content-type") == "application/x-www-form-urlencoded" {
		log.Infof("Detected upload method: binary stream")
		err = web.uploadFileBinary(w, r)
	} else if strings.HasPrefix(r.Header.Get("content-type"), "Content-Type: multipart/form-data") {
		log.Infof("Detected upload method: multipart")
		err = web.uploadFileMultipart(w, r)
	} else {
		log.Infof("Fallback upload method: binary stream")
		err = web.uploadFileBinary(w, r)
	}
	if err == nil {
		log.Infof("Successfully completed upload request")
	} else {
		log.Error(err)
	}
}

func (web *Web) uploadFileBinary(w http.ResponseWriter, r *http.Request) error {
	path := filepath.Join(web.directory, fmt.Sprintf("%d.blob", time.Now().Unix()))
	// Create a  file within our directory
	tFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x13")
		return fmt.Errorf("(ERR - 0x13) Unable to create local file '%s' with error '%v'", path, err)
	}
	defer tFile.Close()

	// read all of the contents of our uploaded file
	_, err = io.Copy(tFile, r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x14")
		return fmt.Errorf("(ERR - 0x14) Failed to read request content with error '%v'", err)
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "OK\n")
	log.Infof("Written file to %s", path)
	return nil
}

func (web *Web) uploadFileMultipart(w http.ResponseWriter, r *http.Request) error {
	r.ParseMultipartForm(web.maxSize << 20)
	file, handler, err := r.FormFile(web.name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x01")
		return fmt.Errorf("(ERR - 0x13) Error retrieving the multipart parameter '%s' with error '%v'", web.name, err)
	}
	defer file.Close()

	cleaned_filename := nonAlphanumericRegex.ReplaceAllString(handler.Filename, "_")

	log.Debugf("Uploaded file: %+v", cleaned_filename)
	log.Debugf("File size: %+v", handler.Size)
	log.Debugf("MIME header: %+v", handler.Header)

	path := filepath.Join(web.directory, fmt.Sprintf("%d.%s", time.Now().Unix(), cleaned_filename))

	//Check for path traversal
	if filepath.Dir(path) != filepath.Dir(web.directory) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x02")
		return fmt.Errorf("(ERR - 0x02) Detected path traversal '%s'", path)
	}

	// Create a  file within our directory
	tFile, err := os.Create(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x03")
		return fmt.Errorf("(ERR - 0x03) Unable to create local file '%s' with error '%v'", path, err)
	}
	defer tFile.Close()

	// read all of the contents of our uploaded file into a byte array
	_, err = io.Copy(tFile, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERR - 0x04")
		return fmt.Errorf("(ERR - 0x04) Failed to read request content with error '%v'", err)
	}
	// return that we have successfully uploaded our file!
	log.Infof("Written file to %s", path)
	fmt.Fprintf(w, "OK")
	return nil
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
		log.Error("Maximal file size can not be 0 or smaller")
		return
	}

	//Create upload folder if not exists
	if _, err := os.Stat(*directory); os.IsNotExist(err) {
		err = os.Mkdir(*directory, 0766)
		if err != nil {
			log.Errorf("Unable to create folder '%s' due to following error %v", *directory, err)
			return
		}
	}
	log.Infof("Listening on :%d", *port)
	http.HandleFunc("/", web.uploadFileGeneric)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
