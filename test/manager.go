package manager

var _ Manager = (*manager)(nil)

type Manager interface {
	i()

	// GetByID 通过ID查找
	GetByID(id int64) (string, error)

	// GetByName 通过名称查找
	GetByName(name string) (string, error)

	// UpdateNameByID 通过ID更新名称
	UpdateNameByID(id int64, name string)

	// ListAll 查找所有
	ListAll() ([]*string, error)
}

type manager struct {
}

func New() Manager {
	return &manager{}
}

func (m *manager) i() {}
