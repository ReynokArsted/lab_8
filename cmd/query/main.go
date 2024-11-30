package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "RA"
	password = "postgres"
	dbname   = "sandbox"
)

type DatabaseProvider struct {
	db *sql.DB
}

type Handlers struct {
	DProvider DatabaseProvider
}

func (h *Handlers) SetName(writer http.ResponseWriter, r *http.Request) {
	str := r.URL.Query().Get("name")

	if str == "" {
		writer.Write([]byte("Попробуй ввести своё имя через query-параметр 'name'"))
		return
	}

	err := h.DProvider.AddName(str)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(201)
	writer.Write([]byte("Запись об имени была добавлена БД"))
}

func (h *Handlers) GetName(writer http.ResponseWriter, r *http.Request) {
	answer, err := h.DProvider.SelectName()

	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(200)
	writer.Write([]byte("Привет, " + answer + "!"))
}

func (Db *DatabaseProvider) AddName(name string) error {
	_, err := Db.db.Exec("UPDATE query SET name = ($1)", name)
	if err != nil {
		return err
	}

	return nil
}

func (Db *DatabaseProvider) SelectName() (string, error) {
	var answer string

	row := Db.db.QueryRow("SELECT name FROM query LIMIT 1")
	err := row.Scan(&answer)
	if err != nil {
		return "", err
	}

	return answer, nil
}

func main() {
	address := flag.String("address", "127.0.0.1:8082", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	Db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer Db.Close()

	dp := DatabaseProvider{db: Db}
	h := Handlers{DProvider: dp}

	http.HandleFunc("/get", h.GetName)
	http.HandleFunc("/post", h.SetName)

	err = http.ListenAndServe(*address, nil)
	if err != nil {
		fmt.Print("error: server does not start")
	}
}
