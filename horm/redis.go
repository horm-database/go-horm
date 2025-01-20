// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package horm

import (
	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/types"
	"github.com/horm-database/go-horm/horm/codec"
)

// Prefix redis 操作
func (s *Query) Prefix(prefix string) *Query {
	s.Unit.Prefix = prefix
	return s
}

// Expire 设置 key 的过期时间，key 过期后将不再可用。单位以秒计。
// param: key string
// param: int seconds 到期时间
func (s *Query) Expire(key string, seconds int) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("EXPIRE")
	s.SetKey(key)
	s.SetParam("seconds", seconds)
	return s
}

// TTL 以秒为单位返回 key 的剩余过期时间。
// param: string key
func (s *Query) TTL(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("TTL")
	s.SetKey(key)
	return s
}

// Exists 查看值是否存在 exists
// param: key string
func (s *Query) Exists(key string) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("EXISTS")
	s.SetKey(key)
	return s
}

// Del 删除已存在的键。不存在的 key 会被忽略。
// param: key string
func (s *Query) Del(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("DEL")
	s.SetKey(key)
	return s
}

// Set 设置给定 key 的值。如果 key 已经存储其他值， Set 就覆写旧值。
// param: key string
// param: val interface{} 任意类型数据
// param: params ...interface{} SET 其他参数:
// 包含 [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
func (s *Query) Set(key string, val interface{}, params ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SET")
	s.SetKey(key)
	s.SetVal(val)
	s.setRedisParams(consts.SetParams, params...)
	return s
}

// SetEX 指定的 key 设置值及其过期时间。如果 key 已经存在， SETEX 命令将会替换旧的值。
// param: key string
// param: val interface{} 任意类型数据
// param: seconds int 到期时间
func (s *Query) SetEX(key string, val interface{}, seconds int) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("SETEX")
	s.SetKey(key)
	s.SetVal(val)
	s.SetParam("seconds", seconds)
	return s
}

// SetNX redis.SetNX
// 指定的 key 不存在时，为 key 设置指定的值。
// param: key string
// param: val interface{} 任意类型数据
func (s *Query) SetNX(key string, val interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SETNX")
	s.SetKey(key)
	s.SetVal(val)
	return s
}

// Get 获取指定 key 的值。如果 key 不存在，返回 nil 。可用 IsNil(err) 判断是否key不存在，如果key储存的值不是字符串类型，返回一个错误。
// param: key string
func (s *Query) Get(key string) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("GET")
	s.SetKey(key)
	return s
}

// GetSet 设置给定 key 的值。如果 key 已经存储其他值， GetSet 就覆写旧值，并返回原来的值，如果原来未设置值，则返回报错 nil returned
// param: key string
// param: val interface{} 任意类型数据
func (s *Query) GetSet(key string, val interface{}) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("GETSET")
	s.SetKey(key)
	s.SetVal(val)
	return s
}

// Incr 将 key 中储存的数字值增一。 如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCR 操作。 如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
func (s *Query) Incr(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("INCR")
	s.SetKey(key)
	return s
}

// Decr 将 key 中储存的数字值减一。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 DECR 操作。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
func (s *Query) Decr(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("DECR")
	s.SetKey(key)
	return s
}

// IncrBy 将 key 中储存的数字加上指定的增量值。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCRBY 命令。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
// param: incr string 自增数量
func (s *Query) IncrBy(key string, incr int) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("INCRBY")
	s.SetKey(key)
	s.SetParam("increment", incr)
	return s
}

// MSet 批量设置一个或多个 key-val 对
// param: val hval 必须是 map 或者 struct
// 注意，本接口 Prefix 一定要在 MSet 之前设置，这里所有的 key 都会被加上 Prefix
func (s *Query) MSet(val interface{}) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("MSET")
	s.setRedisMap(val)
	return s
}

// MGet 返回多个 key 的 val
// param: keys string
func (s *Query) MGet(keys ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("MGET")
	for _, v := range keys {
		s.Unit.Args = append(s.Unit.Args, types.ToString(v))
	}
	return s
}

