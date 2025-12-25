package account

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (s *PaymentService) CanProcessPayment(account *Account, amount Money) error {
	if account == nil {
		return ErrAccountNotFound
	}
	if !account.CanWithdraw(amount) {
		return ErrInsufficientFunds
	}
	return nil
}
