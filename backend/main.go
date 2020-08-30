package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/plaid/plaid-go/plaid"
)

var client = func() *plaid.Client {
	client, err := plaid.NewClient(plaid.ClientOptions{
		PLAID_CLIENT_ID,
		PLAID_SECRET,
		PLAID_ENV,
		&http.Client{},
	})
	if err != nil {
		panic(fmt.Errorf("unexpected error while initializing plaid client %w", err))
	}
	return client
}()

var transactionsDb = func() *sql.DB {
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

	return db
}()

func queryAllAcctsById(user_id uint64) (*sql.Rows, error) {
	println("QUERY ALL USER'S ACCOUNTS")

	query_res, err := transactionsDb.Query(`SELECT institution_name, access_token FROM userInstitutionInfo 
												WHERE user_id = $1`, user_id)

	if err != nil {
		return nil, err
	}

	return query_res, nil
}

func insertNewAccount(user_id uint64, inst_id string, inst_name string, accessToken string) error {
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

func queryAccessToken(user_id uint64, inst_id string, inst_name string, accessToken *string) error {
	println("QUERY ACCESS TOKEN")
	query_res := transactionsDb.QueryRow(`SELECT access_token FROM userInstitutionInfo 
											WHERE user_id = $1
											AND institution_id = $2
											AND institution_name = $3`, user_id, inst_id, inst_name)

	return query_res.Scan(accessToken)
}

func queryUserIdByEmail(email string, queried_user_id *uint64) error {
	println("QUERY USER ID")
	query_res := transactionsDb.QueryRow("SELECT user_id FROM userInfo WHERE email = $1", email)

	return query_res.Scan(queried_user_id)
}

func insertNewUserReturnId(email string, queried_user_id *uint64) error {
	println("INSERT NEW USER")

	insert_res := transactionsDb.QueryRow("INSERT INTO userInfo (email) VALUES ($1) RETURNING user_id", email)
	err := insert_res.Scan(queried_user_id)

	if err != nil {
		return err
	}

	return nil
}

func queryLinkedUserAccts(email string) (map[string]string, error) {
	var user_id uint64 = 0
	var result = make(map[string]string)

	query_err := queryUserIdByEmail(email, &user_id)

	if query_err != nil {
		return nil, query_err
	}

	query_accts, query_accts_err := queryAllAcctsById(user_id)

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

// We store the access_token in memory - in production, store it in a secure
// persistent data store.
var accessToken string
var itemID string

type All_Linked_Accts_Req struct {
	Email string `form:"email" json:"email" binding:"required"`
}

type Public_Token_Req struct {
	Public_Token     string `form:"public_token" json:"public_token" binding:"required"`
	Email            string `form:"email" json:"email" binding:"required"`
	Institution_Id   string `form:"institution_id" json:"institution_id" binding:"required"`
	Institution_Name string `form:"institution_name" json:"institution_name" binding:"required"`
}

type Transactions_Req struct {
	Email            string `form:"email" json:"email" binding:"required"`
	Institution_Id   string `form:"institution_id" json:"institution_id" binding:"required"`
	Institution_Name string `form:"institution_name" json:"institution_name" binding:"required"`
	// StartDate         string
	// EndDate           string
	// Transaction_Types []string
}

func getAccessToken(c *gin.Context) {
	var accessTokExists bool = false
	var accessToken string = ""
	var queried_user_id uint64 = 0
	var httpErrorMsg string = ""

	pub_req := Public_Token_Req{}
	if err := c.ShouldBind(&pub_req); err != nil {
		fmt.Println(err.Error())
	}

	public_token := pub_req.Public_Token
	email := pub_req.Email
	institution_id := pub_req.Institution_Id
	institution_name := pub_req.Institution_Name

	fmt.Println("email: " + email)
	fmt.Println("public_token: " + public_token)
	fmt.Println("institution_id: " + institution_id)
	fmt.Println("institution_name: " + institution_name)

	query_err := queryUserIdByEmail(email, &queried_user_id)

	switch query_err {
	case nil:
		fmt.Println(queried_user_id)
		query_err = queryAccessToken(queried_user_id, institution_id, institution_name, &accessToken)
		if query_err != nil {
			fmt.Println(query_err.Error())
		}
		accessTokExists = accessToken != ""
	case sql.ErrNoRows:
		insert_err := insertNewUserReturnId(email, &queried_user_id)
		if insert_err != nil {
			fmt.Println(insert_err.Error())
		}
	default:
		fmt.Println(query_err.Error())
	}

	if !accessTokExists {
		response, err := client.ExchangePublicToken(public_token)
		if err != nil {
			fmt.Println(err.Error())
		}

		accessToken = response.AccessToken
		itemID = response.ItemID

		err = insertNewAccount(queried_user_id, institution_id, institution_name, accessToken)
	}

	fmt.Println("access token: " + accessToken)
	fmt.Println("item ID: " + itemID)

	if accessToken == "" {
		httpErrorMsg = "ERROR: unable to retrieve access token"
	}
	c.JSON(http.StatusOK, gin.H{
		"error": httpErrorMsg,
	})
}

func transactions(c *gin.Context) {
	var accessToken string = ""
	var user_id uint64 = 0

	trans_req := Transactions_Req{}
	if err := c.ShouldBind(&trans_req); err != nil {
		fmt.Println(err.Error())
	}

	email := trans_req.Email
	institution_id := trans_req.Institution_Id
	institution_name := trans_req.Institution_Name

	query_err := queryUserIdByEmail(email, &user_id)
	if query_err != nil {
		fmt.Println(query_err.Error())
	}

	query_err = queryAccessToken(user_id, institution_id, institution_name, &accessToken)
	if query_err != nil {
		fmt.Println(query_err.Error())
	}

	endDate := time.Now().Local().Format("2006-01-02")
	startDate := time.Now().Local().Add(-30 * 24 * time.Hour).Format("2006-01-02")

	response, err := client.GetTransactions(accessToken, startDate, endDate)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts":     response.Accounts,
		"transactions": response.Transactions,
	})
}

