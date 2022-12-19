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
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	sessionCookie := getSessionCookie(w, req)
	imgs := getImages(w, req)
	tplValues := struct {
		SessionValue string
		Images       []string
	}{
		SessionValue: sessionCookie.Value,
		Images:       imgs,
	}
	tpl.ExecuteTemplate(w, "index.gohtml", tplValues)
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

		saveFileCookie(w, req, fname)
	}

	http.Redirect(w, req, req.Header.Get("Referer"), http.StatusFound)
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
	http.SetCookie(w, c)

	return c
}

func getImages(w http.ResponseWriter, req *http.Request) []string {
	c := getFileCookie(w, req)
	var imgs []string
	if len(c.Value) > 0 {
		imgs = strings.Split(c.Value, "|")
	}

	return imgs
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

func saveFileCookie(w http.ResponseWriter, req *http.Request, fileName string) *http.Cookie {
	c := getFileCookie(w, req)

	appendCookieValue(w, c, fileName)

	return c
}

func getFileCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	c, err := req.Cookie("files")
	if err != nil {
		c = createFileCookie(w, req)
	}

	return c
}

func createFileCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	c := &http.Cookie{
		Name: "files",
	}

	return c
}

func appendCookieValue(w http.ResponseWriter, c *http.Cookie, fileName string) *http.Cookie {
	s := c.Value
	if !strings.Contains(s, fileName) {
		if len(s) != 0 {
			s += "|"
		}
		s += fileName
	}

	c.Value = s
	http.SetCookie(w, c)
	return c
}
