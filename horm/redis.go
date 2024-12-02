package horm

import (
	"strconv"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/go-horm/horm/codec"
)

// Prefix redis 操作
func (s *Query) Prefix(prefix string) *Query {
	s.Unit.Prefix = prefix
	return s
}

// Expire 设置 key 的过期时间，key 过期后将不再可用。单位以秒计。
// param: key string
// param: int ttl 到期时间，ttl秒
func (s *Query) Expire(key string, ttl int) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("EXPIRE")
	s.setKey(key)
	s.append(ttl)
	return s
}

// TTL 以秒为单位返回 key 的剩余过期时间。
// param: string key
func (s *Query) TTL(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("TTL")
	s.setKey(key)
	return s
}

// Exists 查看值是否存在 exists
// param: key string
func (s *Query) Exists(key string) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("EXISTS")
	s.setKey(key)
	return s
}

// Del 删除已存在的键。不存在的 key 会被忽略。
// param: key string
func (s *Query) Del(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("DEL")
	s.setKey(key)
	return s
}

// Set 设置给定 key 的值。如果 key 已经存储其他值， Set 就覆写旧值。
// param: key string
// param: value interface{} 任意类型数据
// param: args ...interface{} set的其他参数
func (s *Query) Set(key string, value interface{}, args ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SET")
	s.setKey(key)
	s.append(s.structToMap(value)).append(args...)
	return s
}

// SetEX 指定的 key 设置值及其过期时间。如果 key 已经存在， SETEX 命令将会替换旧的值。
// param: key string
// param: v interface{} 任意类型数据
// param: ttl int 到期时间
func (s *Query) SetEX(key string, v interface{}, ttl int) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("SETEX")
	s.setKey(key)
	s.append(ttl, s.structToMap(v))
	return s
}

// SetNX redis.SetNX
// 指定的 key 不存在时，为 key 设置指定的值。
// param: key string
// param: v interface{} 任意类型数据
func (s *Query) SetNX(key string, v interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SETNX")
	s.setKey(key)
	s.append(s.structToMap(v))
	return s
}

// Get 获取指定 key 的值。如果 key 不存在，返回 nil 。可用 IsNil(err) 判断是否key不存在，如果key储存的值不是字符串类型，返回一个错误。
// param: key string
func (s *Query) Get(key string) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("GET")
	s.setKey(key)
	return s
}

// GetSet 设置给定 key 的值。如果 key 已经存储其他值， GetSet 就覆写旧值，并返回原来的值，如果原来未设置值，则返回报错 nil returned
// param: key string
// param: v interface{} 任意类型数据
func (s *Query) GetSet(key string, v interface{}) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("GETSET")
	s.setKey(key)
	s.append(s.structToMap(v))
	return s
}

// Incr 将 key 中储存的数字值增一。 如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCR 操作。 如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
func (s *Query) Incr(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("INCR")
	s.setKey(key)
	return s
}

// Decr 将 key 中储存的数字值减一。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 DECR 操作。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
func (s *Query) Decr(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("DECR")
	s.setKey(key)
	return s
}

// IncrBy 将 key 中储存的数字加上指定的增量值。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCRBY 命令。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
// param: n string 自增数量
func (s *Query) IncrBy(key string, n int) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("INCRBY")
	s.setKey(key)
	s.append(n)
	return s
}

// MSet 批量设置一个或多个 key-value 对
// param: values map[string]interface{} // value will marshal
// 注意，本接口 Prefix 一定要在 MSet 之前设置，这里所有的 key 都会被加上 Prefix
func (s *Query) MSet(values map[string]interface{}) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("MSET")

	var i int
	for k, v := range values {
		if i == 0 {
			s.setKey(k)
			s.append(s.structToMap(v))
		} else {
			if s.Unit.Prefix != "" {
				s.append(s.Unit.Prefix+k, s.structToMap(v))
			} else {
				s.append(k, s.structToMap(v))
			}
		}
		i++
	}
	return s
}

