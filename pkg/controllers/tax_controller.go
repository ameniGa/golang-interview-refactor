package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"interview/pkg/calculator"
	"net/http"
	"time"
)

type TaxController struct {
	calculator calculator.Handler
}

func NewTaxController(calculator calculator.Handler) *TaxController {
	return &TaxController{calculator: calculator}
}

func (t *TaxController) ShowAddItemForm(c *gin.Context) {
	_, err := c.Request.Cookie("ice_session_id")
	if errors.Is(err, http.ErrNoCookie) {
		c.SetCookie("ice_session_id", time.Now().String(), 3600, "/", "localhost", false, true)
	}

	t.calculator.GetCartData(c)
}

func (t *TaxController) AddItem(c *gin.Context) {
	cookie, err := c.Request.Cookie("ice_session_id")

	if err != nil || errors.Is(err, http.ErrNoCookie) || (cookie != nil && cookie.Value == "") {
		c.Redirect(302, "/")
		return
	}

	if c.Request.Body == nil {
		c.Redirect(302, "/?error=body cannot be nil")
		return
	}

	form := &CartItemForm{}
	if err := binding.FormPost.Bind(c.Request, form); err != nil {
		c.Redirect(302, "/?error="+err.Error())
		return
	}

	res := t.calculator.AddItemToCart(c, cookie.Value, calculator.CartItem{
		Product:  form.Product,
		Quantity: form.Quantity,
	})
	c.Redirect(res.Code, res.RedirectURL)
}

func (t *TaxController) DeleteCartItem(c *gin.Context) {
	cookie, err := c.Request.Cookie("ice_session_id")

	if err != nil || errors.Is(err, http.ErrNoCookie) || (cookie != nil && cookie.Value == "") {
		c.Redirect(302, "/")
		return
	}

	t.calculator.DeleteCartItem(c)
}
