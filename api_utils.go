package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseRecordPath(req *http.Request) (bool, string, int64) {
	items := strings.Split(req.URL.Path, "/")
	if len(items) != 5 || items[2] != "record" {
		return false, "", -1
	}

	guild := items[3]
	if _, err := strconv.ParseInt(guild, 10, 64); err != nil {
		return false, "", -1
	}

	rec, err := strconv.ParseInt(items[4], 10, 64)
	if err != nil {
		return false, "", -1
	}

	return true, guild, rec
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

func wrap(fn func(http.ResponseWriter, *http.Request) int) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		start := time.Now()
		code := fn(writer, req)
		Info("[http] %v %v %v -> %v %v",
			req.RemoteAddr, req.Method, req.URL, code, time.Now().Sub(start))
	}
}
