package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func PrintRequest(request *http.Request) {
	requestDump, err := httputil.DumpRequest(request, true)
	HandleError(err)
	fmt.Println("REQUEST:\n", string(requestDump))
}

func PrintResponse(response *http.Response) {
	responseDump, err := httputil.DumpResponse(response, true)
	HandleError(err)
	fmt.Println("RESPONSE:\n", string(responseDump))
}

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func DefaultLog(message string) {
	log.Default().Println(message)
}

func MarshalConfigYaml(rootDir string, parallelDownloads int, playlists []string) []byte {
	marshaledString := "root_dir: " + rootDir + "\n"
	marshaledString += "parallel_downloads_per_playlist: " + string(parallelDownloads) + "\n"
	marshaledString += "playlists:\n"
	for _, id := range playlists {
		marshaledString += "  - id: " + id + " # " + id + "\n"
	}
	return []byte(marshaledString)
}
