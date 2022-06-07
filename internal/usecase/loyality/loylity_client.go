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
	a.logger.EasyLogInfo("accrual service", "sending request for order: ", order)
	requestURL := a.accountingServiceURL + fmt.Sprintf("/api/orders/%s", order)
	response, err := http.Get(requestURL)
	a.logger.EasyLogInfo("accrual service", "request complete, got ", fmt.Sprintf("%v", response.StatusCode))
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
		a.logger.EasyLogInfo("accrual service", "got order data, now - handling info in db: ", fmt.Sprintf("%v", orderInfo))
		err = a.HandleOrderInfo(ctx, orderInfo)
		if err != nil {
			return err
		}
		a.logger.EasyLogInfo("accrual service", "order data handled in db", err.Error())
	}
	return err
}

func (a *AccountingService) HandleOrderInfo(ctx context.Context, orderData models.OrderAccountingInfo) error {
	if orderData.Accrual != 0 {
		a.logger.EasyLogInfo("accrual service", "request in db to get userID for order: ", orderData.Order)
		userID, err := a.repository.GetUserIDByOrderNum(ctx, orderData.Order)
		if err != nil {
			a.logger.EasyLogError("accrual", "failed to get userID for order: ", orderData.Order, err)
			return err
		}
		a.logger.EasyLogInfo("accrual service", "request in db to add points for user. Got data: ", fmt.Sprintf("userID: %v, accrual: %v", userID.UserID, orderData.Accrual))
		err = a.repository.AddAccrualPoints(ctx, userID.UserID, orderData.Accrual)
		if err != nil {
			a.logger.EasyLogError("accrual", "failed to add points to user", orderData.Order, err)
			return err
		}
	}
	a.logger.EasyLogInfo("accrual service", "request in db to change order status for order: ", orderData.Order)
	err := a.repository.ChangeOrderStatusByOrderNum(ctx, orderData.Order, orderData.Status)
	if err != nil {
		a.logger.EasyLogError("accrual", "failed to change order status", orderData.Order, err)
	}
	return err
}

func (a *AccountingService) RunAccountingService() {
	for {
		ctx := context.Background()
		a.logger.EasyLogInfo("accrual service", "starting accrual service, collecting orders", "")
		orderList, err := a.repository.GetAllOpenedOrders(ctx)
		if err != nil {
			fmt.Println(err)
			break
		}
		if len(orderList) > 0 {
			for _, v := range orderList {
				a.logger.EasyLogInfo("accrual service", "requesting info for: ", v)
				err = a.GetPointsInfoByOrder(ctx, v)
				if err != nil {
					a.logger.EasyLogError("accrual", "failed to get order info", "", err)
				}
			}
		}
		a.logger.EasyLogInfo("accrual service", "all job done - resting", "")
		time.Sleep(15 * time.Second)
	}
}
