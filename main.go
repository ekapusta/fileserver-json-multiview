package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// File struct
type File struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Checksum string `json:"sha1"`
}

func main() {
	port := flag.String("p", "8100", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()

	http.Handle("/", jsonDirListing(http.FileServer(http.Dir(*directory)), *directory))

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func sha1sum(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return
	}

	result = hex.EncodeToString(hash.Sum(nil))
	return
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
			err := filepath.Walk(fp, func(path string, f os.FileInfo, err error) error {
				relativePath := strings.Replace(path, fp+"/", "", 1)
				var fileType string
				var fileChecksum string
				if f.IsDir() {
					fileType = "folder"
				} else {
					fileType = "file"
					fileChecksum, err = sha1sum(path)
					if err != nil {
						return nil
					}
				}
				if relativePath != fp {
					fileList = append(fileList, File{relativePath, fileType, fileChecksum})
				}
				return nil
			})

			js, err := json.Marshal(fileList)

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
