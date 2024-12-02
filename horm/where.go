package horm

import (
	"fmt"

	"github.com/horm-database/common/consts"
)

func (w Where) OrWhere(where Where, comment ...string) Where {
	key := consts.OR
	if len(comment) > 0 {
		key = fmt.Sprintf("%s #%s", consts.OR, comment[0])
	}

	if w == nil {
		return Where{
			key: where,
		}
	}

	w[key] = where
	return w
}

func (w Where) AndWhere(where Where, comment ...string) Where {
	key := consts.AND
	if len(comment) > 0 {
		key = fmt.Sprintf("%s #%s", consts.AND, comment[0])
	}

	if w == nil {
		return Where{
			key: where,
		}
	}

	w[key] = where
	return w
}

func (w Where) NotWhere(where Where, comment ...string) Where {
	key := consts.NOT
	if len(comment) > 0 {
		key = fmt.Sprintf("%s #%s", consts.NOT, comment[0])
	}

	if w == nil {
		return Where{
			key: where,
		}
	}

	w[key] = where
	return w
}
