package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/arunsrinivas20/personal-finance-alerts/backend/db_actions"
	"github.com/arunsrinivas20/personal-finance-alerts/backend/msg_structs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/plaid/plaid-go/plaid"
)

type App_Info struct {
	PLAID_CLIENT_ID     string
	PLAID_SECRET        string
	PLAID_PRODUCTS      string
	PLAID_COUNTRY_CODES string
	PLAID_REDIRECT_URI  string
	APP_PORT            string
	PLAID_ENV           plaid.Environment
}

var (
	client   *plaid.Client
	app_info App_Info
)

func initClient() {
	cl, err := plaid.NewClient(plaid.ClientOptions{
		app_info.PLAID_CLIENT_ID,
		app_info.PLAID_SECRET,
		app_info.PLAID_ENV,
		&http.Client{},
	})
	if err != nil {
		panic(fmt.Errorf("unexpected error while initializing plaid client %w", err))
	}

	client = cl
}

func getAppInfo() {
	jsonFile, err := os.Open("app_info.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	json.Unmarshal(byteValue, &app_info)

	app_info.PLAID_ENV = plaid.Sandbox
}

// We store the access_token in memory - in production, store it in a secure
// persistent data store.
var accessToken string
var itemID string

func getAccessToken(c *gin.Context) {
	var accessTokExists bool = false
	var accessToken string = ""
	var queried_user_id uint64 = 0
	var httpErrorMsg string = ""

	pub_req := msg_structs.Public_Token_Req{}
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

	query_err := db_actions.QueryUserIdByEmail(email, &queried_user_id)

	switch query_err {
	case nil:
		fmt.Println(queried_user_id)
		query_err = db_actions.QueryAccessToken(queried_user_id, institution_id, institution_name, &accessToken)
		if query_err != nil {
			fmt.Println(query_err.Error())
		}
		accessTokExists = accessToken != ""
	case sql.ErrNoRows:
		insert_err := db_actions.InsertNewUserReturnId(email, &queried_user_id)
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

		err = db_actions.InsertNewAccount(queried_user_id, institution_id, institution_name, accessToken)
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

	trans_req := msg_structs.Transactions_Req{}
	if err := c.ShouldBind(&trans_req); err != nil {
		fmt.Println(err.Error())
	}

	email := trans_req.Email
	institution_id := trans_req.Institution_Id
	institution_name := trans_req.Institution_Name

	query_err := db_actions.QueryUserIdByEmail(email, &user_id)
	if query_err != nil {
		fmt.Println(query_err.Error())
	}

	query_err = db_actions.QueryAccessToken(user_id, institution_id, institution_name, &accessToken)
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

	linked_accts_req := msg_structs.All_Linked_Accts_Req{}
	response := make(map[string]interface{})
	resp_err := ""

	if err := c.ShouldBind(&linked_accts_req); err != nil {
		fmt.Println(err.Error())
	}

	email := linked_accts_req.Email

	res, err := db_actions.QueryLinkedUserAccts(email)
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
		"products":     strings.Split(app_info.PLAID_PRODUCTS, ","),
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
	countryCodes := strings.Split(app_info.PLAID_COUNTRY_CODES, ",")
	products := strings.Split(app_info.PLAID_PRODUCTS, ",")
	redirectURI := app_info.PLAID_REDIRECT_URI
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
	getAppInfo()

	initClient()

	db_actions.Init_Db()

	router := gin.Default()
	router.Use(cors.Default())

	api := router.Group("/api")

	setupHandlers(api)

	err := router.Run(":" + app_info.APP_PORT)
	if err != nil {
		panic("unable to start server")
	}
}