// MGet 返回多个 key 的 value
// param: keys string
func (s *Query) MGet(keys ...string) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("MGET")

	var i int
	for _, k := range keys {
		if i == 0 {
			s.setKey(k)
		} else {
			if s.Unit.Prefix != "" {
				s.append(s.Unit.Prefix + k)
			} else {
				s.append(k)
			}
		}
		i++
	}
	return s
}

// SetBit 设置或清除指定偏移量上的位
// param: key string
// param: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)
// param: value bool true:设置为1,false：设置为0
func (s *Query) SetBit(key string, offset uint32, value bool) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SETBIT")
	s.setKey(key)

	arg := 0
	if value {
		arg = 1
	}

	s.append(offset, arg)
	return s
}

// GetBit 获取指定偏移量上的位
// param: key string
// param: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)
func (s *Query) GetBit(key string, offset uint32) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("GETBIT")
	s.setKey(key)
	s.append(offset)
	return s
}

// BitCount 计算给定字符串中，被设置为 1 的比特位的数量
// param: key string
// param: start int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推
// param: end int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推
func (s *Query) BitCount(key string, start, end int) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("BITCOUNT")
	s.setKey(key)
	s.append(start, end)
	return s
}

// HSet 为哈希表中的字段赋值 。
// param: key string
// param: field interface{} 其中field建议为字符串,可以为整数，浮点数
// param: v interface{} 任意类型数据
// param: args ...interface{} 多条数据，按照filed,value 的格式，其中field建议为字符串,可以为整数，浮点数
func (s *Query) HSet(key string, field, v interface{}, args ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("HSET")
	s.setKey(key)
	s.append(field, s.structToMap(v)).append(args...)

	return s
}

// HSetNx 为哈希表中不存在的的字段赋值 。
// param: key string
// param: field string
// param: value interface{}
func (s *Query) HSetNx(key string, filed interface{}, value interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("HSETNX")
	s.setKey(key)
	s.append(filed, s.structToMap(value))

	return s
}

// HmSet 把 map/struct 数据设置到哈希表中。此命令会覆盖哈希表中已存在的字段。如果哈希表不存在，会创建一个空哈希表，并执行 HMSET 操作。
// param: key string
// param: v 必须是  map[string]interface{} 、或者 struct
func (s *Query) HmSet(key string, v interface{}) *Query {
	s.ResultType = consts.RedisRetTypeNil
	s.Op("HMSET")
	s.setKey(key)

	// struct/map 转化为 redis HMSET 的 field/value 对
	fieldValues, err := s.GetCoder().Encode(codec.EncodeTypeHmSET, v)
	if err != nil {
		s.Error = errs.Newf(errs.RetClientEncodeFail, "redis extend encode error: %v", err)
		return s
	}

	if fieldValues == nil {
		s.Error = errs.Newf(errs.RetClientEncodeFail, "redis extend encode value is nil")
		return s
	}

	tmp := fieldValues.([]interface{})

	return s.append(tmp...)
}

// HIncrBy 为哈希表中的字段值加上指定增量值。
// param: key string
// param: field string
// param: n string 自增数量
func (s *Query) HIncrBy(key string, field string, v int) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HINCRBY")
	s.setKey(key)
	s.append(field, v)

	return s
}

// HIncrByFloat 为哈希表中的字段值加上指定增量浮点数。
// param: key string
// param: field string
// param: v float64 自增数量
func (s *Query) HIncrByFloat(key string, field string, v float64) *Query {
	s.ResultType = consts.RedisRetTypeFloat64
	s.Op("HINCRBYFLOAT")
	s.setKey(key)
	s.append(field, v)

	return s
}

// HGet 数据从redis hget 出来之后反序列化并赋值给 v
// param: key string
// param: field string
func (s *Query) HGet(key string, field interface{}) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("HGET")
	s.setKey(key)
	s.append(field)
	return s
}

// HmGet 返回哈希表中，一个或多个给定字段的值。
// param: key string
// param: fields string 需要返回的域
func (s *Query) HmGet(key string, fields ...string) *Query {
	s.ResultType = consts.RedisRetTypeMapString
	s.Op("HMGET")
	s.setKey(key)

	args := redigo.Args{}.AddFlat(fields)
	s.append(args...)

	return s
}

