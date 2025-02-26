package main

import (
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func main() {
	flag.Parse()

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	contentType := mw.FormDataContentType()

	hello_world := "hello_world"
	r := strings.NewReader(hello_world)

	go func() {
		fw, err := mw.CreateFormFile("file", "again")
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(fw, r); err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := mw.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()
	req, err := http.NewRequest("POST", "http://localhost:8080/create", pr)
	if err != nil {
		fmt.Println("Could not create request:", err.Error())
		return
	}
	req.Header.Add("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making POST request:", err.Error())
		return
	}

	defer resp.Body.Close()
	fmt.Println("Response Status:", resp.Status)
}