func getLinkedAccounts(c *gin.Context) {
	endDate := time.Now().Local().Format("2006-01-02")
	startDate := time.Now().Local().Add(-30 * 24 * time.Hour).Format("2006-01-02")

	linked_accts_req := All_Linked_Accts_Req{}
	response := make(map[string]interface{})
	resp_err := ""

	if err := c.ShouldBind(&linked_accts_req); err != nil {
		fmt.Println(err.Error())
	}

	email := linked_accts_req.Email

	res, err := queryLinkedUserAccts(email)
	if err != nil {
		fmt.Println(err.Error())
		resp_err = err.Error()
	}

	for inst_name, accessToken := range res {
		transactions, err := client.GetTransactions(accessToken, startDate, endDate)
		if err != nil {
			resp_err = "Could not fetch all accounts"
			continue
		}

		response[inst_name] = transactions
	}

	c.JSON(http.StatusOK, gin.H{
		"accountsInfo": response,
		"error":        resp_err,
	})
}

func auth(c *gin.Context) {
	response, err := client.GetAuth(accessToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response.Accounts,
		"numbers":  response.Numbers,
	})
}

func accounts(c *gin.Context) {
	response, err := client.GetAccounts(accessToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response.Accounts,
	})
}

func balance(c *gin.Context) {
	response, err := client.GetBalances(accessToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": response.Accounts,
	})
}

func item(c *gin.Context) {
	response, err := client.GetItem(accessToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	institution, err := client.GetInstitutionByID(response.Item.InstitutionID)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item":        response.Item,
		"institution": institution.Institution,
	})
}

func identity(c *gin.Context) {
	response, err := client.GetIdentity(accessToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"identity": response.Accounts,
	})
}

func investmentTransactions(c *gin.Context) {
	endDate := time.Now().Local().Format("2006-01-02")
	startDate := time.Now().Local().Add(-30 * 24 * time.Hour).Format("2006-01-02")
	response, err := client.GetInvestmentTransactions(accessToken, startDate, endDate)
	fmt.Println("error", err)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"investment_transactions": response,
	})
}

func holdings(c *gin.Context) {
	response, err := client.GetHoldings(accessToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"holdings": response,
	})
}

func info(context *gin.Context) {
	context.JSON(200, map[string]interface{}{
		"item_id":      itemID,
		"access_token": accessToken,
		"products":     strings.Split(PLAID_PRODUCTS, ","),
	})
}

func createLinkToken(c *gin.Context) {
	linkToken, err := linkTokenCreate(nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"link_token": linkToken})
}

type httpError struct {
	errorCode int
	error     string
}

func (httpError *httpError) Error() string {
	return httpError.error
}

// linkTokenCreate creates a link token using the specified parameters
func linkTokenCreate(
	paymentInitiation *plaid.PaymentInitiation,
) (string, *httpError) {
	countryCodes := strings.Split(PLAID_COUNTRY_CODES, ",")
	products := strings.Split(PLAID_PRODUCTS, ",")
	redirectURI := PLAID_REDIRECT_URI
	configs := plaid.LinkTokenConfigs{
		User: &plaid.LinkTokenUser{
			// This should correspond to a unique id for the current user.
			ClientUserID: "user-id",
		},
		ClientName:        "Plaid Quickstart",
		Products:          products,
		CountryCodes:      countryCodes,
		Language:          "en",
		RedirectUri:       redirectURI,
		PaymentInitiation: paymentInitiation,
	}
	resp, err := client.CreateLinkToken(configs)
	if err != nil {
		return "", &httpError{
			errorCode: http.StatusBadRequest,
			error:     err.Error(),
		}
	}
	return resp.LinkToken, nil
}

func assets(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "unfortunate the go client library does not support assets report creation yet."})
}

func setupHandlers(api *gin.RouterGroup) {
	api.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	api.POST("/info", info)
	api.GET("/oauth-response.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "oauth-response.html", gin.H{})
	})
	api.POST("/set_access_token", getAccessToken)
	api.GET("/auth", auth)
	api.GET("/accounts", accounts)
	api.GET("/balance", balance)
	api.GET("/item", item)
	api.POST("/item", item)
	api.GET("/identity", identity)
	api.GET("/transactions", transactions)
	api.POST("/transactions", transactions)
	api.POST("/create_link_token", createLinkToken)
	api.GET("/investment_transactions", investmentTransactions)
	api.GET("/holdings", holdings)
	api.GET("/assets", assets)
	api.POST("/all_linked_accounts", getLinkedAccounts)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	api := router.Group("/api")

	setupHandlers(api)

	err := router.Run(":" + APP_PORT)
	if err != nil {
		panic("unable to start server")
	}
}