// SetBit 设置或清除指定偏移量上的位
// param: key string
// param: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)
// param: value int 1-设置, 0-清除
func (s *Query) SetBit(key string, offset uint32, value int) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SETBIT")
	s.SetKey(key)

	if value != 0 && value != 1 {
		s.Error = errs.Newf(errs.ErrReqParamInvalid, "SETBIT value must be 0 or 1")
		return s
	}

	s.SetParam("offset", offset)
	s.SetParam("value", value)
	return s
}

// GetBit 获取指定偏移量上的位
// param: key string
// param: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)
func (s *Query) GetBit(key string, offset uint32) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("GETBIT")
	s.SetKey(key)
	s.SetParam("offset", offset)
	return s
}

// BitCount 计算给定字符串中，被设置为 1 的比特位的数量
// param: key string
// param: start int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推
// param: end int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推
// param: [BYTE | BIT]
func (s *Query) BitCount(key string, params ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("BITCOUNT")
	s.SetKey(key)
	s.setRedisParams(consts.BitCountParams, params...)
	return s
}

// HSet 为哈希表中的字段赋值 。
// param: key string
// param: field interface{} 其中field建议为字符串,可以为整数，浮点数
// param: val interface{} 任意类型数据
// param: kvs ...interface{} 多条数据，按照filed,val 的格式，其中field建议为字符串,可以为整数，浮点数
func (s *Query) HSet(key string, field, val interface{}, kvs ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("HSET")
	s.SetKey(key)

	fieldStr := types.ToString(field)
	v, _ := s.GetCoder().Encode(codec.EncodeTypeRedisVal, val)

	if len(kvs) == 0 {
		s.SetField(fieldStr)
		s.SetVal(v)
		return s
	}

	// 多个值则放在 data 里面
	s.Unit.Data = map[string]interface{}{}
	s.Unit.Data[fieldStr] = v

	if len(kvs)%2 != 0 {
		s.Error = errs.New(errs.ErrReqParamInvalid, "number of values not a multiple of 2")
		return s
	}

	for i := 0; i < len(kvs); i += 2 {
		value, _ := s.GetCoder().Encode(codec.EncodeTypeRedisVal, kvs[i+1])
		s.Unit.Data[types.ToString(kvs[i])] = value
	}

	return s
}

// HSetNx 为哈希表中不存在的的字段赋值 。
// param: key string
// param: field string
// param: val interface{}
func (s *Query) HSetNx(key string, field interface{}, val interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("HSETNX")
	s.SetKey(key)
	s.SetField(types.ToString(field))
	v, _ := s.GetCoder().Encode(codec.EncodeTypeRedisVal, val)
	s.SetVal(v)
	return s
}

// HmSet 把 map/struct 数据设置到哈希表中。此命令会覆盖哈希表中已存在的字段。如果哈希表不存在，会创建一个空哈希表，并执行 HMSET 操作。
// param: key string
// param: hval 必须是 map 或者 struct
func (s *Query) HmSet(key string, hval interface{}) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("HMSET")
	s.SetKey(key)
	s.setRedisMap(hval)
	return s
}

// HIncrBy 为哈希表中的字段值加上指定增量值。
// param: key string
// param: field string
// param: incr string 自增数量
func (s *Query) HIncrBy(key string, field interface{}, incr int) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HINCRBY")
	s.SetKey(key)
	s.SetField(types.ToString(field))
	s.SetParam("increment", incr)
	return s
}

// HIncrByFloat 为哈希表中的字段值加上指定增量浮点数。
// param: key string
// param: field string
// param: incr float64 自增数量
func (s *Query) HIncrByFloat(key string, field interface{}, incr float64) *Query {
	s.ResultType = consts.RedisRetTypeFloat64
	s.Op("HINCRBYFLOAT")
	s.SetKey(key)
	s.SetField(types.ToString(field))
	s.SetParam("increment", incr)
	return s
}

// HGet 数据从redis hget 出来之后反序列化并赋值给 val
// param: key string
// param: field string
func (s *Query) HGet(key string, field interface{}) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("HGET")
	s.SetKey(key)
	s.SetField(types.ToString(field))
	return s
}

// HmGet 返回哈希表中，一个或多个给定字段的值。
// param: key string
// param: fields string 需要返回的域
func (s *Query) HmGet(key string, fields ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeMapString
	s.Op("HMGET")
	s.SetKey(key)

	for _, v := range fields {
		s.Unit.Args = append(s.Unit.Args, types.ToString(v))
	}

	return s
}

