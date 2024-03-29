package calculator

type Response struct {
	Code        int
	RedirectURL string
	Data        interface{}
	Error       error
}
