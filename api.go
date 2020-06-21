package main

import (
	"net/http"
	"strings"
)

func apiRecord(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	ok, guild, rec := parseRecordPath(req)
	if !ok {
		return respCode(http.StatusNotFound, writer, req)
	}

	if record := DatabaseRecord(guild, rec); record == nil {
		return respCode(http.StatusNotFound, writer, req)
	} else {
		return respJson(record, writer, req)
	}
}

func assetsRecord(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	ok, guild, rec := parseRecordPath(req)
	if !ok {
		return respCode(http.StatusNotFound, writer, req)
	}

	if record := DatabaseRecord(guild, rec); record == nil {
		return respCode(http.StatusNotFound, writer, req)
	} else {
		http.Redirect(writer, req, record.Path, http.StatusFound)
		return http.StatusFound
	}
}

func apiQuery(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}

	if ok, query := NewQuery(req.URL); !ok {
		return respCode(http.StatusNotFound, writer, req)
	} else {
		return respJson(DatabaseQuery(query), writer, req)
	}
}

func apiStats(writer http.ResponseWriter, req *http.Request) int {
	if req.Method != http.MethodGet {
		return respCode(http.StatusMethodNotAllowed, writer, req)
	}
	return respJson(DatabaseStats(), writer, req)
}

func htmlFile(writer http.ResponseWriter, req *http.Request) int {
	items := strings.Split(req.URL.Path, "/")
	http.ServeFile(writer, req, asset(items[1]+".html"))
	return http.StatusOK
}

func redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://thearchivist.xyz:443"+r.RequestURI, http.StatusMovedPermanently)
}

func ApiInit() {
	http.HandleFunc("/api/record/", wrap(apiRecord))
	http.HandleFunc("/api/query/", wrap(apiQuery))
	http.HandleFunc("/api/stats", wrap(apiStats))

	http.HandleFunc("/asset/record/", wrap(assetsRecord))

	http.HandleFunc("/gallery/", wrap(htmlFile))
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