// HGetAll 返回哈希表中，所有的字段和值。
// param: key string
func (s *Query) HGetAll(key string) *Query {
	s.ResultType = consts.RedisRetTypeMapString
	s.Op("HGETALL")
	s.setKey(key)
	return s
}

// HDel 删除哈希表 key 中的一个或多个指定字段，不存在的字段将被忽略。
// param: keyfield interface{}，删除指定key的field数据，这里输入的第一参数为key，其他为多个field，至少得有一个field
func (s *Query) HDel(key string, field ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HDEL")
	s.setKey(key)
	s.append(field...)

	return s
}

// HExists 查看哈希表的指定字段是否存在。
// param: key string
// param: field interface{}
func (s *Query) HExists(key string, field interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("HEXISTS")
	s.setKey(key)
	s.append(field)

	return s
}

// HLen 获取哈希表中字段的数量。
// param: key string
func (s *Query) HLen(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HLEN")
	s.setKey(key)

	return s
}

// HStrLen 获取哈希表某个字段长度。
// param: key string
// param: field string
func (s *Query) HStrLen(key string, field interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("HSTRLEN")
	s.setKey(key)
	s.append(field)

	return s
}

// Hkeys 获取哈希表中的所有域（field）。
// param: key string
func (s *Query) Hkeys(key string) *Query {

	s.ResultType = consts.RedisRetTypeStrings
	s.Op("HKEYS")
	s.setKey(key)
	return s
}

// HVals 返回所有的 value
// param: key string
func (s *Query) HVals(key string) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("HVALS")
	s.setKey(key)

	return s
}

// LPush 将一个或多个值插入到列表头部。 如果 key 不存在，一个空列表会被创建并执行 LPUSH 操作。 当 key 存在但不是列表类型时，返回一个错误。
// param: key string
// param: values interface{} 任意类型数据
func (s *Query) LPush(key string, values ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("LPUSH")
	s.setKey(key)
	for k, v := range values {
		values[k] = s.structToMap(v)
	}
	s.append(values...)

	return s
}

// RPush 将一个或多个值插入到列表的尾部(最右边)。如果列表不存在，一个空列表会被创建并执行 RPUSH 操作。 当列表存在但不是列表类型时，返回一个错误。
// param: key string
// param: v interface{} 任意类型数据
func (s *Query) RPush(key string, values ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("RPUSH")
	s.setKey(key)
	for k, v := range values {
		values[k] = s.structToMap(v)
	}
	s.append(values...)
	return s
}

// LPop 移除并返回列表的第一个元素。
// param: key string
func (s *Query) LPop(key string) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("LPOP")
	s.setKey(key)
	return s
}

// RPop 移除列表的最后一个元素，返回值为移除的元素。
// param: key string
func (s *Query) RPop(key string) *Query {
	s.ResultType = consts.RedisRetTypeString
	s.Op("RPOP")
	s.setKey(key)
	return s
}

// LLen 返回列表的长度。 如果列表 key 不存在，则 key 被解释为一个空列表，返回 0 。 如果 key 不是列表类型，返回一个错误。
// param: key string
func (s *Query) LLen(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("LLEN")
	s.setKey(key)
	return s
}

// SAdd 将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。
// param: key string
// param: v ...interface{} 任意类型的多条数据，但是务必确保各条数据的类型保持一致
func (s *Query) SAdd(key string, values ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SADD")
	s.setKey(key)

	for k, v := range values {
		values[k] = s.structToMap(v)
	}
	s.append(values...)
	return s
}

// SMembers 返回集合中的所有的成员。 不存在的集合 key 被视为空集合。
// param: key string
func (s *Query) SMembers(key string) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("SMEMBERS")
	s.setKey(key)
	return s
}

