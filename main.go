package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// File struct
type File struct {
	Name string `json:"name"`
}

// Folder struct
type Folder struct {
	File
}

// FileList struct
type FileList struct {
	Folders []Folder `json:"folders"`
	Files   []File   `json:"files"`
}

func main() {
	port := flag.String("p", "8100", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()

	http.Handle("/", jsonDirListing(http.FileServer(http.Dir(*directory)), *directory))

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func jsonDirListing(h http.Handler, directory string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fp := filepath.Join(directory, filepath.Clean(r.URL.Path))
		info, err := os.Stat(fp)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
		}

		if info.IsDir() {
			fileList := []File{}
			dirList := []Folder{}
			err := filepath.Walk(fp, func(path string, f os.FileInfo, err error) error {
				relativePath := strings.Replace(path, fp+"/", "", 1)
				if !f.IsDir() {
					fileList = append(fileList, File{relativePath})
				} else {
					if relativePath != fp {
						dirList = append(dirList, Folder{File{relativePath}})
					}
				}
				return nil
			})

			js, err := json.Marshal(FileList{dirList, fileList})

			if err != nil {
				http.Error(w, "500", http.StatusInternalServerError)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)

			log.Printf("LST %s", r.URL)
			return
		}

		log.Printf("GET %s", r.URL)
		h.ServeHTTP(w, r)
	})
}
