package db_actions

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	DB_NAME = "transactions"
	DB_USER = "asrini19"
)

var transactionsDb *sql.DB

func Init_Db() {
	dbinfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable", DB_USER, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		panic(fmt.Errorf("unexpected error while initializing database %w", err))
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS userInfo(	
						user_id SERIAL PRIMARY KEY,
						email TEXT NOT NULL);`)

	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS userInstitutionInfo(
						user_inst_id SERIAL PRIMARY KEY,	
						institution_name TEXT NOT NULL,
						institution_id TEXT NOT NULL,
						access_token TEXT NOT NULL, 
						user_id INT NOT NULL,
						CONSTRAINT fk_user
							FOREIGN KEY(user_id) 
							REFERENCES userInfo(user_id)
							ON DELETE CASCADE);`)

	if err != nil {
		fmt.Println(err.Error())
	}

	transactionsDb = db
}

func QueryAllAcctsById(user_id uint64) (*sql.Rows, error) {
	println("QUERY ALL USER'S ACCOUNTS")

	query_res, err := transactionsDb.Query(`SELECT institution_name, access_token FROM userInstitutionInfo 
												WHERE user_id = $1`, user_id)

	if err != nil {
		return nil, err
	}

	return query_res, nil
}

func InsertNewAccount(user_id uint64, inst_id string, inst_name string, accessToken string) error {
	println("INSERT NEW ACCOUNT")
	query_prep, err := transactionsDb.Prepare(`INSERT INTO userInstitutionInfo 
												(institution_name,institution_id,
												access_token,user_id) VALUES ($1,$2,$3,$4)`)
	if err != nil {
		return err
	}

	_, err = query_prep.Exec(inst_name, inst_id, accessToken, user_id)

	if err != nil {
		return err
	}

	return nil
}

func QueryAccessToken(user_id uint64, inst_id string, inst_name string, accessToken *string) error {
	println("QUERY ACCESS TOKEN")
	query_res := transactionsDb.QueryRow(`SELECT access_token FROM userInstitutionInfo 
											WHERE user_id = $1
											AND institution_id = $2
											AND institution_name = $3`, user_id, inst_id, inst_name)

	return query_res.Scan(accessToken)
}

func QueryUserIdByEmail(email string, queried_user_id *uint64) error {
	println("QUERY USER ID")
	query_res := transactionsDb.QueryRow("SELECT user_id FROM userInfo WHERE email = $1", email)

	return query_res.Scan(queried_user_id)
}

func InsertNewUserReturnId(email string, queried_user_id *uint64) error {
	println("INSERT NEW USER")

	insert_res := transactionsDb.QueryRow("INSERT INTO userInfo (email) VALUES ($1) RETURNING user_id", email)
	err := insert_res.Scan(queried_user_id)

	if err != nil {
		return err
	}

	return nil
}

func QueryLinkedUserAccts(email string) (map[string]string, error) {
	var user_id uint64 = 0
	var result = make(map[string]string)

	query_err := QueryUserIdByEmail(email, &user_id)

	if query_err != nil {
		return nil, query_err
	}

	query_accts, query_accts_err := QueryAllAcctsById(user_id)

	if query_accts_err != nil {
		return nil, query_accts_err
	}

	defer query_accts.Close()

	for query_accts.Next() {
		var inst_name string
		var accessToken string

		err := query_accts.Scan(&inst_name, &accessToken)
		if err != nil {
			return nil, err
		}

		result[inst_name] = accessToken
	}
	// get any error encountered during iteration
	err := query_accts.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}
