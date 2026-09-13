// Package auth implements the authentication module.
package auth

// MiniAppTelegramRequest is the request of the Telegram mini-app endpoint.
// init_data is ignored by the stub and will be validated later.
type MiniAppTelegramRequest struct {
	InitData string `json:"init_data" example:"query_id=AAHdF6IQAAAAAN0XohDhrOrc"`
}

// MiniAppTelegramResponse carries the issued token.
type MiniAppTelegramResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}
