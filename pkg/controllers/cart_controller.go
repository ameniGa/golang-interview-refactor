package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"interview/pkg/calculator"
	"interview/pkg/helpers"
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
	cookie, err := c.Request.Cookie("ice_session_id")
	if errors.Is(err, http.ErrNoCookie) {
		c.SetCookie("ice_session_id", time.Now().String(), 3600, "/", "localhost", false, true)
	}

	data := map[string]interface{}{
		"Error": c.Query("error"),
	}

	res := t.calculator.GetCartData(c, cookie.Value)
	if res.Data != nil {
		data["CartItems"] = res.Data
	}

	html, err := helpers.RenderTemplate(data, "static/add_item_form.html")
	if err != nil {
		handleError(c, 500, err)
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(200, html)
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
	handleError(c, res.Code, res.Error)
}

func (t *TaxController) DeleteCartItem(c *gin.Context) {
	cookie, err := c.Request.Cookie("ice_session_id")
	if err != nil || errors.Is(err, http.ErrNoCookie) || (cookie != nil && cookie.Value == "") {
		c.Redirect(302, "/")
		return
	}

	cartItemIDString := c.Query("cart_item_id")
	if cartItemIDString == "" {
		c.Redirect(302, "/")
		return
	}

	res := t.calculator.DeleteCartItem(c, cookie.Value, cartItemIDString)
	handleError(c, res.Code, res.Error)
}

func handleError(c *gin.Context, code int, err error) {
	switch code {
	case 302:
		if err != nil {
			c.Redirect(302, fmt.Sprintf("/?error=%v", err.Error()))
			return
		}
		c.Redirect(302, "/")
		return
	case 500:
		c.AbortWithStatus(500)
		return
	}
}
