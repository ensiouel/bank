package sort

type Filter interface {
	GetSort() string
	GetOrder() string
	GetCount() uint64
	GetOffset() uint64
}

type Pagination struct {
	Count  uint64 `query:"count"`
	Offset uint64 `query:"offset"`
}

func (pagination Pagination) GetCount() uint64 {
	return pagination.Count
}

func (pagination Pagination) GetOffset() uint64 {
	return pagination.Offset
}
