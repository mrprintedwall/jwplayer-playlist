//Author: John Pili
//Website: https://www.johnpili.com

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-zoo/bone"
)

type movie struct {
	File        string `json:"file"`
	Image       string `json:"image"`
	Title       string `json:"title"`
	Description string `json:"description"`
	MediaId     string `json:"mediaid"`
}

var targetPath string // Target to scan .mp4 files. Example: /mnt/movies/
var webPath string    // Your web path. This is the path you set in nginx or apache. Example: /movies -> part of your http://www.abc.com/movies/

func main() {
	args := os.Args[1:] // Fetch the program arguments excluding the program name

	router := bone.New()                              // Bone is awesome!
	router.Get("/", http.HandlerFunc(defaultHandler)) // Default page. You can specify ?k for the search keyword

	targetPath = args[1] // Set OS path from program argument
	webPath = args[2]    // Set Web path from program argument

	log.Fatal(http.ListenAndServe(":"+args[0], router)) // Start the HTTP server with your specified port
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	respondWithJson(w, http.StatusOK, fileWalker(v.Get("k")))
}

// This method will walk into your specified path recursively to find all .mp4 files
func fileWalker(keyword string) []movie {
	files := []movie{}
	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".mp4" { // My requirement is just for .mp4 files, perhaps you can change to suit your needs
			x := strings.Replace(path, targetPath, webPath, 1) // Replace OS path with web path
			x = url.PathEscape(x)                              // URL encode the path
			x = strings.Replace(x, "%2F", "/", -1)             // Put back "/" in the string after URL encode. There is a better way of doing this. I will check later
			if len(keyword) > 0 {                              // If keyword is not empty, do a search using contains
				var s1 = strings.ToUpper(keyword)
				var s2 = strings.ToUpper(filepath.Base(path)) // For the sake of simplicity I just convert both the path string and the keywords to lowercase
				if strings.Contains(s2, s1) {                 // Search using strings.Contains. This is enough for my use case
					var movie movie
					movie.File = x
					movie.Title = filepath.Base(path)
					movie.Image = ""
					files = append(files, movie)
				}
			} else { // Otherwise just generate the list
				var movie movie
				movie.File = x
				movie.Title = filepath.Base(path)
				movie.Image = ""
				files = append(files, movie)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("walk error [%v]\n", err)
	}
	return files
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
