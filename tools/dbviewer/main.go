package main

import (
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"strings"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var db *sql.DB

func main() {
	var err error
	// read-only, WAL-safe: never mutates the live DB
	db, err = sql.Open("sqlite3", "file:/app/exg.db?mode=ro&_pragma=busy_timeout(5000)&_pragma=query_only(true)")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", handle)
	fmt.Println("DB viewer on :8090")
	http.ListenAndServe(":8090", nil)
}

const style = `<style>
body{font-family:Menlo,Consolas,monospace;background:#12121a;color:#e8e8e8;padding:16px}
a{color:#4da3ff;text-decoration:none;margin-right:10px}a:hover{text-decoration:underline}
table{border-collapse:collapse;width:100%;margin-top:10px;font-size:12px}
th,td{border:1px solid #333;padding:4px 8px;text-align:left;white-space:nowrap;max-width:340px;overflow:hidden;text-overflow:ellipsis}
th{background:#22223a;position:sticky;top:0}
tr:nth-child(even){background:#1a1a26}
h2{color:#00ffcc}.tables a{display:inline-block;background:#22223a;padding:4px 10px;border-radius:6px;margin:3px}
input[type=text]{width:70%;background:#1a1a26;color:#e8e8e8;border:1px solid #444;padding:6px;font-family:inherit}
button{background:#0b5fff;color:#fff;border:0;padding:6px 12px;border-radius:4px;cursor:pointer}
.wrap{overflow-x:auto}
</style>`

func handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fmt.Fprint(w, "<!doctype html><html><head><meta charset=utf-8><title>exg.db viewer</title>", style, "</head><body>")
	fmt.Fprint(w, `<h2>&#128190; exg.db &mdash; read-only viewer</h2>`)

	// table list
	fmt.Fprint(w, `<div class="tables"><b>Tables:</b> `)
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`)
	if err == nil {
		for rows.Next() {
			var name string
			rows.Scan(&name)
			var cnt int64
			db.QueryRow(`SELECT COUNT(*) FROM "` + name + `"`).Scan(&cnt)
			fmt.Fprintf(w, `<a href="/?t=%s">%s (%d)</a>`, name, html.EscapeString(name), cnt)
		}
		rows.Close()
	}
	fmt.Fprint(w, `</div>`)

	// query box
	sqlStr := q.Get("sql")
	table := q.Get("t")
	if sqlStr == "" && table != "" {
		sqlStr = fmt.Sprintf(`SELECT * FROM "%s" ORDER BY rowid DESC LIMIT 200`, table)
	}
	fmt.Fprintf(w, `<form method=get style="margin-top:12px"><input type=text name=sql placeholder="SELECT ... (read-only)" value="%s"><button>Run</button></form>`, html.EscapeString(sqlStr))

	if sqlStr != "" {
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlStr)), "SELECT") {
			fmt.Fprint(w, `<p style="color:#ff6666">Only SELECT queries are allowed.</p>`)
		} else {
			renderQuery(w, sqlStr)
		}
	}
	fmt.Fprint(w, "</body></html>")
}

func renderQuery(w http.ResponseWriter, sqlStr string) {
	rows, err := db.Query(sqlStr)
	if err != nil {
		fmt.Fprintf(w, `<p style="color:#ff6666">%s</p>`, html.EscapeString(err.Error()))
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	fmt.Fprint(w, `<div class="wrap"><table><tr>`)
	for _, c := range cols {
		fmt.Fprintf(w, "<th>%s</th>", html.EscapeString(c))
	}
	fmt.Fprint(w, "</tr>")
	n := 0
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		fmt.Fprint(w, "<tr>")
		for _, v := range vals {
			fmt.Fprintf(w, "<td>%s</td>", html.EscapeString(fmt.Sprintf("%v", v)))
		}
		fmt.Fprint(w, "</tr>")
		n++
	}
	fmt.Fprintf(w, "</table></div><p>%d rows</p>", n)
}