// HGetAll 返回哈希表中，所有的字段和值。
// param: key string
func (s *Query) HGetAll(key string) *Query {
	s.ResultType = consts.RedisRetTypeMapString
	s.Op("HGETALL")
	s.SetKey(key)
	return s
}

// HDel 删除哈希表 key 中的一个或多个指定字段，不存在的字段将被忽略。
// param: key string
// param: fields interface{}，删除指定key的fields数据，多个field，至少得有一个field
func (s *Query) HDel(key string, fields ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HDEL")
	s.SetKey(key)

	for _, v := range fields {
		s.Unit.Args = append(s.Unit.Args, types.ToString(v))
	}

	return s
}

// HExists 查看哈希表的指定字段是否存在。
// param: key string
// param: field interface{}
func (s *Query) HExists(key string, field interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("HEXISTS")
	s.SetKey(key)
	s.SetField(types.ToString(field))
	return s
}

// HLen 获取哈希表中字段的数量。
// param: key string
func (s *Query) HLen(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HLEN")
	s.SetKey(key)

	return s
}

// HStrLen 获取哈希表某个字段长度。
// param: key string
// param: field string
func (s *Query) HStrLen(key string, field interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HSTRLEN")
	s.SetKey(key)
	s.SetField(types.ToString(field))

	return s
}

// Hkeys 获取哈希表中的所有域（field）。
// param: key string
func (s *Query) Hkeys(key string) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("HKEYS")
	s.SetKey(key)
	return s
}

// HVals 返回所有的 val
// param: key string
func (s *Query) HVals(key string) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("HVALS")
	s.SetKey(key)

	return s
}

// LPush 将一个或多个值插入到列表头部。 如果 key 不存在，一个空列表会被创建并执行 LPUSH 操作。 当 key 存在但不是列表类型时，返回一个错误。
// param: key string
// param: values interface{} 任意类型数据
func (s *Query) LPush(key string, values ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("LPUSH")
	s.SetKey(key)
	s.addValues(values)
	return s
}

// RPush 将一个或多个值插入到列表的尾部(最右边)。如果列表不存在，一个空列表会被创建并执行 RPUSH 操作。 当列表存在但不是列表类型时，返回一个错误。
// param: key string
// param: v interface{} 任意类型数据
func (s *Query) RPush(key string, values ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("RPUSH")
	s.SetKey(key)
	s.addValues(values)
	return s
}

// LPop 移除并返回列表的第一个元素。
// param: key string
// param: count int 可选，移除并返回列表元素个数，如果不输入，返回是字符串，否则返回字符串数组
func (s *Query) LPop(key string, count ...int) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("LPOP")
	s.SetKey(key)
	if len(count) > 0 {
		s.ResultType = consts.RedisRetTypeStrings
		s.SetParam("count", count[0])
	}
	return s
}

// RPop 移除列表的最后一个元素，返回值为移除的元素。
// param: key string
// param: count int 可选，移除并返回列表元素个数，如果不输入，返回是字符串，否则返回字符串数组
func (s *Query) RPop(key string, count ...int) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("RPOP")
	s.SetKey(key)

	if len(count) > 0 {
		s.ResultType = consts.RedisRetTypeStrings
		s.SetParam("count", count[0])
	}
	return s
}

// LLen 返回列表的长度。 如果列表 key 不存在，则 key 被解释为一个空列表，返回 0 。 如果 key 不是列表类型，返回一个错误。
// param: key string
func (s *Query) LLen(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("LLEN")
	s.SetKey(key)
	return s
}

// SAdd 将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。
// param: key string
// param: values ...interface{} 任意类型的多条数据
func (s *Query) SAdd(key string, members ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SADD")
	s.SetKey(key)
	s.addValues(members)
	return s
}

// SMembers 返回集合中的所有的成员。 不存在的集合 key 被视为空集合。
// param: key string
func (s *Query) SMembers(key string) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("SMEMBERS")
	s.SetKey(key)
	return s
}

