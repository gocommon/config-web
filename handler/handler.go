package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/yosssi/ace"
	"golang.org/x/net/context"

	//	proto "github.com/micro/go-platform/config/proto"
	config "github.com/micro/config-srv/proto/config"
)

var (
	opts         *ace.Options
	configClient config.ConfigClient
)

func Init(dir string, t config.ConfigClient) {
	configClient = t

	opts = ace.InitializeOptions(nil)
	opts.BaseDir = dir
	opts.DynamicReload = true
	opts.FuncMap = template.FuncMap{
		"JSON": func(d string) string {
			return prettyJSON(d)
		},
		"TimeAgo": func(t int64) string {
			return timeAgo(t)
		},
		"Timestamp": func(t int64) string {
			return time.Unix(t, 0).Format(time.RFC822)
		},
		"Colour": func(s string) string {
			return colour(s)
		},
	}
}

func render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	basePath := hostPath(r)

	opts.FuncMap["URL"] = func(path string) string {
		return filepath.Join(basePath, path)
	}

	tpl, err := ace.Load("layout", tmpl, opts)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 302)
		return
	}

	if err := tpl.Execute(w, data); err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 302)
	}
}

// The index page
func Index(w http.ResponseWriter, r *http.Request) {
	rsp, err := configClient.AuditLog(context.TODO(), &config.AuditLogRequest{
		Reverse: true,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	sort.Sort(sortedLogs{logs: rsp.Changes, reverse: false})

	render(w, r, "index", map[string]interface{}{
		"Latest": rsp.Changes,
	})
}

// The Audit Log
func AuditLog(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limit := 25
	from, _ := strconv.Atoi(r.Form.Get("from"))
	to, _ := strconv.Atoi(r.Form.Get("to"))

	page, err := strconv.Atoi(r.Form.Get("p"))
	if err != nil {
		page = 1
	}

	if page < 1 {
		page = 1
	}

	offset := (page * limit) - limit

	rsp, err := configClient.AuditLog(context.TODO(), &config.AuditLogRequest{
		From:    int64(from),
		To:      int64(to),
		Limit:   int64(limit),
		Offset:  int64(offset),
		Reverse: true,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	sort.Sort(sortedLogs{logs: rsp.Changes, reverse: false})

	var less, more int

	if len(rsp.Changes) == limit {
		more = page + 1
	}

	if page > 1 {
		less = page - 1
	}

	render(w, r, "audit", map[string]interface{}{
		"Latest": rsp.Changes,
		"Less":   less,
		"More":   more,
	})
}

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		limit := 25
		id := r.Form.Get("id")
		author := r.Form.Get("author")

		page, err := strconv.Atoi(r.Form.Get("p"))
		if err != nil {
			page = 1
		}

		if page < 1 {
			page = 1
		}

		offset := (page * limit) - limit

		rsp, err := configClient.Search(context.TODO(), &config.SearchRequest{
			Id:     id,
			Author: author,
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			http.Redirect(w, r, filepath.Join(hostPath(r), "search"), 302)
			return
		}

		q := ""

		if len(id) > 0 {
			q += "id: " + id + ", "
		}

		if len(author) > 0 {
			q += "author: " + author
		}

		var less, more int

		if len(rsp.Configs) == limit {
			more = page + 1
		}

		if page > 1 {
			less = page - 1
		}

		sort.Sort(sortedConfigs{configs: rsp.Configs})

		render(w, r, "results", map[string]interface{}{
			"Name":    q,
			"Results": rsp.Configs,
			"Less":    less,
			"More":    more,
		})

		return
	}
	render(w, r, "search", map[string]interface{}{})
}

func Config(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) == 0 {
		http.Redirect(w, r, "/", 302)
		return
	}

	rsp, err := configClient.Read(context.TODO(), &config.ReadRequest{
		Id: id,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	render(w, r, "config", map[string]interface{}{
		"Id":     id,
		"Config": rsp.Change,
	})
}