// SRem 移除集合中的一个或多个成员元素，不存在的成员元素会被忽略
// param: key string
// param: v ...interface{} 任意类型的多条数据
func (s *Query) SRem(key string, members ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SREM")
	s.setKey(key)

	s.append(members...)

	return s
}

// SCard 返回集合中元素的数量。
// param: key string
func (s *Query) SCard(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SCARD")
	s.setKey(key)

	return s
}

// SIsMember 判断成员元素是否是集合的成员。
// param: key string
// param: member interface{} 要检索的任意类型数据
func (s *Query) SIsMember(key string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeBool
	s.Op("SISMEMBER")
	s.setKey(key)

	s.append(member)

	return s
}

// SRandMember 返回集合中的count个随机元素。
// param: key string
// param: count int 随机返回元素个数。
// 如果 count 为正数，且小于集合基数，那么命令返回一个包含 count 个元素的数组，数组中的元素各不相同。
// 如果 count 大于等于集合基数，那么返回整个集合。
// 如果 count 为负数，那么命令返回一个数组，数组中的元素可能会重复出现多次，而数组的长度为 count 的绝对值。
func (s *Query) SRandMember(key string, count int) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("SRANDMEMBER")
	s.setKey(key)
	s.append(count)

	return s
}

// SPop 移除集合中的指定 key 的一个或多个随机成员，移除后会返回移除的成员。
// param: key string
// param: int count
func (s *Query) SPop(key string, count int) *Query {
	s.ResultType = consts.RedisRetTypeStrings
	s.Op("SPOP")
	s.setKey(key)
	s.append(count)

	return s
}

// SMove 将指定成员 member 元素从 source 集合移动到 destination 集合。
// param: source string
// param: destination string
// param: member interface{} 要移动的成员，任意类型
func (s *Query) SMove(source, destination string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("SMOVE")
	s.setKey(source)
	s.append(destination, member)

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
	s.setKey(key)

	if len(args) < 2 || len(args)%2 != 0 {
		s.Error = errs.Newf(errs.RetClientParamInvalid, "ZADD args should contain pair of memhber and score")
		return s
	}

	// 交换参数中的奇偶未知
	for i := 0; i < len(args); i += 2 {
		args[i], args[i+1] = args[i+1], s.structToMap(args[i])
	}

	s.append(args...)

	return s
}

// ZRem 移除有序集中的一个或多个成员，不存在的成员将被忽略。
// param: key string
// param: members ...interface{} 任意类型的多条数据
func (s *Query) ZRem(key string, members ...interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREM")
	s.setKey(key)
	s.append(members...)

	return s
}

// ZRemRangeByScore 移除有序集中，指定分数（score）区间内的所有成员。
// param: key string
// param: interface{} min max 分数区间，类型为整数或者浮点数
func (s *Query) ZRemRangeByScore(key string, min, max interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREMRANGEBYSCORE")
	s.setKey(key)
	s.append(min, max)

	return s
}

// ZRemRangeByRank 移除有序集中，指定排名(rank)区间内的所有成员。
// param: key string
// param: start stop int 排名区间
func (s *Query) ZRemRangeByRank(key string, start, stop int) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREMRANGEBYRANK")
	s.setKey(key)

	s.append(start, stop)

	return s
}

// ZCard 返回有序集成员个数
// param: key string
func (s *Query) ZCard(key string) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZCARD")
	s.setKey(key)

	return s
}

// ZScore 返回有序集中，成员的分数值。
// param: key string
// param: member interface{} 成员
func (s *Query) ZScore(key string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeFloat64
	s.Op("ZSCORE")
	s.setKey(key)
	s.append(member)

	return s
}

// ZRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。
// param: key string
// param: member interface{} 成员，任意类型
func (s *Query) ZRank(key string, member interface{}) *Query {

	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZRANK")
	s.setKey(key)
	s.append(member)

	return s
}

// ZRevRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从大到小)顺序排列。
// param: key string
// param: member interface{} 成员，任意类型
func (s *Query) ZRevRank(key string, member interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZREVRANK")
	s.setKey(key)
	s.append(member)

	return s
}

