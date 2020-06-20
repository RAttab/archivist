package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func pathRecordId(req *http.Request, pos int) (bool, int64) {
	items := strings.Split(req.URL.Path, "/")[1:]
	if len(items) != pos+1 {
		return false, 0
	}

	id, err := strconv.ParseInt(items[pos], 10, 64)
	if err != nil {
		return false, 0
	}

	return true, id
}

func pathTag(req *http.Request, pos int) (bool, string) {
	items := strings.Split(req.URL.Path, "/")[1:]
	if len(items) != pos+1 {
		return false, ""
	}
	return true, items[pos]
}

func apiTagGet(tag string, writer http.ResponseWriter, req *http.Request) {
	if ids := DatabaseTag(tag); ids != nil {
		jsonResponse(ids, writer, req)
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func apiTagPut(tag string, writer http.ResponseWriter, req *http.Request) {
	val, ok := req.URL.Query()["record"]
	if !ok {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(val) != 1 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(val[0], 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	DatabaseTagsSet(id, []string{tag})
	writer.WriteHeader(http.StatusOK)
}

func apiTag(writer http.ResponseWriter, req *http.Request) {
	ok, tag := pathTag(req, 2)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if req.Method == http.MethodGet {
		apiTagGet(tag, writer, req)
	} else if req.Method == http.MethodPut {
		apiTagPut(tag, writer, req)
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func apiTags(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	jsonResponse(DatabaseTags(), writer, req)
}

func apiRecord(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ok, id := pathRecordId(req, 2)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if record := DatabaseRecord(id); record != nil {
		jsonResponse(record, writer, req)
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func apiRecords(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	jsonResponse(DatabaseRecords(100), writer, req)
}

func apiStats(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	jsonResponse(DatabaseStats(), writer, req)
}

func jsonResponse(resp interface{}, writer http.ResponseWriter, req *http.Request) {
	if resp == nil {
		writer.WriteHeader(http.StatusNoContent)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("ERROR: unable to marshal '%v': %v", req.URL.Path, err)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(bytes)
}

func assetsRecord(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ok, id := pathRecordId(req, 2)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	record := DatabaseRecord(id)
	if record == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Add("Location", record.Path)
	writer.WriteHeader(http.StatusFound)
}

func htmlTag(writer http.ResponseWriter, req *http.Request) {
	http.ServeFile(writer, req, Config.AssetPath+"/tag.html")
}

func htmlRecord(writer http.ResponseWriter, req *http.Request) {
	http.ServeFile(writer, req, Config.AssetPath+"/record.html")
}

func redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://thearchivist.xyz:443"+r.RequestURI, http.StatusMovedPermanently)
}

func ApiInit() {
	http.HandleFunc("/api/tags", apiTags)
	http.HandleFunc("/api/tag/", apiTag)
	http.HandleFunc("/api/records", apiRecords)
	http.HandleFunc("/api/record/", apiRecord)
	http.HandleFunc("/api/stats", apiStats)
	http.HandleFunc("/asset/record/", assetsRecord)
	http.HandleFunc("/tag/", htmlTag)
	http.HandleFunc("/record/", htmlRecord)
	http.Handle("/", http.FileServer(http.Dir(Config.AssetPath)))

	if Config.CertFile != "" || Config.KeyFile != "" {
		go func() {
			if err := http.ListenAndServeTLS(Config.Bind, Config.CertFile, Config.KeyFile, nil); err != nil {
				log.Fatalf("ERROR ListenAndServeTLS error: %v", err)
			}
		}()
		go func() {
			if err := http.ListenAndServe(":80", http.HandlerFunc(redirectTLS)); err != nil {
				log.Fatalf("ERROR ListenAndServe error: %v", err)
			}
		}()
	} else {
		go func() {
			if err := http.ListenAndServe(Config.Bind, nil); err != nil {
				log.Fatalf("ERROR ListenAndServe error: %v", err)
			}
		}()
	}
}
