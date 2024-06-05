package main

import (
	"bytes"
	"embed"
	"errors"
	"html"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/runreveal/pql"
	"github.com/runreveal/pql/parser"
)

//go:embed index.html
//go:embed app.js
var static embed.FS

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveFileFS(w, r, static, "index.html")
	})
	http.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		serveFileFS(w, r, static, "app.js")
	})
	http.HandleFunc("/compile", compile)
	http.HandleFunc("/suggest", suggest)

	log.Println("Listening")
	http.ListenAndServe(":8080", nil)
}

func compile(w http.ResponseWriter, r *http.Request) {
	sql, err := pql.Compile(r.FormValue("source"))
	buf := new(bytes.Buffer)
	if err != nil {
		buf.WriteString(`<div class="italic text-red-900">`)
		buf.WriteString(html.EscapeString(err.Error()))
		buf.WriteString("</div>")
	} else {
		buf.WriteString(`<div class="font-mono bg-slate-300 text-black p-4">`)
		for _, line := range strings.Split(sql, "\n") {
			buf.WriteString(`<pre class="whitespace-pre-line ps-8 -indent-8"><code>`)
			buf.WriteString(html.EscapeString(line))
			buf.WriteString("</code></pre>\n")
		}
		buf.WriteString("</div>\n")
	}
	writeBuffer(w, buf)
}

func suggest(w http.ResponseWriter, r *http.Request) {
	ctx := &pql.AnalysisContext{
		Tables: map[string]*pql.AnalysisTable{
			"People": {
				Columns: []*pql.AnalysisColumn{
					{Name: "FirstName"},
					{Name: "LastName"},
					{Name: "PhoneNumber"},
				},
			},
		},
	}
	start, err := strconv.Atoi(r.FormValue("start"))
	if err != nil {
		http.Error(w, "start: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	end, err := strconv.Atoi(r.FormValue("end"))
	if err != nil {
		http.Error(w, "end: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}
	completions := ctx.SuggestCompletions(r.FormValue("source"), parser.Span{
		Start: start,
		End:   end,
	})

	buf := new(bytes.Buffer)
	if len(completions) == 0 {
		buf.WriteString(`<li class="p-2">No completions.</li>`)
	} else {
		for _, c := range completions {
			buf.WriteString(`<li class="p-2 hover:bg-white/25 has-[:focus]:bg-white/50 has-[:active]:bg-white/50"><a class="outline-none" href="#" data-action="analysis#fill" data-analysis-insert-param="`)
			buf.WriteString(html.EscapeString(c.Insert))
			buf.WriteString(`">`)
			buf.WriteString(html.EscapeString(c.Label))
			buf.WriteString("</a></li>\n")
		}
	}
	writeBuffer(w, buf)
}

func writeBuffer(w http.ResponseWriter, buf *bytes.Buffer) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	io.Copy(w, buf)
}

func serveFileFS(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) {
	f, err := fsys.Open(name)
	if errors.Is(err, fs.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	content, ok := f.(io.ReadSeeker)
	if !ok {
		contentBytes, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		content = bytes.NewReader(contentBytes)
	}
	http.ServeContent(w, r, name, time.Time{}, content)
}
