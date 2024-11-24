package horm

import (
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
)

// Type elastic search 版本 v7 以前有 type， v7之后 type 统一为 _doc
func (s *Query) Type(typ string) *Query {
	s.Unit.Type = typ
	return s
}

// ID 主键 _id 查询
func (s *Query) ID(value interface{}) *Query {
	return s.Eq("_id", value)
}

// Scroll 查询，size 为每次 scroll 大小，where 为 scroll 条件。
func (s *Query) Scroll(scroll string, size int, where ...Where) *Query {
	s.Op("scroll")

	if len(where) != 0 {
		s.Where(where[0])
	}

	if size <= 0 {
		s.Error = errs.New(errs.RetClientParamInvalid, "scroll size can`t be zero")
		return s
	}

	s.Limit(size)

	if s.Unit.Scroll == nil {
		s.Unit.Scroll = new(proto.Scroll)
	}

	s.Unit.Scroll.Info = scroll
	return s
}

// ScrollByID 根据 scrollID 滚动查询。
func (s *Query) ScrollByID(id string) *Query {
	s.Op("scroll")

	if s.Unit.Scroll == nil {
		s.Unit.Scroll = new(proto.Scroll)
	}

	s.Unit.Scroll.ID = id
	return s
}

// Refresh 更新数据立即刷新
func (s *Query) Refresh() *Query {
	s.SetParam("refresh", true)
	return s
}

// Routing 路由
func (s *Query) Routing(routing string) *Query {
	s.SetParam("routing", routing)
	return s
}

// HighLight 返回高亮
func (s *Query) HighLight(fields []string, preTag, postTag string) *Query {
	highLight := map[string]interface{}{}
	highLight["fields"] = fields
	highLight["pre_tag"] = preTag
	highLight["post_tag"] = postTag

	s.SetParam("highlight", highLight)

	return s
}

// Collapse collapse search results
func (s *Query) Collapse(field string) *Query {
	collapse := map[string]interface{}{}
	collapse["field"] = field
	s.SetParam("collapse", collapse)
	return s
}
