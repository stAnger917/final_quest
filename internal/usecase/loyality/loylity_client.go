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
	a.logger.EasyLogInfo("accrual service", "request complete, got ", string(response.StatusCode))
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
	for {
		ctx := context.Background()
		a.logger.EasyLogInfo("accrual service", "starting accrual service, collecting orders", "")
		orderList, err := a.repository.GetAllOpenedOrders(ctx)
		if err != nil {
			fmt.Println("ERRROR")
			break
		}
		a.logger.EasyLogInfo("accrual service", "got order list for accrual pointing", "")
		if len(orderList) > 0 {
			for _, v := range orderList {
				// todo: make go func
				a.logger.EasyLogInfo("accrual service", "requesting info for: ", v)
				err = a.GetPointsInfoByOrder(ctx, v)
				if err != nil {
					a.logger.EasyLogError("accrual", "failed to get order info", v, err)
				}
			}
		}
		a.logger.EasyLogInfo("accrual service", "all job done - resting", "")
		time.Sleep(1800)
	}
}
