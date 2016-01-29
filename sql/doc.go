package sql

// TODO
/*
Dynamic Queries

 import conn "database/sql"
 import "github.com/reflexionhealth/vanilla/sql"

 type Customer { Name string, Purchases int }
 type Admin { Name string }

 func queryUser() {
 	var table string
 	var columns []sql.Column
 	if isAdmin {
 		table = "admins"
 		columns = sql.Columns(Admin{})
 	} else {
 		table = "customers"
 		columns = sql.Columns(Customer{})
 	}

	// Scan a single row dynamically
 	row, err := db.QueryRow(sql.SelectColumns(columns).From(table).Sql())
 	if err != nil {
 		panic("or something")
 	}
 	sql.ScanStruct(columns, row)

 	// Scan many rows dynamically
	prepared := sql.PrepareStruct()
 	for rows.Next() {
 		row := sql.ScanColumns(rows, columns)
 	}
 }
*/
