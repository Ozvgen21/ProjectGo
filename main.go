package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Note struct {
	ID      int
	Title   string
	Content string
}

var (
	notes  = []Note{}
	nextID = 1
	mu     sync.Mutex
)

func main() {
	http.HandleFunc("/", listNotes)
	http.HandleFunc("/create", createNote)
	http.HandleFunc("/edit", editNote)
	http.HandleFunc("/delete", deleteNote)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func listNotes(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"truncate": func(s string, n int) string {
			if len(s) > n {
				return s[:n] + "..."
			}
			return s
		},
	}
	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, notes)
}

func createNote(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")

		mu.Lock()
		notes = append(notes, Note{ID: nextID, Title: title, Content: content})
		nextID++
		mu.Unlock()

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "create.html")
}

func editNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")

		mu.Lock()
		for i := range notes {
			if notes[i].ID == id {
				notes[i].Title = title
				notes[i].Content = content
				break
			}
		}
		mu.Unlock()

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	mu.Lock()
	var note Note
	for _, n := range notes {
		if n.ID == id {
			note = n
			break
		}
	}
	mu.Unlock()

	tmpl, err := template.ParseFiles("edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, note)
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	for i := range notes {
		if notes[i].ID == id {
			notes = append(notes[:i], notes[i+1:]...)
			break
		}
	}
	mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
