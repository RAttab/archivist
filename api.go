package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func apiTag(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	ok, guild, tag := path(req)
	if !ok || tag == "" {
		return respCode(http.StatusNotFound, writer, req)
	}

	if ids := DatabaseTag(guild, tag); ids != nil {
		return respJson(ids, writer, req)
	} else {
		return respCode(http.StatusNotFound, writer, req)
	}
}

func apiTags(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	ok, guild, tag := path(req)
	if !ok || tag != "" {
		return respCode(http.StatusNotFound, writer, req)
	}

	return respJson(DatabaseTags(guild), writer, req)
}

func apiRecord(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	ok, guild, rec := pathIntId(req)
	if !ok || rec < 0 {
		return respCode(http.StatusNotFound, writer, req)
	}

	if record := DatabaseRecord(guild, rec); record != nil {
		return respJson(record, writer, req)
	} else {
		return respCode(http.StatusNotFound, writer, req)
	}
}

func apiRecords(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	ok, guild, rec := path(req)
	if !ok || rec != "" {
		return respCode(http.StatusNotFound, writer, req)
	}

	return respJson(DatabaseRecords(guild, 20), writer, req)
}

func apiStats(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}
	return respJson(DatabaseStats(), writer, req)
}

func assetsRecord(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	ok, guild, rec := pathIntId(req)
	if !ok || rec < 0 {
		return respCode(http.StatusNotFound, writer, req)
	}

	record := DatabaseRecord(guild, rec)
	if record == nil {
		return respCode(http.StatusNotFound, writer, req)
	}

	http.Redirect(writer, req, record.Path, http.StatusFound)
	return http.StatusFound
}

func htmlFile(writer http.ResponseWriter, req *http.Request) int {
	items := strings.Split(req.URL.Path, "/")
	http.ServeFile(writer, req, asset(items[1]+".html"))
	return http.StatusOK
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

func respCode(code int, writer http.ResponseWriter, req *http.Request) int {
	writer.WriteHeader(code)
	return code
}

func respJson(resp interface{}, writer http.ResponseWriter, req *http.Request) int {
	if resp == nil {
		return respCode(http.StatusNoContent, writer, req)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		Warning("unable to marshal '%v': %v", req.URL.Path, err)
		return respCode(http.StatusInternalServerError, writer, req)
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

func path(req *http.Request) (bool, string, string) {
	items := strings.Split(req.URL.Path, "/")[1:]

	if n := len(items); n != 3 && n != 4 {
		return false, "", ""
	} else if n == 3 {
		return true, items[2], ""
	} else {
		return true, items[2], items[3]
	}
}

func pathIntId(req *http.Request) (bool, string, int64) {
	ok, guild, id := path(req)
	if !ok || id == "" {
		return false, "", -1
	}

	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return false, "", -1
	}

	return true, guild, intId
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
	http.HandleFunc("/api/tags/", wrap(apiTags))
	http.HandleFunc("/api/tag/", wrap(apiTag))
	http.HandleFunc("/api/records/", wrap(apiRecords))
	http.HandleFunc("/api/record/", wrap(apiRecord))
	http.HandleFunc("/api/stats", wrap(apiStats))
	http.HandleFunc("/asset/record/", wrap(assetsRecord))
	http.HandleFunc("/tag/", wrap(htmlFile))
	http.HandleFunc("/record/", wrap(htmlFile))
	http.HandleFunc("/records/", wrap(htmlFile))
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
