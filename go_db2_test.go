package main

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/alexbrainman/odbc"
	_ "github.com/alexbrainman/odbc"
)

func NewConn() (*sql.DB, error) {
	//db, err := sql.Open("odbc", `Driver={DB2};DSN=DB2_SAMPLE`)
	db, err := sql.Open("odbc", `DSN=DB2_SAMPLE;uid=db2inst1;pwd=db2inst1`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestNewConnect(t *testing.T) {
	db, err := NewConn()
	if err != nil {
		t.Fatal("connect error", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Fatalf("ping error: %v", err)
	}

	db2, err := sql.Open("odbc", `DSN=AAA;uid=db2inst1;pwd=db2inst1`)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	defer db2.Close()

	err = db.Ping()
	if err != nil {
		t.Fatalf("ping error: %v", err)
	}

}

func TestNewConnect2(t *testing.T) {
	db2, err := sql.Open("odbc", `DSN=DB2_SAMPLE;uid=db2inst1;pwd=db2inst1`)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	defer db2.Close()

	err = db2.Ping()
	if err != nil {
		t.Fatalf("ping error: %v", err)
	}

}

func TestQuery(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	var n int
	rows, err := db.Query("SELECT 1 FROM sysibm.sysdummy1");
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&n)
		if err != nil || n != 1 {
			t.Fatalf("got %d, expected %d, err:%v", n, 1, err)
		}
	}

	if err = rows.Err(); err != nil {
		t.Fatalf("got err: %v", err)
	}
}

func TestPrepareQuery(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	var tabName, tabSchema string
	stmt, err := db.Prepare("select tabname, tabschema from syscat.tables where tabschema = ?")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query("SYSCAT")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		err := rows.Scan(&tabName, &tabSchema)
		if err != nil {
			t.Fatalf("err:%v", err)
		}
		n++
	}
	log.Printf("got %d records\n", n)

	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}

}

func TestQuerySingleRow(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	var tabName string
	// example 1
	err := db.QueryRow("select tabname from syscat.tables where tabschema = ?", "SYSCAT").Scan(&tabName)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	log.Println(tabName)

	// example 2
	stmt, err := db.Prepare("select tabname from syscat.tables where tabschema = ?")
	if err != nil {
		t.Fatal(err)
	}

	err = stmt.QueryRow("SYSCAT").Scan(&tabName)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(tabName)
}

func TestInsert(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO TEST(ID, NAME) VALUES(?, ?)")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(1, "AAA")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransaction(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO TEST(ID, NAME) VALUES(?, ?)")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close() // danger!

	for i := 0; i < 10; i++ {
		_, err = stmt.Exec(i, "AAA")
		if err != nil {
			t.Fatal(err)
		}
	}
	// time.Sleep(time.Second * 10)

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}
	// stmt.Close() runs here!
}

func TestHandleError(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	rows, err := db.Query("select * from AAA")
	if err != nil {
		log.Println("query err:", err)
		if driverErr, ok := err.(*odbc.Error); ok {
			log.Println(">>>", "db2 sqlcode:", driverErr.Diag[0].NativeError)
		}
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	defer rows.Close()

	if err = rows.Err(); err != nil {
		// handle the error here
		log.Println(err)
	}

}

func TestNullString(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	rows, err := db.Query("select COALESCE(name, '') from test")
	if err != nil {
		t.Fatalf("error:%v", err)
	}
	defer rows.Close()

	var name sql.NullString
	for rows.Next() {
		_ = rows.Scan(&name)
		if name.Valid {
			// use s.String
		} else {
			// error handle
		}
	}
}

func TestUnknownColumn(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	rows, err := db.Query("select * from test where name is not null")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	//log.Println(rows.Columns())
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("err:%v", err)
	}

	vals := make([]interface{}, len(cols))
	for i, _ := range vals {
		vals[i] = new(sql.RawBytes)
	}

	for rows.Next() {
		err := rows.Scan(vals...)
		if err != nil {
			t.Fatalf("%v", err)
		}
		//log.Println(vals...)
		//log.Println(vals[0], vals[1])
	}
}

func TestConnectionPool(t *testing.T) {
	db, _ := NewConn()
	defer db.Close()

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	for i := 0; i < 100; i++ {
		db.Query("select * from test where name is not null")
	}

	time.Sleep(time.Hour)
	log.Println(db.Stats())
}