// SRem 移除集合中的一个或多个成员元素，不存在的成员元素会被忽略
// param: key string
// param: members ...interface{} 任意类型的多条数据
func (s *Query) SRem(key string, members ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SREM")
	s.SetKey(key)
	s.addValues(members)
	return s
}

// SCard 返回集合中元素的数量。
// param: key string
func (s *Query) SCard(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SCARD")
	s.SetKey(key)

	return s
}

// SIsMember 判断成员元素是否是集合的成员。
// param: key string
// param: member interface{} 要检索的任意类型数据
func (s *Query) SIsMember(key string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SISMEMBER")
	s.SetKey(key)
	s.SetVal(member)
	return s
}

// SRandMember 返回集合中的count个随机元素。
// param: key string
// param: count int 随机返回元素个数。如果不输入，返回是字符串，否则返回字符串数组
// 如果 count 为正数，且小于集合基数，那么命令返回一个包含 count 个元素的数组，数组中的元素各不相同。
// 如果 count 大于等于集合基数，那么返回整个集合。
// 如果 count 为负数，那么命令返回一个数组，数组中的元素可能会重复出现多次，而数组的长度为 count 的绝对值。
func (s *Query) SRandMember(key string, count ...int) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("SRANDMEMBER")
	s.SetKey(key)

	if len(count) > 0 {
		s.ResultType = consts.RedisRetTypeStrings
		s.SetParam("count", count[0])
	}
	return s
}

// SPop 移除集合中的指定 key 的一个或多个随机成员，移除后会返回移除的成员。
// param: key string
// param: int count，可选，如果不输入，返回是字符串，否则返回字符串数组
func (s *Query) SPop(key string, count ...int) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("SPOP")
	s.SetKey(key)

	if len(count) > 0 {
		s.ResultType = consts.RedisRetTypeStrings
		s.SetParam("count", count[0])
	}

	return s
}

// SMove 将指定成员 member 元素从 source 集合移动到 destination 集合。
// param: source string
// param: destination string
// param: member interface{} 要移动的成员，任意类型
func (s *Query) SMove(source, destination string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SMOVE")
	s.SetKey(source)
	s.SetVal(member)
	s.SetParam("destination", destination)
	return s
}

// ZAdd redis.ZAdd
// 将成员元素及其分数值加入到有序集当中。如果某个成员已经是有序集的成员，那么更新这个成员的分数值，并通过重新插入这个成员元素，
// 来保证该成员在正确的位置上。分数值可以是整数值或双精度浮点数。
// param: key string
// param: args ...interface{} 添加更多成员，需要按照  member, score, member, score 依次排列
// 注意：⚠️ 与 redis 命令不一样，需要按照  member, score, member, score, 格式传入
func (s *Query) ZAdd(key string, args ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZADD")
	s.SetKey(key)

	if len(args) < 2 || len(args)%2 != 0 {
		s.Error = errs.Newf(errs.ErrReqParamInvalid, "ZADD args should contain pair of memhber and score")
		return s
	}

	values := []interface{}{}
	scores := []interface{}{}

	for k, v := range args {
		if k%2 == 0 {
			values = append(values, v)
		} else {
			scores = append(scores, v)
		}
	}

	s.addValues(values)

	if s.Unit.Params == nil {
		s.Unit.Params = map[string]interface{}{}
	}

	if len(scores) == 1 {
		s.Unit.Params["score"] = scores[0]
	} else {
		s.Unit.Params["scores"] = scores
	}

	return s
}

// ZRem 移除有序集中的一个或多个成员，不存在的成员将被忽略。
// param: key string
// param: members ...interface{} 任意类型的多条数据
func (s *Query) ZRem(key string, members ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREM")
	s.SetKey(key)
	s.addValues(members)
	return s
}

// ZRemRangeByScore 移除有序集中，指定分数（score）区间内的所有成员。
// param: key string
// param: interface{} min max 分数区间，类型为整数或者浮点数
func (s *Query) ZRemRangeByScore(key string, min, max interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREMRANGEBYSCORE")
	s.SetKey(key)
	s.SetParam("min", min)
	s.SetParam("max", max)
	return s
}

// ZRemRangeByRank 移除有序集中，指定排名(rank)区间内的所有成员。
// param: key string
// param: start stop int 排名区间
func (s *Query) ZRemRangeByRank(key string, start, stop int) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREMRANGEBYRANK")
	s.SetKey(key)
	s.SetParam("start", start)
	s.SetParam("stop", stop)
	return s
}