// ZCount 计算有序集合中指定分数区间的成员数量
// param: key string
// param: min interface{}
// param: max interface{}
func (s *Query) ZCount(key string, min, max interface{}) *Query {
	s.ResultType = consts.RedisRetTypeInt64
	s.Op("ZCOUNT")
	s.setKey(key)

	s.append(min, max)

	return s
}

// ZPopMin 移除并弹出有序集合中分值最小的 count 个元素
// redis v5.0.0+
// param: key string
// param: count ...int64 不设置count参数时，弹出一个元素
func (s *Query) ZPopMin(key string, count ...int64) *Query {
	s.ResultType = consts.RedisRetTypeMemberScore
	s.Op("ZPOPMIN")
	s.setKey(key)

	if len(count) != 0 {
		s.append(count[0])
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
	s.setKey(key)

	if len(count) != 0 {
		s.append(count[0])
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
	s.setKey(key)
	s.append(incr, member)

	return s
}

// ZRange 返回有序集中，指定区间内的成员。其中成员的位置按分数值递增(从小到大)来排序。
// param: key string
// param: int start, stop 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// param: withScore 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
func (s *Query) ZRange(key string, start, stop int, withScore ...bool) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	if len(withScore) > 0 && withScore[0] {
		s.SetParam("with_scores", true)
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	s.Op("ZRANGE")
	s.setKey(key)
	s.append(start, stop)

	return s
}

// ZRangeByScore 根据分数返回有序集中指定区间的成员，顺序从小到大
// param: key string
// param: int min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
// param: withScores 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// param: limit offset count 游标
func (s *Query) ZRangeByScore(key string, min, max interface{}, withScores bool, limit ...int64) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	if withScores {
		s.SetParam("with_scores", true)
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	s.Op("ZRANGEBYSCORE")
	s.setKey(key)

	if len(limit) == 2 {
		offset := limit[0]
		count := limit[1]
		s.append(min, max, "limit", offset, count)
	} else {
		s.append(min, max)
	}

	return s
}

// ZRevRange 返回有序集中指定区间的成员，其中成员的位置按分数值递减(从大到小)来排列。
// param: key string
// param: start, stop 排名区间，以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// param: withScore 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
func (s *Query) ZRevRange(key string, start, stop int, withScore ...bool) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	if len(withScore) > 0 && withScore[0] {
		s.SetParam("with_scores", true)
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	s.Op("ZREVRANGE")
	s.setKey(key)
	s.append(start, stop)

	return s
}

// ZRevRangeByScore 返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。
// param: key string
// param: max, min  interface{} 分数区间，类型为整数或双精度浮点数，但是 -inf +inf 表示负正无穷大
// param: withScore 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// param: limit offset count 游标
func (s *Query) ZRevRangeByScore(key string, max, min interface{}, withScore bool, limit ...int64) *Query {
	s.ResultType = consts.RedisRetTypeStrings

	if withScore {
		s.SetParam("with_scores", true)
		s.ResultType = consts.RedisRetTypeMemberScore
	}

	s.Op("ZREVRANGEBYSCORE")
	s.setKey(key)

	if len(limit) == 2 {
		offset := limit[0]
		count := limit[1]
		s.append(max, min, "limit", offset, count)
	} else {
		s.append(max, min)
	}

	return s
}

// setKey 给 key 赋值
func (s *Query) setKey(key string) *Query {
	s.Unit.Key = key
	return s
}

// append arg 参数拼接
func (s *Query) append(items ...interface{}) *Query {
	coder := s.GetCoder()

	if len(s.Unit.DataType) == 0 {
		s.Unit.DataType = make(map[string]consts.DataType)
	}

	for k, item := range items {
		mv, err := coder.Encode(codec.EncodeTypeRedisArg, item)
		if err != nil {
			s.Error = err
			return s
		}

		s.Unit.Args = append(s.Unit.Args, mv)

		typ := consts.GetDataType(mv)
		if typ != consts.DataTypeOther {
			s.Unit.DataType[strconv.Itoa(k)] = typ
		}
	}
	return s
}
