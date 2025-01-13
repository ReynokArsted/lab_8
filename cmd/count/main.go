package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "RA"
	password = "postgres"
	dbname   = "sandbox"
)

type DatabaseProvider struct { // Структура с полем, которое хранит ссылку на СУБД
	db *sql.DB
}

type Handlers struct {
	DProvider DatabaseProvider
}

func (h *Handlers) GetCount(writer http.ResponseWriter, r *http.Request) {
	answer, err := h.DProvider.SelectCount()
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(200)
	writer.Write([]byte("Значение счётчика: " + answer))
}

func (h *Handlers) SetCount(writer http.ResponseWriter, r *http.Request) {
	input := struct {
		Massage string `json:massage`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte(err.Error()))
		return
	}

	value, err := strconv.Atoi(input.Massage)
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte("Было введено не число или присутствуют пробелы в записи числа"))
		return
	}

	err = h.DProvider.UpdateCount(value)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(201)
	if value > 0 {
		writer.Write([]byte("Значение счётчика было изменено на +" + strconv.Itoa(value)))
	} else {
		writer.Write([]byte("Значение счётчика было изменено на " + strconv.Itoa(value)))
	}
}

func (Dp *DatabaseProvider) SelectCount() (string, error) {
	var dbAnswer string

	row := Dp.db.QueryRow("SELECT value FROM count")
	err := row.Scan(&dbAnswer) // Проверка на то, есть ли искомые данные в БД
	if err != nil {
		return "", err
	}

	return dbAnswer, nil
}

func (Dp *DatabaseProvider) UpdateCount(n int) error {
	_, err := Dp.db.Exec("UPDATE count SET value = value + ($1)", n)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	address := flag.String("address", "127.0.0.1:8083", "адрес для запуска сервера")
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

	http.HandleFunc("/get", h.GetCount)
	http.HandleFunc("/post", h.SetCount)

	err = http.ListenAndServe(*address, nil)
	if err != nil {
		fmt.Print("error: server does not start")
	}
}
