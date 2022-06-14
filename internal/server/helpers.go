package server

import (
	"final_quest/internal/errs"
	"final_quest/internal/models"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func (h *AppHandler) jsonRegistrationRequestHandler(c *gin.Context, data *models.RegistrationData) error {
	if err := c.ShouldBindJSON(&data); err != nil {
		h.logger.EasyLogError("handlers", "error while parsing request body", "", err)
		return err
	}
	if data.Login == "" || data.Password == "" {
		h.logger.EasyLogError("handlers", "error while parsing request body", "", errs.ErrEmptyRegistrationData)
		return errs.ErrEmptyRegistrationData
	}
	return nil
}

func (h *AppHandler) textPlainRequestHandler(c *gin.Context) (string, error) {
	reqContentType := c.Request.Header.Get("Content-Type")
	if reqContentType != "text/plain" {
		return "", errs.ErrIncorrectOrderBody
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}
	result := string(body)
	return result, nil
}

func (h *AppHandler) jsonWithdrawRequestHandler(c *gin.Context, data *models.WithdrawData) error {
	if err := c.ShouldBindJSON(&data); err != nil {
		h.logger.EasyLogError("handlers", "error while parsing request body", "", err)
		return err
	}
	if data.Order == "" || data.Sum == 0 {
		h.logger.EasyLogError("handlers", "error while parsing request body", "", errs.ErrEmptyRegistrationData)
		return errs.ErrIncorrectWithdrawReqBody
	}
	return nil
}
