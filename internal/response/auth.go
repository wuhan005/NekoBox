package response

type SignIn struct {
	Profile *SignInUserProfile `json:"profile"`
	// SessionID is sent back as the `Authorization: Token <sessionID>` header to authenticate the later requests.
	SessionID string `json:"sessionID"`
} // @name SignInResp

type SignInUserProfile struct {
	UID    string `json:"uid"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
} // @name SignInUserProfile

type RecoverPassword struct {
	Name string `json:"name"`
} // @name RecoverPasswordResp
