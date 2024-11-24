package horm

import (
	"github.com/horm-database/common/proto"
)

func createUnits(q *Query) ([]*proto.Unit, error) {
	units := make([]*proto.Unit, 0)

	err := addUnit(&units, q)
	if err != nil {
		return nil, err
	}

	return units, nil
}

func addUnit(units *[]*proto.Unit, q *Query) error {
	if q.Error != nil {
		return q.Error
	}

	if q.Unit.Size < 0 {
		q.Unit.Size = 0
	}

	*units = append(*units, q.Unit)

	if q.sub != nil {
		q.Unit.Sub = make([]*proto.Unit, 0)
		err := addUnit(&q.Unit.Sub, q.sub)
		if err != nil {
			return err
		}
	}

	if q.trans != nil {
		q.Unit.Trans = make([]*proto.Unit, 0)
		err := addUnit(&q.Unit.Trans, q.trans)
		if err != nil {
			return err
		}
	}

	if q.next != nil {
		err := addUnit(units, q.next)
		if err != nil {
			return err
		}
	}

	return nil
}
