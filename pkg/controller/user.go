package controller

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// UserCols ...
type UserCols struct {
	ID        int    `db:"id" json:"id"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
	Mail      string `db:"mail" json:"mail"`
	Password  string `db:"password" json:"password"`
	Name      string `db:"name" json:"name"`
	Age       int    `db:"age" json:"age"`
}

type userlist []UserCols

// UserGet : get users data list
var UserGet = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Open(KindDb, Dsn)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Queryx("SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var buf UserCols
	var buflist userlist
	for rows.Next() {

		err := rows.StructScan(&buf)
		if err != nil {
			log.Fatal(err)
		}
		buflist = append(buflist, buf)
	}

	json.NewEncoder(w).Encode(buflist)
})

// UserGetByID : get users data by id
var UserGetByID = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Open(KindDb, Dsn)
	if err != nil {
		log.Fatal(err)
	}

	var buf UserCols
	err = db.QueryRowx("SELECT * FROM users WHERE id=?", mux.Vars(r)["id"]).StructScan(&buf)
	if err == nil {
		// pass
	} else if err == sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(buf)
})

// UserPost : post new data to users table
var UserPost = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	len, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := make([]byte, len)
	len, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var buf UserCols
	err = json.Unmarshal(body, &buf)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if buf.Mail == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if buf.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(buf.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open(KindDb, Dsn)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO users (mail, password, name, age) VALUES (?,?,?,?)", buf.Mail, string(hash), buf.Name, buf.Age)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == mysqlerr.ER_DUP_ENTRY {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	return
})

// UserPut : update user data by id
var UserPut = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	len, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := make([]byte, len)
	len, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var buf UserCols
	err = json.Unmarshal(body, &buf)
	if err != nil {
		log.Fatal(err)
	}

	if buf.Mail == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if buf.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := sqlx.Open(KindDb, Dsn)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("UPDATE users SET mail=?, password=?, name=?, age=? WHERE id=?", buf.Mail, buf.Password, buf.Name, buf.Age, mux.Vars(r)["id"])
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
})

// UserDelete : delete user data by id
var UserDelete = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Open(KindDb, Dsn)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM users WHERE id=?", mux.Vars(r)["id"])
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
})
