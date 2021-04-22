package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"chromedpsample/headless"
	"chromedpsample/utils"
)

var (
	cliHeadless *headless.Client
)

func handler(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get working directory: %v", err), http.StatusInternalServerError)
		return
	}

	id := utils.GenCode(6)
	dlURL := "https://dl-cdn.alpinelinux.org/alpine/v3.13/releases/x86_64/alpine-standard-3.13.5-x86_64.iso"
	// dlURL := "http://www.mersenne.org/ftp_root/gimps/p95v287.MacOSX.noGUI.tar.gz"
	dlDir := path.Join(wd, "download", id)
	if err = cliHeadless.Download(id, dlURL, dlDir); err != nil {
		http.Error(w, fmt.Sprintf("ID:%s failed to download file", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func main() {
	cliHeadless = headless.NewClient()

	http.HandleFunc("/download", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
