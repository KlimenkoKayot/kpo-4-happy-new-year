package order

type DomainService struct{}

func NewDomainService() *DomainService {
	return &DomainService{}
}

func (s *DomainService) CanCreateOrder(userID string, amount float64) error {
	if userID == "" {
		return ErrInvalidUserID
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}
