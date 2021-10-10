package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/dhamith93/shortcut/internal/command"
	"github.com/dhamith93/shortcut/internal/fileops"
	"github.com/dhamith93/shortcut/internal/logger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Meta struct {
	Url string
}

var fileList = fileops.GetFileList()
var server = http.Server{}

// Run starts the server in given port
func Run() {
	fileops.CleanUp()
	port := fileops.ReadFile("port.txt", ":5500")
	handleRequests(port)
}

// Shutdown stops the server after cleaning up the files
func Shutdown(ctx context.Context) {
	fileops.CleanUp()
	server.Shutdown(ctx)
}

func handleRequests(port string) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/meta", sendMeta)
	router.Path("/upload").Methods("POST").HandlerFunc(handleFile)
	router.HandleFunc("/files", getFiles)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	server.Addr = port
	server.Handler = handlers.CompressHandler(router)
	server.SetKeepAlivesEnabled(false)

	logger.Log("info", "Shortcut started on http://"+getOutboundIP()+port)
	go command.Open("http://" + getOutboundIP() + port)
	log.Fatal(server.ListenAndServe())
}

func sendMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	meta := Meta{
		Url: "http://" + getOutboundIP() + server.Addr,
	}
	json.NewEncoder(w).Encode(&meta)
}

func getFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&fileList)
}

func handleFile(w http.ResponseWriter, r *http.Request) {
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

	fileList, err = fileops.HandleFile(file, device, fileName)
	if err != nil {
		logger.Log("error", err.Error())
	}
	getFiles(w, r)
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
