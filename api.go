package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func apiTagGet(tag string, writer http.ResponseWriter, req *http.Request) int {
	if ids := DatabaseTag(tag); ids != nil {
		return jsonResponse(ids, writer, req)
	} else {
		return simpleResponse(http.StatusNotFound, writer, req)
	}
}

func apiTagPut(tag string, writer http.ResponseWriter, req *http.Request) int {
	val, ok := req.URL.Query()["record"]
	if !ok {
		return simpleResponse(http.StatusBadRequest, writer, req)
	}

	if len(val) != 1 {
		return simpleResponse(http.StatusBadRequest, writer, req)
	}

	id, err := strconv.ParseInt(val[0], 10, 64)
	if err != nil {
		return simpleResponse(http.StatusBadRequest, writer, req)
	}

	DatabaseTagsSet(id, []string{tag})
	return simpleResponse(http.StatusOK, writer, req)
}

func apiTag(writer http.ResponseWriter, req *http.Request) int {
	ok, tag := pathTag(req, 2)
	if !ok {
		return simpleResponse(http.StatusNotFound, writer, req)
	}

	if req.Method == http.MethodGet {
		return apiTagGet(tag, writer, req)
	} else if req.Method == http.MethodPut {
		return apiTagPut(tag, writer, req)
	}

	return simpleResponse(http.StatusMethodNotAllowed, writer, req)
}

func apiTags(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return simpleResponse(http.StatusMethodNotAllowed, writer, req)
	}
	return jsonResponse(DatabaseTags(), writer, req)
}

func apiRecord(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return simpleResponse(http.StatusMethodNotAllowed, writer, req)
	}

	ok, id := pathRecordId(req, 2)
	if !ok {
		return simpleResponse(http.StatusNotFound, writer, req)
	}

	if record := DatabaseRecord(id); record != nil {
		return jsonResponse(record, writer, req)
	} else {
		return simpleResponse(http.StatusNotFound, writer, req)
	}
}

func apiRecords(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return simpleResponse(http.StatusMethodNotAllowed, writer, req)
	}
	return jsonResponse(DatabaseRecords(100), writer, req)
}

func apiStats(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return simpleResponse(http.StatusMethodNotAllowed, writer, req)
	}
	return jsonResponse(DatabaseStats(), writer, req)
}

func assetsRecord(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return simpleResponse(http.StatusMethodNotAllowed, writer, req)
	}

	ok, id := pathRecordId(req, 2)
	if !ok {
		return simpleResponse(http.StatusNotFound, writer, req)
	}

	record := DatabaseRecord(id)
	if record == nil {
		return simpleResponse(http.StatusNotFound, writer, req)
	}

	http.Redirect(writer, req, record.Path, http.StatusFound)
	return http.StatusFound
}

func htmlTag(writer http.ResponseWriter, req *http.Request) int {
	http.ServeFile(writer, req, asset("tag.html"))
	return http.StatusOK
}

func htmlRecord(writer http.ResponseWriter, req *http.Request) int {
	http.ServeFile(writer, req, asset("record.html"))
	return http.StatusOK
}

func redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://thearchivist.xyz:443"+r.RequestURI, http.StatusMovedPermanently)
}

func simpleResponse(code int, writer http.ResponseWriter, req *http.Request) int {
	writer.WriteHeader(code)
	return code
}

func jsonResponse(resp interface{}, writer http.ResponseWriter, req *http.Request) int {
	if resp == nil {
		return simpleResponse(http.StatusNoContent, writer, req)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		Warning("unable to marshal '%v': %v", req.URL.Path, err)
		return simpleResponse(http.StatusInternalServerError, writer, req)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(bytes)
	return http.StatusOK
}

func asset(path string) string {
	return Config.Http.Assets + "/" + path
}

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

func wrap(fn func(http.ResponseWriter, *http.Request) int) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		start := time.Now()
		code := fn(writer, req)
		Info("[http] %v %v %v -> %v %v",
			req.RemoteAddr, req.Method, req.URL, code, time.Now().Sub(start))
	}
}

func ApiInit() {
	http.HandleFunc("/api/tags", wrap(apiTags))
	http.HandleFunc("/api/tag/", wrap(apiTag))
	http.HandleFunc("/api/records", wrap(apiRecords))
	http.HandleFunc("/api/record/", wrap(apiRecord))
	http.HandleFunc("/api/stats", wrap(apiStats))
	http.HandleFunc("/asset/record/", wrap(assetsRecord))
	http.HandleFunc("/tag/", wrap(htmlTag))
	http.HandleFunc("/record/", wrap(htmlRecord))
	http.Handle("/", http.FileServer(http.Dir(asset(""))))

	if Config.Http.BindTls != "" {
		go func() {
			err := http.ListenAndServeTLS(Config.Http.BindTls, Config.Http.TlsCert, Config.Http.TlsKey, nil)
			if err != nil {
				Fatal("ERROR ListenAndServeTLS error: %v", err)
			}
		}()
		go func() {
			if err := http.ListenAndServe(Config.Http.Bind, http.HandlerFunc(redirectTLS)); err != nil {
				Fatal("ERROR ListenAndServe error: %v", err)
			}
		}()
	} else {
		go func() {
			if err := http.ListenAndServe(Config.Http.Bind, nil); err != nil {
				Fatal("ERROR ListenAndServe error: %v", err)
			}
		}()
	}
}
