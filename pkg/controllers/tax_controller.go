package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"html/template"
	"interview/pkg/calculator"
	"log"
	"net/http"
	"strings"
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
	fmt.Println(res)
	if res.Data != nil {
		data["CartItems"] = res.Data
	}

	html, err := renderTemplate(data)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(500)
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
	c.Redirect(res.Code, res.RedirectURL)
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
	c.Redirect(res.Code, res.RedirectURL)
}

func renderTemplate(pageData interface{}) (string, error) {
	// Read and parse the HTML template file
	tmpl, err := template.ParseFiles("static/add_item_form.html")
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Create a strings.Builder to store the rendered template
	var renderedTemplate strings.Builder

	err = tmpl.Execute(&renderedTemplate, pageData)
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Convert the rendered template to a string
	resultString := renderedTemplate.String()

	return resultString, nil
}
