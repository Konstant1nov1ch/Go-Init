package orm

import "encoding/json"

type (
	GenericID     *int
	PersInterface interface {
		String() string
		Name() string
		GenericID() GenericID
	}
)

func ModelToString(model any) string {
	res, err := json.Marshal(model)
	if err != nil {
		return ``
	}
	return string(res)
}
