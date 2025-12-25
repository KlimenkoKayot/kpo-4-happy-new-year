package account

type Money float64

func NewMoney(amount float64) (Money, error) {
	if amount < 0 {
		return 0, ErrInvalidAmount
	}
	return Money(amount), nil
}

func (m Money) Add(other Money) Money {
	return m + other
}

func (m Money) Subtract(other Money) Money {
	return m - other
}

func (m Money) CanSubtract(other Money) bool {
	return m >= other
}

func (m Money) Float64() float64 {
	return float64(m)
}
