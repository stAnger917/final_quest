package loyality

import (
	"context"
	"encoding/json"
	"final_quest/internal/errs"
	"final_quest/internal/models"
	"final_quest/internal/repository"
	"final_quest/pkg/logging"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type AccountingService struct {
	repository           *repository.AppRepo
	logger               *logging.Logger
	accountingServiceURL string
}

func NewAccountingService(repo *repository.AppRepo, logger *logging.Logger, URL string) *AccountingService {
	return &AccountingService{
		repository:           repo,
		logger:               logger,
		accountingServiceURL: URL,
	}
}

func (a *AccountingService) GetPointsInfoByOrder(ctx context.Context, order string) error {
	requestURL := a.accountingServiceURL + fmt.Sprintf("/api/orders/%s", order)
	response, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusTooManyRequests {
		return errs.ErrToManyRequests
	}
	if response.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				a.logger.EasyLogError("accounting service", "failed to close response.Body", "", err)
			}
		}(response.Body)
		if err != nil {
			return err
		}
		var orderInfo models.OrderAccountingInfo
		err = json.Unmarshal(body, &orderInfo)
		if err != nil {
			return err
		}
		err = a.HandleOrderInfo(ctx, orderInfo)
		if err != nil {
			return err
		}
	}
	return err
}

func (a *AccountingService) HandleOrderInfo(ctx context.Context, orderData models.OrderAccountingInfo) error {
	if orderData.Accrual != 0 {
		userID, err := a.repository.GetUserIDByOrderNum(ctx, orderData.Order)
		if err != nil {
			return err
		}
		err = a.repository.AddAccrualPoints(ctx, userID.UserID, orderData.Accrual)
		if err != nil {
			return err
		}
	}
	err := a.repository.ChangeOrderStatusByOrderNum(ctx, orderData.Order, orderData.Status)
	return err
}

func (a *AccountingService) RunAccountingService() {
	ctx := context.Background()
	for {
		orderList, err := a.repository.GetAllOpenedOrders(ctx)
		if err != nil {
			break
		}
		if len(orderList) > 0 {
			for _, v := range orderList {
				// todo: make go func
				err = a.GetPointsInfoByOrder(ctx, v)
				if err != nil {
					a.logger.EasyLogError("accrual", "failed to get order info", v, err)
				}
			}
		} else {
			time.Sleep(180000)
		}
	}
}
