package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var (
	//go:embed templates
	templateFS embed.FS
	//go:embed static
	staticFS embed.FS
)

type Task struct {
	id		int
	title	string
}

func ListTasks(w http.ResponseWriter, r *http.Request) {

}

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	tpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Handle("/static/*", http.FileServer(http.FS(staticFS)))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tpl.ExecuteTemplate(w, "index.html", nil)
	})
	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			rows, err := db.Query("SELECT * FROM tasks")
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var tasks []Task

			for rows.Next() {
				var task Task
				err := rows.Scan(&task.id, &task.title)
				if err != nil {
					log.Fatal(err)
				}
				tasks = append(tasks, task)
			}

			if tasks != nil {
				tpl.ExecuteTemplate(w, "task-list.html", tasks)
				return
			}
		})
		r.Post("/add", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				return
			}

			title := r.Form.Get("task")
			task := Task{title: title}

			_, err = db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
			if err != nil {
				log.Fatal(err)
			}

			tpl.ExecuteTemplate(w, "list-item.html", task)
		})
	})

	fmt.Printf("server started on :3000")
	http.ListenAndServe(":3000", r)
}
