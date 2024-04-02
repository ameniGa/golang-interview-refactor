package calculator

// Response is a type of response used by service layer
type Response struct {
	Code  int
	Data  interface{}
	Error error
}
