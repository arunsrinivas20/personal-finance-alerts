package msg_structs

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
