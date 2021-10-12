package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type meta struct {
	Url string
}

type handler struct {
	server   http.Server
	fileList []FileList
}

func (h *handler) handleRequests(port string) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/meta", h.sendMeta)
	router.Path("/upload").Methods("POST").HandlerFunc(h.handleFile)
	router.HandleFunc("/files", h.getFiles)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	h.server.Addr = port
	h.server.Handler = handlers.CompressHandler(router)
	h.server.SetKeepAlivesEnabled(false)

	Log("info", "Shortcut started on http://"+getOutboundIP()+port)
	go openBrowser("http://" + getOutboundIP() + port)
	log.Fatal(h.server.ListenAndServe())
}

func (h *handler) shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *handler) sendMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(&meta{
		Url: "http://" + getOutboundIP() + h.server.Addr,
	})
}

func (h *handler) getFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&h.fileList)
}

func (h *handler) handleFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	var buf bytes.Buffer
	defer buf.Reset()
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()
	device := r.Form.Get("device-name")
	fileName := header.Filename

	h.fileList, err = HandleFile(file, device, fileName)
	if err != nil {
		Log("error", err.Error())
	}
	h.getFiles(w, r)
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}