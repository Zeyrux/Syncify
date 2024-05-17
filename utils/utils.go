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