// ZCard 返回有序集成员个数
// param: key string
func (s *Query) ZCard(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZCARD")
	s.SetKey(key)
	return s
}

// ZScore 返回有序集中，成员的分数值。
// param: key string
// param: member interface{} 成员
func (s *Query) ZScore(key string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeFloat64
	s.Op("ZSCORE")
	s.SetKey(key)
	s.SetVal(member)
	return s
}

// ZRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。
// param: key string
// param: member interface{} 成员，任意类型
func (s *Query) ZRank(key string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZRANK")
	s.SetKey(key)
	s.SetVal(member)
	return s
}

// ZRevRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从大到小)顺序排列。
// param: key string
// param: member interface{} 成员，任意类型
func (s *Query) ZRevRank(key string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREVRANK")
	s.SetKey(key)
	s.SetVal(member)
	return s
}

// ZCount 计算有序集合中指定分数区间的成员数量
// param: key string
// param: min interface{}
// param: max interface{}
func (s *Query) ZCount(key string, min, max interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZCOUNT")
	s.SetKey(key)
	s.SetParam("min", min)
	s.SetParam("max", max)
	return s
}

// ZPopMin 移除并弹出有序集合中分值最小的 count 个元素
// redis v5.0.0+
// param: key string
// param: count ...int64 不设置count参数时，弹出一个元素
func (s *Query) ZPopMin(key string, count ...int64) *Query {
	s.ResultType = consts.RedisRetTypeMemberScore
	s.Op("ZPOPMIN")
	s.SetKey(key)

	if len(count) > 0 {
		s.SetParam("count", count[0])
	}

	return s
}

// ZPopMax 移除并弹出有序集合中分值最大的的 count 个元素
// redis v5.0.0+
// param: key string
// param: count ...int64 不设置count参数时，弹出一个元素
func (s *Query) ZPopMax(key string, count ...int64) *Query {
	s.ResultType = consts.RedisRetTypeMemberScore
	s.Op("ZPOPMAX")
	s.SetKey(key)

	if len(count) > 0 {
		s.SetParam("count", count[0])
	}

	return s
}

// ZIncrBy 对有序集合中指定成员的分数加上增量 increment，可以通过传递一个负数值 increment ，
// 让分数减去相应的值，比如 ZINCRBY key -5 member ，
// 就是让 member 的 score 值减去 5 。当 key 不存在，或分数不是 key 的成员时，
// ZINCRBY key increment member 等同于 ZADD key
// increment member 。当 key 不是有序集类型时，返回一个错误。分数值可以是整数值或双精度浮点数。
// param: key string
// param: member interface{} 任意类型数据
// param: incr interface{} 增量值，可以为整数或双精度浮点
func (s *Query) ZIncrBy(key string, member, incr interface{}) *Query {
	s.ResultType = consts.RedisRetTypeFloat64
	s.Op("ZINCRBY")
	s.SetKey(key)
	s.SetVal(member)
	s.SetParam("increment", incr)
	return s
}

