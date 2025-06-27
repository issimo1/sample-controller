package dao

type Dao struct {
	Service *Service
}

func NewDao() *Dao {
	return &Dao{
		Service: &Service{},
	}
}
