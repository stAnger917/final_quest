package server

import (
	"context"
	"errors"
	"final_quest/internal/errs"
	"final_quest/internal/models"
	"final_quest/pkg/authMW"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *AppHandler) UserRegistration(c *gin.Context) {
	var requestData models.RegistrationData
	err := h.jsonRegistrationRequestHandler(c, &requestData)
	if err != nil {
		if errors.Is(err, errs.ErrEmptyRegistrationData) {
			c.JSON(http.StatusBadRequest, map[string]string{"message": "empty email / password in request body denied"})
			return
		}
		c.JSON(http.StatusBadRequest, map[string]string{"message": "error while parsing request body"})
		return
	}

	err = h.userService.CreateNewUser(context.Context(c), requestData.Login, requestData.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, map[string]string{"message": errs.ErrUserAlreadyExists.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, map[string]string{"status": "created"})
	return
}

func (h *AppHandler) UserLogin(c *gin.Context) {
	var requestData models.RegistrationData
	err := h.jsonRegistrationRequestHandler(c, &requestData)
	if err != nil {
		if errors.Is(err, errs.ErrEmptyRegistrationData) {
			c.JSON(http.StatusBadRequest, map[string]string{"message": "empty email / password in request body denied"})
			return
		}
		c.JSON(http.StatusBadRequest, map[string]string{"message": "error while parsing request body"})
		return
	}
	err = h.userService.LoginUser(context.Context(c), requestData.Login, requestData.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]string{"message": errs.ErrLoginMismatch.Error()})
		return
	}
	userID, err := h.userService.GetUserID(context.Context(c), requestData.Login)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	token, _ := authMW.CreateToken(userID)
	c.SetCookie("session_token", token, 60*60*24, "", "localhost", false, false)
	c.JSON(http.StatusOK, map[string]string{"message": "welcome"})
	return
}

func (h *AppHandler) PostOrders(c *gin.Context) {
	orderNumber, err := h.textPlainRequestHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "empty order number"})
		return
	}
	token, err := c.Cookie("session_token")
	if err != nil {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}
	userID := authMW.GetLoginFromToken(token)
	err = h.userService.SaveOrderNumber(context.Context(c), userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidOrderNumber):
			c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "invalid order`s number"})
			return
		case errors.Is(err, errs.ErrOrderAlreadyExists):
			c.JSON(http.StatusOK, map[string]string{"status": "already uploaded!"})
			return
		case errors.Is(err, errs.ErrOrderBelongsToAnotherUser):
			c.JSON(http.StatusConflict, map[string]string{"error": "already uploaded by another user!"})
			return
		default:
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
	c.JSON(http.StatusAccepted, map[string]string{"status": "ok"})
	return
}

func (h *AppHandler) GetOrders(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err != nil {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}
	userID := authMW.GetLoginFromToken(token)
	res, err := h.userService.GetUserOrders(context.Context(c), userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if len(res) == 0 {
		c.String(http.StatusNoContent, "")
		return
	}
	c.JSON(http.StatusOK, res)
	return
}

func (h *AppHandler) GetBalance(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err != nil {
		c.String(http.StatusUnauthorized, err.Error())
		return
	}
	userID := authMW.GetLoginFromToken(token)
	res, err := h.userService.GetUserBalance(context.Context(c), userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
	return
}
