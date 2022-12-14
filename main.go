package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	uuid "github.com/satori/go.uuid"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/fileupload", fileuploadHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	sessionCookie := getSessionCookie(w, req)
	tpl.ExecuteTemplate(w, "index.gohtml", sessionCookie)
}

func fileuploadHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		mf, fh, err := req.FormFile("nf")
		if err != nil {
			fmt.Println(err)
		}
		defer mf.Close()

		fname := generateFileSHAname(mf, fh)

		saveFile(mf, fname)
	}

	http.Redirect(w, req, req.Header.Get("Referer"), 302)
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "assets/favicon.ico")
}

func getSessionCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	c, err := req.Cookie("session")
	if err != nil {
		c = createSessionCookie(w, req)
	}

	return c
}

func createSessionCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	sId := uuid.NewV4()
	c := &http.Cookie{
		Name:  "session",
		Value: sId.String(),
	}

	return c
}

func generateFileSHAname(mf multipart.File, fh *multipart.FileHeader) string {
	ext := strings.Split(fh.Filename, ".")[1]
	h := sha1.New()
	io.Copy(h, mf)

	return fmt.Sprintf("%x", h.Sum(nil)) + "." + ext
}

func saveFile(mf multipart.File, fileName string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	path := filepath.Join(wd, "public", "upload", fileName)
	nf, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	defer nf.Close()

	mf.Seek(0, 0)
	io.Copy(nf, mf)
}