// ZRange 返回有序集中，指定区间内的成员。其中成员的位置按分数值递增(从小到大)来排序。
// param: key string
// param: int start, stop 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
// param: args [BYSCORE | BYLEX] [REV] [LIMIT offset count] [WITHSCORES]
// WITHSCORES 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
func (s *Query) ZRange(key string, start, stop interface{}, args ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	s.Op("ZRANGE")
	s.SetKey(key)

	params := []interface{}{"start", start, "stop", stop}
	params = append(params, args...)
	s.setRedisParams(consts.ZRangeParams, params...)

	withScores, _ := types.GetBool(s.Unit.Params, "WITHSCORES")
	if withScores {
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	return s
}

// ZRangeByScore 根据分数返回有序集中指定区间的成员，顺序从小到大
// param: key string
// param: int min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
// param: WITHSCORES 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// param: LIMIT信息，包含 offset count
func (s *Query) ZRangeByScore(key string, min, max interface{}, args ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	s.Op("ZRANGEBYSCORE")
	s.SetKey(key)

	params := []interface{}{"min", min, "max", max}
	params = append(params, args...)
	s.setRedisParams(consts.ZRangeByScoreParams, params...)

	withScores, _ := types.GetBool(s.Unit.Params, "WITHSCORES")
	if withScores {
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	return s
}

// ZRevRange 返回有序集中指定区间的成员，其中成员的位置按分数值递减(从大到小)来排列。
// param: key string
// param: start, stop 排名区间，以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
// param: WITHSCORES 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
func (s *Query) ZRevRange(key string, start, stop int, args ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	s.Op("ZREVRANGE")
	s.SetKey(key)

	params := []interface{}{"start", start, "stop", stop}
	params = append(params, args...)
	s.setRedisParams(consts.ZRevRangeParams, params...)

	withScores, _ := types.GetBool(s.Unit.Params, "WITHSCORES")
	if withScores {
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	return s
}

// ZRevRangeByScore 返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。
// param: key string
// param: max, min  interface{} 分数区间，类型为整数或双精度浮点数，但是 -inf +inf 表示负正无穷大
// param: WITHSCORES 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// param: LIMIT 信息，包含 offset count
func (s *Query) ZRevRangeByScore(key string, max, min interface{}, args ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	s.Op("ZREVRANGEBYSCORE")
	s.SetKey(key)

	params := []interface{}{"max", max, "min", min}
	params = append(params, args...)
	s.setRedisParams(consts.ZRevRangeByScoreParams, params...)

	withScores, _ := types.GetBool(s.Unit.Params, "WITHSCORES")
	if withScores {
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	return s
}

func (s *Query) addValues(values []interface{}) *Query {
	if len(values) == 0 {
		s.Error = errs.Newf(errs.ErrReqParamInvalid, "%s empty value", s.Unit.Op)
		return s
	}

	if len(values) == 1 {
		s.SetVal(values[0])
	} else {
		for _, v := range values {
			ev, err := s.GetCoder().Encode(codec.EncodeTypeRedisVal, v)
			if err != nil {
				s.Error = err
				return s
			}
			s.Unit.Args = append(s.Unit.Args, ev)
		}
	}

	return s
}

func (s *Query) setRedisMap(val interface{}) *Query {
	m, err := types.ToMap(val, s.GetCoder().GetTag())
	if err != nil {
		s.Error = errs.Newf(errs.ErrReqParamInvalid, "mset error: %v", err)
		return s
	}

	for k, v := range m {
		ev, e := s.GetCoder().Encode(codec.EncodeTypeRedisVal, v)
		if e != nil {
			s.Error = e
			return s
		}
		m[k] = ev
	}

	s.Unit.Data = m
	return s
}

func (s *Query) setRedisParams(paramInfos []*consts.RedisParamInfo, params ...interface{}) {
	var jump int

	for k, v := range params {
		if jump > 0 {
			jump--
			continue
		}

		name, ok := v.(string)
		if !ok {
			s.Error = errs.Newf(errs.ErrReqParamInvalid, "redis cmd %s arg name must be string [%v]", s.Unit.Op, v)
			return
		}

		paramInfo, find := consts.FindRedisParam(paramInfos, name)
		if !find {
			s.Error = errs.Newf(errs.ErrReqParamInvalid, "redis cmd %s not support arg %s", s.Unit.Op, name)
			return
		}

		if s.Unit.Params == nil {
			s.Unit.Params = map[string]interface{}{}
		}

		switch paramInfo.Cnt {
		case 0, 1:
			s.Unit.Params[name] = true
		case 2:
			if len(params) < k+2 {
				s.Error = errs.Newf(errs.ErrReqParamInvalid, "redis cmd %s params number is invalid", s.Unit.Op)
				return
			}
			s.Unit.Params[name] = params[k+1]
			jump = 1
		default:
			if len(params) < k+paramInfo.Cnt {
				s.Error = errs.Newf(errs.ErrReqParamInvalid, "redis cmd %s params number is invalid", s.Unit.Op)
				return
			}

			tmp := []interface{}{}
			for i := 1; i < paramInfo.Cnt; i++ {
				tmp = append(tmp, params[k+i])
			}

			s.Unit.Params[name] = tmp
			jump = paramInfo.Cnt - 1
		}
	}

	return
}
