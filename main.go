package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

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
	ID    int
	Title string
}

type TaskRepo struct {
	db *sql.DB
	//tasks []Task
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "tasks.db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			title TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func initMux(db *sql.DB) *chi.Mux {
	repo := TaskRepo{db: db}
	tpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Handle("/static/*", http.FileServer(http.FS(staticFS)))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tasks, err := repo.GetTasks()
		if err != nil {
			log.Fatal(err)
		}
		tpl.ExecuteTemplate(w, "index.html",tasks)
	})
	r.Post("/add", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			return
		}
		repo.AddTask(r.FormValue("title"))
		tasks, err := repo.GetTasks()
		if err != nil {
			log.Fatal(err)
		}
		tpl.ExecuteTemplate(w, "task-list.html", tasks)
	})
	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Fatal(err)
			return
		}
		repo.DeleteTask(id)
		tasks, err := repo.GetTasks()
		if err != nil {
			log.Fatal(err)
		}
		tpl.ExecuteTemplate(w, "task-list.html", tasks)
	})
	return r
}

func (tr *TaskRepo) AddTask(title string) error {
	_, err := tr.db.Exec("INSERT INTO tasks(title) VALUES (?)", title)
	if err != nil {
		return err
	}
	return nil
}

func (tr *TaskRepo) GetTasks() ([]Task, error) {
	rows, err := tr.db.Query("SELECT rowid,title FROM tasks")
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.ID, &task.Title)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (tr *TaskRepo) DeleteTask(id int) error {
	_, err := tr.db.Exec("DELETE FROM tasks WHERE rowid = ?", id)
	return err
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := initMux(db)
	fmt.Println("server started on :3000")
	http.ListenAndServe(":3000", r)
}
