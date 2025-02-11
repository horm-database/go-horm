# horm 介绍
本文档是数据统一接入服务 UDAS（unified data access services）的客户端 golang sdk，完整实现了数据统一接入协议，支持 elastic search、redis、mysql/postgresql/clickhouse 等数据库相关操作。

数据统一接入协议：https://github.com/horm-database/server

```go
const ( // 后端服务支持的数据库类型
  DBTypeNil        = 0  // 空操作，仅走插件
  DBTypeElastic    = 1  // elastic search
  DBTypeMongo      = 2  // mongo 暂未支持
  DBTypeRedis      = 3  // redis
  DBTypeMySQL      = 10 // mysql
  DBTypePostgreSQL = 11 // postgresql
  DBTypeClickHouse = 12 // clickhouse
  DBTypeOracle     = 13 // oracle 暂未支持
  DBTypeDB2        = 14 // DB2 暂未支持
  DBTypeSQLite     = 15 // sqlite 暂未支持
  DBTypeRPC        = 40 // rpc 协议，暂未支持，spring cloud 协议可以选 grpc、thrift、tars、dubbo 协议
  DBTypeHTTP       = 50 // http 请求
)
```

使用 horm 访问数据统一接入服务的优势：
* 复杂业务拼写sql/es/redis 语句访问DB，可读性差，可维护性差，开发效率低下，不同的数据库需要拼写语句，在 UDAS，通过便捷的 horm ，可以极大提升开发效率，已经支持 mysql、postgresql、clickhouse、es、redis 等协议。
* 统一接入层协议可以极大提升跨部门、项目之间协作的效率，降低沟通成本。
* 所有业务模块单独支持数据库访问，开发成本高，权限分散，不易管理。在 UDAS ，可以在服务端做数据配置化管理，包括接入方授权、分表、日志级别等配置。
* 当有兄弟部门有数据需求，需要单独为其开发接口，效率低下，在 UDAS，可以直接给兄弟部门授权表级别数据权限，兄弟部门可以通过 horm 接入。也可以避免数据这里存一份、那里存一份，降低存储成本，降本增效。
* 组件解决并发性、高可用问题。比如缓存、多个执行单元并发、异步化，降级方案，针对指定接入方，可以降级为管理平台配置好的返回数据，不执行SQL。
* 高效的异常定位与解决方案，超时重试、失败手动重试功能，数据大盘可以用于 sql 性能分析，优化，可以对错误进行分析，数据暴增、快速定位暴增接入应用。
* horm 支持 GO、NODE、JAVA、C++、PHP。

#  示例表与结构体
## 示例表
```sql
CREATE TABLE `student` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `identify` bigint NOT NULL COMMENT '学生编号',
    `gender` tinyint NOT NULL DEFAULT '1' COMMENT '1-male 2-female',
    `age` int unsigned NOT NULL DEFAULT '0' COMMENT '年龄',
    `name` varchar(64) NOT NULL COMMENT '名称',
    `score` double DEFAULT NULL COMMENT '分数',
    `image` blob COMMENT 'image',
    `article` text COMMENT 'publish article',
    `exam_time` time DEFAULT NULL COMMENT '考试时间',
    `birthday` date DEFAULT NULL COMMENT '出生日期',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `identity` (`identify`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学生表'

CREATE TABLE `student_course` (
    `id` int NOT NULL AUTO_INCREMENT,
    `identify` bigint NOT NULL COMMENT '学生编号',
    `course` varchar(64) NOT NULL COMMENT '课程',
    `hours` int DEFAULT '0' COMMENT '课时数',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学生课程表';

CREATE TABLE `course_info` (
    `course` varchar(64) NOT NULL COMMENT '课程',
    `teacher` varchar(64) NOT NULL COMMENT '课程老师',
    `time` time NOT NULL COMMENT '上课时间',
    PRIMARY KEY (`course`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程信息';

CREATE TABLE `teacher_info` (
    `teacher` varchar(32) NOT NULL COMMENT '老师',
    `age` int NOT NULL DEFAULT '0' COMMENT '年龄',
    PRIMARY KEY (`teacher`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='老师信息';

CREATE TABLE `score_rank_reward` (
    `rank` int NOT NULL COMMENT '排名',
    `reward` varchar(128) NOT NULL DEFAULT '' COMMENT '奖励',
    PRIMARY KEY (`rank`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分数排名奖励'
```

## Golang 结构体：
```go
import (
  "time"
  
  "github.com/horm-database/common/types"
)

type Student struct {
	Id        uint64     `orm:"id,uint64,onuniqueid,omitempty" json:"id"`
	Identify  int64      `orm:"identify,int64,omitempty" json:"identify"`                 //学生编号
	Gender    int8       `orm:"gender,int8,omitinsertempty" json:"gender"`                //1-male 2-female
	Age       uint       `orm:"age,uint,omitreplaceempty" json:"age"`                     //年龄
	Name      string     `orm:"name,string,omitupdateempty" json:"name"`                  //名称
	Score     float64    `orm:"score,double,omitempty" json:"score"`                      //分数
	Image     []byte     `orm:"image,bytes,omitempty" json:"image"`                       //image
	Article   string     `orm:"article,string,omitempty" json:"article"`                  //publish article
	ExamTime  string     `orm:"exam_time,string,omitempty" json:"exam_time"`              //考试时间
	Birthday  types.Time `orm:"birthday,time,time_fmt='2006-01-02'" json:"birthday"`      //出生日期
	CreatedAt time.Time  `orm:"created_at,time,oninserttime,omitempty" json:"created_at"` //创建时间
    UpdatedAt time.Time  `orm:"updated_at,time,onupdatetime" json:"updated_at"`           //修改时间
}
```

```go
type StudentCourse struct {
	Id       int    `orm:"id,int,omitempty" json:"id"`
	Identify int64  `orm:"identify,int64" json:"identify"` //学生编号
	Course   string `orm:"course,string" json:"course"`    //课程
	Hours    int    `orm:"hours,int" json:"hours"`         //课时数
}

type CourseInfo struct {
	Course  string `orm:"course,string" json:"course"`   //课程
	Teacher string `orm:"teacher,string" json:"teacher"` //课程老师
	Time    string `orm:"time,string" json:"time"`       //上课时间
}

type TeacherInfo struct {
	Teacher string `orm:"teacher,string" json:"teacher"` //课程老师
	Age     int    `orm:"age,int" json:"age"`            //年龄
}

type ScoreRankReward struct {
	Rank   int    `orm:"rank,int" json:"rank"`        //排名
	Reward string `orm:"reward,string" json:"reward"` //奖励
}

```

## 结构体标签
支持通过 golang 结构体标签来描述数据库表字段，以及如何将结构体编码为请求，如下 Student 结构体，标签以 `orm` 开头，第一个参数为表字段名，
第二个参数为 orm 类型，其他则是属性，horm 会将结构体的 `orm` 标签解析后，将字段、类型、属性信息缓存到内存，用于编解码结构体：
```go
type Student struct {
	Id        uint64     `orm:"id,uint64,onuniqueid,omitempty" json:"id"`
	Identify  int64      `orm:"identify,int64,omitempty" json:"identify"`                 //学生编号
	Gender    int8       `orm:"gender,int8,omitinsertempty" json:"gender"`                //1-male 2-female
	Age       uint       `orm:"age,uint,omitreplaceempty" json:"age"`                     //年龄
	Name      string     `orm:"name,string,omitupdateempty" json:"name"`                  //名称
	Score     float64    `orm:"score,double,omitempty" json:"score"`                      //分数
	Image     []byte     `orm:"image,bytes,omitempty" json:"image"`                       //image
	Article   string     `orm:"article,string,omitempty" json:"article"`                  //publish article
	ExamTime  string     `orm:"exam_time,string,omitempty" json:"exam_time"`              //考试时间
	Birthday  types.Time `orm:"birthday,time,time_fmt='2006-01-02'" json:"birthday"`      //出生日期
	CreatedAt time.Time  `orm:"created_at,time,oninserttime,omitempty" json:"created_at"` //创建时间
	UpdatedAt time.Time  `orm:"updated_at,time,onupdatetime" json:"updated_at"`           //修改时间
}
```

`orm` 字段类型包含如下类型，更多详情参考章节 [基础数据类型](#基础数据类型)：
```go
var OrmType = map[string]Type{
	"time":   TypeTime,
	"bytes":  TypeBytes,
	"float":  TypeFloat,
	"double": TypeDouble,
	"int":    TypeInt,
	"uint":   TypeUint,
	"int8":   TypeInt8,
	"int16":  TypeInt16,
	"int32":  TypeInt32,
	"int64":  TypeInt64,
	"uint8":  TypeUint8,
	"uint16": TypeUint16,
	"uint32": TypeUint32,
	"uint64": TypeUint64,
	"string": TypeString,
	"bool":   TypeBool,
	"json":   TypeJSON,
}
```

`orm` 标签目前支持的属性如下：
- `omitinsertempty`: INSERT 时忽略零值，在插入数据的时候，如果字段为零值，这该字段被忽略插入，让其取数据库表默认值。
- `omitreplaceempty`: REPLACE 时忽略零值，在替换数据的时候，如果字段为零值，这该字段被忽略替换。
- `omitupdateempty`: UPDATE 时忽略零值，在更新数据的时候，如果字段为零值，这该字段被忽略更新，保持原值。
- `omitempty`: 数据新增、替换、修改时都忽略零值，= omitinsertempty + omitreplaceempty + omitupdateempty 。
- `oninserttime`: 在 INSERT 数据时，如果该字段为零值，初始化为当前时间，具体格式根据 Type 决定，如果是数字类型包括 int、int32、int64 等，则是时间戳，否则就是 time.Time 类型，如果设置了该属性，则在插入数据时 omit empty 属性失效。
- `onupdatetime`: 在 INSERT/REPLACE/UPDATE 数据时，如果该字段为零值，初始化为当前时间，具体格式根据 Type 决定，如果是数字类型包括 int、int32、int64 等，则是时间戳，否则就是 time.Time 类型，则在插入/修改数据时 omit empty 属性失效。
- `time_fmt`: 当字段底层类型为 time.Time 时，格式化时间，仅针对请求格式化，返回数据的解析在 codec 内。例如上面的 birthday 字段，time_fmt 会使得请求的时候，会将 birthday 转化为 `2006-01-02` 格式，在服务端返回字符串是 "2006-01-02" 格式时，只有类型 types.Time 才能正确接收结果。
- `onuniqueid`: 新增数据时候，如果字段为零值，而且类型为 uint64，则自动生成唯一 ID，如果设置了该属性，则在插入数据时 omit empty 属性失效，记得务必在 orm.yaml 配置里面为每台机器设置不同的 machine_id，否则生成的ID可能会有冲突，当然，你也可以采用数据库的自增id作为主键，这时候，最好加上 omitempty。
- `es_id`: 当出现该属性，表示本字段的值作为 es 主键，那么在插入数据的时候，我们会加入 `_id` 字段并赋值为该字段值。数据统一存储系统会将 `_id` 提取并作为该数据的主键。


# horm 客户端
为了访问数据统一接入服务，我们需要创建 Client 来与服务端建立连接，horm 提供了2种方式来指定 Query 语句使用的客户端。
一、为Query语句指定特定的客户端。二、配置全局客户端，在未指定特定客户端的情况下，所有 Query 都采用该全局客户端。

## 指定客户端
我们首先通过 horm.NewClient 创建一个客户端，该函数的第一个参数是允许传入一个 caller name， 他将读取配置文件 orm.yaml 里面的 
server.caller.name 对应的数据统一接入服务 workspace_id、 encryption、token、target、appid、secret等信息， 然后用 WithClient 为
Query 指定该 Client。 我们可能会访问多个数据统一接入服务的不同数据（每个服务都有唯一的 workspace id）。

```go
import (
  ...
  "github.com/horm-database/go-horm/horm"
)

func queryByClient(ctx context.Context) {
  cli := horm.NewClient("ws_test.app1.server1.service1")

  var result = make([]*Student, 0)
  _, err := horm.NewQuery("student").FindAll().WithClient(cli).Exec(ctx, &result)

  ...
}

```

orm.yaml 配置信息如下：

```yaml
machine: server.access.gz003      # 本地机器名（容器名）
machine_id: 3                     # 本地机器编号（容器编号），当我们将struct标签属性 onuniqueid 打开时，插入数据需要用到 snowflake.GenerateID() 自动生成ID，这时必须为每台机器设置不同的 machine_id，否则生成的ID可能会有冲突
local_ip: 127.0.0.1               # 本地IP，容器内为容器ip，物理机或虚拟机为本机ip
location:                         # 本地机器所属区域，主要用于 polaris 就近路由，如无服务发现可不填
  region: 腾讯云-华南
  zone: 广州
  compus: 园区1

server:                           # 数据统一接入服务端配置
  - workspace_id: 31              # workspace id
    encryption: 1                 # 帧签名方式 0-无（默认）1-签名 2-加密（针对 Android、IOS 等外网非安全客户端）
    token: QUIs32ODQUIs32OD       # workspace token
    target: ip://127.0.0.1:8180   # 服务端地址
    timeout: 3000000              # 接口调用超时时间（毫秒）
    caller:                       # 调用方信息
      - name: ws_test.app1.server1.service1 # 调用名
        appid: 10002              # 调用方 appid
        secret: S959223456        # 调用方秘钥
        timeout: 10000000         # 超时时间
      - name: ws_test.app2.server2.service2 # 调用名
        appid: 10003              # 调用方 appid
        secret: S499721834        # 调用方秘钥
        timeout: 2000000          # 超时时间

log:
  - writer: console               # 控制台标准输出 默认
    level: debug                  # 标准输出日志的级别
  - writer: file
    level: debug
    escape: true
    file_config:
      filename: ./server.log              # 本地文件滚动日志存放的路径
      max_size: 100                      # 本地文件滚动日志的大小 单位 MB
      max_backups: 30                    # 最大日志文件数
      max_day: 3                         # 最大日志保留天数
      compress: false                    # 日志文件是否压缩
```

另外，horm 提供 WithAppID、WithSecret 等一系列函数来为 Client 指定参数。

```go
import (
  ...
  "github.com/horm-database/common/codec"
  "github.com/horm-database/go-horm/horm"
)

func queryByClientWithOption(ctx context.Context) {
  cli := horm.NewClient("",
  horm.WithWorkspaceID(31),
  horm.WithEncryption(codec.FrameTypeSignature),
  horm.WithToken("QUIs32ODQUIs32OD"),
  horm.WithTarget("ip://127.0.0.1:8180"),
  horm.WithAppID(10099),
  horm.WithSecret("S499721834"),
  horm.WithTimeout(500))
  
  var result = make([]*Student, 0)
  _, err := horm.NewQuery("student").FindAll().WithClient(cli).Exec(ctx, &result)
  
  ...
}
```

## 配置全局Client
配置全局变量之后，如果 Query 没有用 WithClient 指定客户端的话，就使用全局客户端
```go
import (
	...
	"github.com/horm-database/go-horm/horm"
)

//init 配置全局Client
func init() {
  horm.SetGlobalClient("ws_test.app1.server1.service1")
}

func queryByGlobalClient(ctx context.Context) {
  var result = Student{}
  where := horm.Where{"identify": 2024061211}
  isNil, err := horm.NewQuery("student").Find(where).Exec(ctx, &result)

  ...
}

```

# 查询单元（执行单元）
## 数据名称
我们在客户端通过 horm.NewQuery 来创建一个查询，每个查询语句需要指定一个名称，如下的 `horm.NewQuery("student") 中的 student`，
horm 会生成一个执行单元（查询单元），并发送到数据统一接入服务， 在数据统一接入服务通过 `数据名称` 找到对应的mysql表/es索引/redis配置信息、
及其数据库信息，然后根据协议将执行单元转化为对应数据库 sql语句、elastic 请求或 redis 请求，并将执行结果返回到客户端。

```go
import (
	...
	"github.com/horm-database/go-horm/horm"
)

func Test(ctx context.Context) {
	var result = make([]*Students, 0)
	err := horm.NewQuery("student").FindAll().Exec(ctx, &result)
	...
}
```

如果存在相同的数据名称的时候，我们可以通过增加库名来区分如下，否则会报错，不允许存在相同的库名+数据名。
```go
import (
	...
	"github.com/horm-database/go-horm/horm"
)

func Test(ctx context.Context) {
	var result = make([]*Students, 0)
	err := horm.NewQuery("test::student").FindAll().Exec(ctx, &result)
	...
}
```
## 查询单元结构体
一个完整的执行单元包含如下信息：
```go
// github.com/horm-database/common/proto
package proto

import (
	"github.com/horm-database/common/proto/sql"
	"github.com/horm-database/common/types"
)

// Unit 查询单元（执行单元）
type Unit struct {
	// query base info
	Name  string   `json:"name,omitempty"`  // name
	Op    string   `json:"op,omitempty"`    // operation
	Shard []string `json:"shard,omitempty"` // 分片、分表、分库

	// 结构化查询共有
	Column []string               `json:"column,omitempty"` // columns
	Where  map[string]interface{} `json:"where,omitempty"`  // query condition
	Order  []string               `json:"order,omitempty"`  // order by
	Page   int                    `json:"page,omitempty"`   // request pages. when page > 0, the request is returned in pagination.
	Size   int                    `json:"size,omitempty"`   // size per page
	From   uint64                 `json:"from,omitempty"`   // offset

	// data maintain
	Val      interface{}              `json:"val,omitempty"`       // 单条记录 val (not map/[]map)
	Data     map[string]interface{}   `json:"data,omitempty"`      // maintain one map data
	Datas    []map[string]interface{} `json:"datas,omitempty"`     // maintain multiple map data
	Args     []interface{}            `json:"args,omitempty"`      // multiple args, 还可用于 query 语句的参数，或者 redis 协议，如 MGET、HMGET、HDEL 等
	DataType map[string]types.Type    `json:"data_type,omitempty"` // 数据类型（主要用于 clickhouse，对于数据类型有强依赖），请求 json 不区分 int8、int16、int32、int64 等，只有 Number 类型，bytes 也会被当成 string 处理。

	// group by
	Group  []string               `json:"group,omitempty"`  // group by
	Having map[string]interface{} `json:"having,omitempty"` // group by condition

	// for databases such as mysql ...
	Join []*sql.Join `json:"join,omitempty"`

	// for databases such as elastic ...
	Type   string  `json:"type,omitempty"`   // type, such as elastic`s type, it can be customized before v7, and unified as _doc after v7
	Scroll *Scroll `json:"scroll,omitempty"` // scroll info

	// for databases such as redis ...
	Prefix string   `json:"prefix,omitempty"` // prefix, It is strongly recommended to bring it to facilitate finer-grained summary statistics, otherwise the statistical granularity can only be cmd ，such as GET、SET、HGET ...
	Key    string   `json:"key,omitempty"`    // key
	Keys   []string `json:"keys,omitempty"`   // keys

	// bytes 字节流
	Bytes []byte `json:"bytes,omitempty"`

	// params 与数据库特性相关的附加参数，例如 redis 的 withscores、EX、NX、等，以及 elastic 的 refresh、collapse、runtime_mappings、track_total_hits 等等。
	Params map[string]interface{} `json:"params,omitempty"`

	// 直接送 Query 语句，需要拥有库的表操作权限、或 root 权限。具体参数为 args
	Query string        `json:"query,omitempty"`

	// Extend 扩展信息，作用于插件
	Extend map[string]interface{} `json:"extend,omitempty"`

	Sub   []*Unit `json:"sub,omitempty"`   // 子查询
	Trans []*Unit `json:"trans,omitempty"` // 事务，该事务下的所有 Unit 必须同时成功或失败（注意：仅适合支持事务的数据库回滚，如果数据库不支持事务，则操作不会回滚）
}

// Scroll 滚动查询
type Scroll struct {
	ID   string `json:"id,omitempty"`   // 滚动 id
	Info string `json:"info,omitempty"` // 滚动查询信息，如时间
}

type Join struct {
	Type  string            `json:"type,omitempty"`
	Table string            `json:"table,omitempty"`
	Using []string          `json:"using,omitempty"`
	On    map[string]string `json:"on,omitempty"`
}
```

## 基础数据类型
执行单元中的 data、datas、args 等数据参数，可以包含如下一些基础数据类型下：
```go
// github.com/horm-database/common/types
package types

type Type int8

const (
	TypeTime   Type = 1 // 类型是 time.Time
	TypeBytes  Type = 2 // 类型是 []byte
	TypeFloat  Type = 3
	TypeDouble Type = 4
	TypeInt    Type = 5
	TypeUint   Type = 6
	TypeInt8   Type = 7
	TypeInt16  Type = 8
	TypeInt32  Type = 9
	TypeInt64  Type = 10
	TypeUint8  Type = 11
	TypeUint16 Type = 12
	TypeUint32 Type = 13
	TypeUint64 Type = 14
	TypeString Type = 15
	TypeBool   Type = 16
	TypeJSON   Type = 17
)

var OrmType = map[string]Type{
	"time":   TypeTime,
	"bytes":  TypeBytes,
	"float":  TypeFloat,
	"double": TypeDouble,
	"int":    TypeInt,
	"uint":   TypeUint,
	"int8":   TypeInt8,
	"int16":  TypeInt16,
	"int32":  TypeInt32,
	"int64":  TypeInt64,
	"uint8":  TypeUint8,
	"uint16": TypeUint16,
	"uint32": TypeUint32,
	"uint64": TypeUint64,
	"string": TypeString,
	"bool":   TypeBool,
	"json":   TypeJSON,
}
```

我们发送请求到数据统一调度服务的时候，绝大多数情况下可以不指定数据类型，服务端也可以正常解析并执行 query 语句，但是在某些特殊情况下，
比如 clickhouse 对类型有强限制，又或者字段是一个超大 uint64 整数，json 编码之后请求服务端，由于 json 的基础类型只包含 string、 
number(当成float64)、bool，数字在服务端会被解析为 float64，存在精度丢失问题，所以在 golang horm 中，当类型为 time、[]byte、
int、 int8`~`int64、uint、uint8`~`uint64 时，需要在执行单元 data_type 字段里将数据类型带上，当然 horm-sdk 会自动帮我们处理，如下案例：

```go
import (
  ...
  "github.com/horm-database/go-horm/horm"
)

func queryDataType(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1987-08-27")

	data := Student{
		Identify: 2024080313,
		Gender:   1,
		Age:      23,
		Name:     "caohao",
		Score:    91.5,
		Image:    []byte("IMAGE.PCG"),
		Article:  "groundbreaking work in cryptography and complexity theory",
		ExamTime: "15:30:00",
		Birthday: types.Time(birthday),
	}
	
	var addRet = proto.ModRet{}

	//下面操作有加别名
	_, err := horm.NewQuery("student(add)").Insert(&data).Exec(ctx, &addRet)

	...
}
```
上面代码生成的 json 请求为：
```json
{
  "name": "student(add)",
  "op": "insert",
  "data": {
    "identify": 2024080313,
    "name": "caohao",
    "score": 91.5,
    "created_at": "2025-01-05T20:14:50.702248+08:00",
    "exam_time": "15:30:00",
    "birthday": "1987-08-27",
    "updated_at": "2025-01-05T20:14:50.702249+08:00",
    "article": "groundbreaking work in cryptography and complexity theory",
    "id": 234047220842770433,
    "image": "SU1BR0UuUENH",
    "gender": 1,
    "age": 23
  },
  "data_type": {
    "id": 14,
    "image": 2,
    "created_at": 1,
    "identify": 10,
    "gender": 7,
    "age": 6,
    "updated_at": 1
  }
}
```

horm 基础类型，会在数据统一接入服务根据指定的数据源引擎映射、解析成对应的类型，例如在 mysql 和 clickhouse 类型映射为：
```go
//github.com/orm/database/sql/type.go

var MySQLTypeMap = map[string]types.Type{
  "INT":                types.TypeInt,
  "TINYINT":            types.TypeInt8,
  "SMALLINT":           types.TypeInt16,
  "MEDIUMINT":          types.TypeInt32,
  "BIGINT":             types.TypeInt64,
  "UNSIGNED INT":       types.TypeUint,
  "UNSIGNED TINYINT":   types.TypeUint8,
  "UNSIGNED SMALLINT":  types.TypeUint16,
  "UNSIGNED MEDIUMINT": types.TypeUint32,
  "UNSIGNED BIGINT":    types.TypeUint64,
  "BIT":                types.TypeBytes,
  "FLOAT":              types.TypeFloat,
  "DOUBLE":             types.TypeDouble,
  "DECIMAL":            types.TypeDouble,
  "VARCHAR":            types.TypeString,
  "CHAR":               types.TypeString,
  "TEXT":               types.TypeString,
  "BLOB":               types.TypeBytes,
  "BINARY":             types.TypeBytes,
  "VARBINARY":          types.TypeBytes,
  "TIME":               types.TypeString,
  "DATE":               types.TypeTime,
  "DATETIME":           types.TypeTime,
  "TIMESTAMP":          types.TypeTime,
  "JSON":               types.TypeJSON,
}

var ClickHouseTypeMap = map[string]types.Type{
  "Int":         types.TypeInt,
  "Int8":        types.TypeInt8,
  "Int16":       types.TypeInt16,
  "Int32":       types.TypeInt32,
  "Int64":       types.TypeInt64,
  "UInt":        types.TypeUint,
  "UInt8":       types.TypeUint8,
  "UInt16":      types.TypeUint16,
  "UInt32":      types.TypeUint32,
  "UInt64":      types.TypeUint64,
  "Float":       types.TypeFloat,
  "Float32":     types.TypeFloat,
  "Float64":     types.TypeDouble,
  "Decimal":     types.TypeDouble,
  "String":      types.TypeString,
  "FixedString": types.TypeString,
  "UUID":        types.TypeString,
  "DateTime":    types.TypeTime,
  "DateTime64":  types.TypeTime,
  "Date":        types.TypeTime,
}
```

## 别名
如果我们用到 mysql 的别名，或者在并发查询、复合查询模式下、同一层级的多个查询单元如果访问同一张表，为了结果的正常，我们必须在括号里加上别名，
如下代码的`horm.NewQuery("student(add)")` 和 `Next("student(find_all)")` ，我们都是访问 redis_student。
```go
import (
    ...
    "github.com/horm-database/common/proto"
    "github.com/horm-database/common/types"
    "github.com/horm-database/go-horm/horm"
)

func queryAlias(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1995-03-24")

	data := Student{
		Identify: 2024092316,
		Gender:   1,
		Age:      17,
		Name:     "jerry",
		Score:    82.5,
		Image:    []byte("IMAGE.PCG"),
		Article:  "contributions to deep learning in artificial intelligence",
		ExamTime: "15:30:00",
		Birthday: types.Time(birthday),
	}

	var isNil bool
	var addErr, findErr error
	var addRet = proto.ModRet{}
	var student = Student{}

	//下面操作有加别名
	err := horm.NewQuery("student(add)").
		Insert(&data).WithReceiver(nil, &addErr, &addRet).
		Next("student(find)").
		Find(horm.Where{"@id": "add.id"}).WithReceiver(&isNil, &findErr, &student).
		PExec(ctx)

	...
}
```

以下是上面请求的返回结果，是一个 map[string]interface，其中 map 的 key 就是执行单元的名称或别名，如果都用 student，则无法区分是返回
是哪个执行单元的结果，而且会丢失一个执行单元的结果，这时候需要用别名来区别。

```json
{
  "add": {
    "id": "234049805125431297",
    "rows_affected": 1
  },
  "find": {
    "name": "jerry",
    "image": "SU1BR0UuUENH",
    "article": "contributions to deep learning in artificial intelligence",
    "updated_at": "2025-01-05T20:30:22+08:00",
    "birthday": "1995-03-24T00:00:00+08:00",
    "created_at": "2025-01-05T20:30:22+08:00",
    "id": 234049805125431297,
    "identify": 2024092316,
    "gender": 1,
    "age": 17,
    "score": 82.5,
    "exam_time": "15:30:00"
  }
}
```

另外一种情况就是作为 mysql 的别名存在
```go
func queryAlias2(ctx context.Context) {
	var result = []map[string]interface{}{}
	_, err := horm.NewQuery("student_course(sc)").Column("sc.*", "s.name").FindAll().
		LeftJoin("student(s)", horm.On{"identify": "identify"}).
		Exec(ctx, &result)

	...
}
```

上面的代码生成的 sql 语句如下：
```sql
SELECT  `sc`.* , `s`.`name`  FROM `student_course` AS `sc` 
	LEFT JOIN `student` AS `s` ON `sc`.`identify`=`s`.`identify`
```

## 分片、分表、分库
默认情况下，我们的表名就等于数据名，但是，如果有mysql分表、elastic分索引、redis分库的情况，我们需要用到 shard 功能来指定分表，
如下案例我们 student 表，根据 identify % 100 分了100张分表。
```go
func queryShard(ctx context.Context) {
	var student = Student{}

	_, err := horm.NewQuery("student").
		Shard("student_33").Find(horm.Where{"identify": 2024070733}).Exec(ctx, &student)

	...
}
```

在统一接入服务，我们会校验 shard 表是否符合该数据的表校验规则，表校验规则支持单一表名、逗号分隔的多个表名、正则表达式 regex/student_*?/、
还有就是比较常用 `...` 校验， 例如咱们例子中的student_0...99 表示 从 student_0 一直到 student_99。

# 查询模式
## 单查询单元
整个查询仅包含一个执行单元
### 单结果接收
执行单条语句，`isNil`, `error` 直接通过 Exec 函数返回，当查询结果为空时，isNil=true，可以将 result 指针传入 Exec 第二个参数，接收返回结果。
```go
import (
	...
	"github.com/horm-database/go-horm/horm"
)

func queryModeSingle(ctx context.Context) {
  var result = []*Student{}
  where := horm.Where{"name ~": "%caohao%"}
  isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

  ...
}
```

### 多结果接收
有时候，可能会返回多个结果，需要两个参数去接受结果，例如 redis 的 ZRangeByScore：
```go
import (
    ...
    "github.com/horm-database/go-horm/horm"
    "github.com/horm-database/common/types"
)

func queryMultiReturn(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1987-08-27")
	data := Student{
		Identify: 2024092316,
		Gender:   1,
		Age:      17,
		Name:     "jerry",
		Score:    82.5,
		Image:    []byte("IMAGE.PCG"),
		Article:  "contributions to deep learning in artificial intelligence",
		ExamTime: "15:30:00",
		Birthday: types.Time(birthday),
	}

	_, err := horm.NewQuery("redis_student").
		ZAdd("student_age_rank", data, float64(data.Age)).Exec(ctx)

	results := make([]*Student, 0)
	ages := make([]float64, 0)
	_, err = horm.NewQuery("redis_student").
		ZRangeByScore("student_age_rank", 10, 50, true).Exec(ctx, &results, &ages)

	...
}
```

返回结果如下：
![image](https://github.com/horm-database/image/blob/master/4-1.png)

### 分页返回
当我们请求参数 page > 1 时，返回结果会以分页形式返回，接收数据有两种方式：
```go
import (
    ...
    "github.com/horm-database/common/proto"
    "github.com/horm-database/go-horm/horm"
)

func queryPageReturn(ctx context.Context) {
	pageInfo := proto.Detail{}
	students := make([]*Student, 0)

	isNil, err := horm.NewQuery("student").
		FindAll().Page(1, 10).Exec(ctx, &pageInfo, &students)

	...
}
```

实际上统一接入服务返回的分页数据结构如下：

```go
// PageResult 当 page > 1 时会返回分页结果
type PageResult struct {
	Detail *Detail       `orm:"detail,omitempty" json:"detail,omitempty"` // 查询细节信息
	Data   []interface{} `orm:"data,omitempty" json:"data,omitempty"`     // 分页结果
}

// Detail 其他查询细节信息，例如 分页信息、滚动翻页信息、其他信息等。
type Detail struct {
	Total     uint64                 `orm:"total" json:"total"`                               // 总数
	TotalPage uint32                 `orm:"total_page,omitempty" json:"total_page,omitempty"` // 总页数
	Page      int                    `orm:"page,omitempty" json:"page,omitempty"`             // 当前分页
	Size      int                    `orm:"size,omitempty" json:"size,omitempty"`             // 每页大小
	Scroll    *Scroll                `orm:"scroll,omitempty" json:"scroll,omitempty"`         // 滚动翻页信息
	Extras    map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"`         // 更多详细信息
}
```

所以我们也可以直接用如下方式去接收返回结果：

```go
func queryPageReturn2(ctx context.Context) {
	result := proto.PageResult{}
	isNil, err := horm.NewQuery("student").FindAll().Page(1, 10).Exec(ctx, &result)

	...
}

```

## 并行查询
### 并发同时执行
为了高效并发，我们可以用 `PExec` 函数将多个语句一同上传到数据统一接入服务，由数据统一接入服务并发执行，并返回结果，在 Query 语句里面，
可以通过 `Next` 新建一个并发语句，然后通过 `WithReceiver` 传入对应指针来接收每个执行语句返回的 isNil、error 和结果。

`注意：如果并行执行访问同一个数据时，为了区别，可以像下面一样在括号里面加别名：redis_student(zadd) 和 redis_student(range)。`<br><br>

`另外我们注意看返回结果，ZRangeByScore 仅返回了2条数据，实际上应该有3条数据，也就是 ZAdd 的数据并未出现在 ZRangeByScore 结果中， 这是
因为在并发执行过程中，两个语句是同时执行，我们并不知道哪个语句先执行完，如果 ZRangeByScore 先于 ZAdd 执行完成，就会导致数据还未插入完成就
获取了排序结果，这显然与我们的预期不符，所以当遇到两条执行语句有先后要求时，我们最好拆成两条独立的语句先后执行，而不是放在一个并发执行中。`

```go
func queryModeParallel(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1987-08-27")
	data := Student{
		Identify: 2024092316,
		Gender:   1,
		Age:      17,
		Name:     "jerry",
		Score:    82.5,
		Image:    []byte("IMAGE.PCG"),
		Article:  "contributions to deep learning in artificial intelligence",
		ExamTime: "15:30:00",
		Birthday: types.Time(birthday),
	}

	var isNil bool
	var zaddErr, rangeErr error
	students := make([]*Student, 0)
	ages := make([]float64, 0)

	//下面操作有加别名
	err := horm.NewQuery("redis_student(zadd)").
		ZAdd("student_age_rank", &data, float64(data.Age)).WithReceiver(nil, &zaddErr).
		Next("redis_student(range)").
		ZRangeByScore("student_age_rank", 10, 50, true).WithReceiver(&isNil, &rangeErr, &students, &ages).
		PExec(ctx)

	...
}
```

返回结果如下：<br>

![image](https://github.com/horm-database/image/blob/master/4-2.png)

### 引用
引用是指的一个查询单元的请求参数来自另外一个查询的返回结果，当出现引用的时候，并行执行会退化为串行执行。引用有多种方式，
如下 horm.Where{"@identify": "student.identify"} 中 map 的 key 以`@`开头的时候，表示 identify 的值引用自 student 执行单元
的返回结果的 identify 字段。`.` 之前表示引用路径，之后表示引用的 field， 被引用的执行单元必须在引用的执行单元之前被执行，否则就会报错。

```go
import (
	...
	"github.com/horm-database/common/proto"
	"github.com/horm-database/go-horm/horm"
)

func queryReference(ctx context.Context) {
	var studentIsNil, courseIsNil bool
	var studentErr, courseErr error
	var page = proto.Detail{}
	var students = make([]*Student, 0)
	var studentCourse = make([]*StudentCourse, 0)

	err := horm.NewQuery("student").FindAll().Page(1, 10).
		WithReceiver(&studentIsNil, &studentErr, &page, &students).
		Next("student_course").FindAll(horm.Where{"@identify": "student.identify"}).
		WithReceiver(&courseIsNil, &courseErr, &studentCourse).
		PExec(ctx)

    ...
}
```

当引用参数是 key (string) 或者 args ([]interface{}) 而不是 horm.Where (map[string]interface{}) 的时候， 
需要 `@{}` 方式，例如 @{student.identify} 来表示该参数来自于引用 student.identify。 例如下面这个例子，
我们需要先查询 name="caohao" 的学生，然后根据学生的 identify 来获取他的排名：

```go
func queryReference2(ctx context.Context) {
	var isNil bool
	var studentErr, rankErr error
	var student = Student{}
	var rank int

	err := horm.NewQuery("student").Find(horm.Where{"name": "caohao"}).
		WithReceiver(&isNil, &studentErr, &student).
		Next("redis_student").ZRank("student_score_rank", "@{student.identify}").
		WithReceiver(nil, &rankErr, &rank).
		PExec(ctx)

	...
}
```

当被引用的值不是一个 map，而是一个具体数值的时候，我们不需要 `.` 来指定 field，而是直接采用被引用的执行单元即可。 例如下面我们获取了
一个学生的排名， 我们期望在一个并行执行单元中知道该排名的奖励：
```go
func queryReference3(ctx context.Context) {
	var isNil bool
	var rankErr, rankRewardErr error
	var rank int
	var rankReward = ScoreRankReward{}

	err := horm.NewQuery("redis_student(score_rank)").ZRank("student_score_rank", 2024092316).
		WithReceiver(nil, &rankErr, &rank).
		Next("score_rank_reward").Find(horm.Where{"@rank": "score_rank"}).
		WithReceiver(&isNil, &rankRewardErr, &rankReward).
		PExec(ctx)

	...
}
```

## 复合查询
### 返回结构
复合执行包含并行执行加上子查询，在复合查询的结果，如果返回的是一个数组，我们会为每个数组结果都执行一遍该查询的子查询，每个复合查询的结果
都包含 error、is_nil、detail 和 data 4个参数，当 error 不存在或者等于 nil 的时候，则结果正常无报错，分页等详情再 detail 中，
如果返回数据为空则 is_nil=true，当 is_nil 不存在，或者等于 false 时，返回数据存在于 data 中。子查询也在父查询的返回 data 中。

```go
package proto // "github.com/horm-database/common/proto"

// CompResult 混合查询返回结果
type CompResult struct {
	RetBase             // 返回基础信息
	Data    interface{} `json:"data"` // 返回数据
}

// RetBase 混合查询返回结果基础信息
type RetBase struct {
	Error  *Error  `json:"error,omitempty"`  // 错误返回
	IsNil  bool    `json:"is_nil,omitempty"` // 是否为空
	Detail *Detail `json:"detail,omitempty"` // 查询细节信息
}
```

下面是一个复杂的例子：
```go
import (
	...
    "github.com/horm-database/common/proto"
    "github.com/horm-database/go-horm/horm"
)

func queryModeCompound(ctx context.Context) {
  type RetInfo struct {
    Student struct {
      proto.RetBase // 返回基础信息
      Data []*struct {
        Student
        StudentCourse struct {
          proto.RetBase
          Data []*struct {
            StudentCourse
            CourseInfo struct {
              proto.RetBase
              Data *CourseInfo `json:"data,omitempty"`
            } `json:"course_info"` // 课程信息
          } `json:"data,omitempty"`
        } `json:"student_course"` // 学生选修的课程
        TeacherInfo struct {
          proto.RetBase
          Data []struct {
            TeacherInfo
            TestNil struct {
              proto.RetBase
              Data string `json:"data,omitempty"`
            } `json:"test_nil"` // 测试空返回
          } `json:"data,omitempty"`
        } `json:"teacher_info"` //教师信息
      } `json:"data,omitempty"`
    } `json:"student"`
    TestError struct {
      proto.RetBase
      Data *TeacherInfo `json:"data,omitempty"`
    } `json:"test_error"` // 测试 error 返回
  }
  
  ret := RetInfo{}
  
  //下面操作有加别名
  err := horm.NewQuery("student").Page(1, 10).FindAll().
      AddSub(horm.NewQuery("student_course").FindAll(horm.Where{"@identify": "/student.identify"}).
          AddSub(horm.NewQuery("course_info").Find(horm.Where{"@course": "../.course"})).
          AddNext(horm.NewQuery("teacher_info").FindAll(horm.Where{"@teacher": "student_course/course_info.teacher"}).
              AddSub(horm.NewQuery("redis_student(test_nil)").Get("not_exists")),
          ),
      ).
      Next("teacher_info(test_error)").Find(horm.Where{"not_exist_field": 55}).
      CompExec(ctx, &ret)
  
  ...
}
```

生成的请求json如下：<br>

```json
[
    {
        "name": "student",
        "op": "find_all",
        "page": 1,
        "size": 10,
        "sub": [
            {
                "name": "student_course",
                "op": "find_all",
                "where": {
                    "@identify": "/student.identify"
                },
                "size": 100,
                "sub": [
                    {
                        "name": "course_info",
                        "op": "find",
                        "where": {
                            "@course": "../.course"
                        }
                    }
                ]
            },
            {
                "name": "teacher_info",
                "op": "find_all",
                "where": {
                    "@teacher": "student_course/course_info.teacher"
                },
                "size": 100,
                "sub": [
                    {
                        "name": "redis_student(test_nil)",
                        "op": "get",
                        "key": "not_exists"
                    }
                ]
            }
        ]
    },
    {
        "name": "teacher_info(test_error)",
        "op": "find",
        "where": {
            "not_exist_field": 55
        }
    }
]
```

返回结果如下，整个返回结果会 json.Unmarshal 到接收结构体，即上面的 RetInfo 结构体:<br>

```json
{
  "student": {
    "detail": {
      "total": 2,
      "total_page": 1,
      "page": 1,
      "size": 10
    },
    "data": [
      {
        "id": 234047220842770433,
        "identify": 2024080313,
        "gender": 1,
        "age": 23,
        "name": "caohao",
        "score": 91.5,
        "image": "SU1BR0UuUENH",
        "article": "groundbreaking work in cryptography and complexity theory",
        "exam_time": "15:30:00",
        "birthday": "1987-08-27T00:00:00+09:00",
        "created_at": "2025-01-05T20:20:06+08:00",
        "updated_at": "2025-01-05T20:20:06+08:00",
        "student_course": {
          "data": [
            {
              "id": 1,
              "identify": 2024080313,
              "course": "Math",
              "hours": 54,
              "course_info": {
                "data": {
                  "course": "Math",
                  "teacher": "Simon",
                  "time": "11:00:00"
                }
              }
            },
            {
              "id": 2,
              "identify": 2024080313,
              "course": "Physics",
              "hours": 32,
              "course_info": {
                "data": {
                  "course": "Physics",
                  "teacher": "Richard",
                  "time": "14:00:00"
                }
              }
            }
          ]
        },
        "teacher_info": {
          "data": [
            {
              "teacher": "Richard",
              "age": 57,
              "test_nil": {
                "is_nil": true
              }
            },
            {
              "teacher": "Simon",
              "age": 61,
              "test_nil": {
                "is_nil": true
              }
            }
          ]
        }
      },
      {
        "id": 234049805125431297,
        "identify": 2024092316,
        "gender": 1,
        "age": 17,
        "name": "jerry",
        "score": 82.5,
        "image": "SU1BR0UuUENH",
        "article": "contributions to deep learning in artificial intelligence",
        "exam_time": "15:30:00",
        "birthday": "1995-03-24T00:00:00+08:00",
        "created_at": "2025-01-05T20:30:22+08:00",
        "updated_at": "2025-01-05T20:30:22+08:00",
        "student_course": {
          "data": [
            {
              "id": 3,
              "identify": 2024092316,
              "course": "English",
              "hours": 68,
              "course_info": {
                "data": {
                  "course": "English",
                  "teacher": "Dennis",
                  "time": "15:30:00"
                }
              }
            }
          ]
        },
        "teacher_info": {
          "data": [
            {
              "teacher": "Dennis",
              "age": 39,
              "test_nil": {
                "is_nil": true
              }
            }
          ]
        }
      }
    ]
  },
  "test_error": {
    "error": {
      "type": 2,
      "code": 1054,
      "msg": "mysql query error: [Unknown column 'not_exist_field' in 'where clause']"
    }
  }
}
```

### 引用路径
不同于并行查询的所有查询单元都在同一个层级，在复合查询中，有了子查询，在不同层级的情况下，引用会变得复杂，我们可以采用相对路径和绝对路径，
来指向我们需要被引用的查询单元。 如果 `/` 开头，则表是该路径属于绝对路径，例如上面实例中的 `/student.identify`，否则，就是相对路径，
相对路径在计算的时候，会把当前层级所在的父查询的绝对路径加在相对路径前，例如上面案例的 `student_course/course_info.teacher` ，
会变成 `/student/student_course/course_info.teacher`如果以 `../` 开头的相对路径，则会把`../` 转化为父查询的绝对路径，
例如上面案例的 `../.course`，会变成 `/student/student_course.course`，在相对路径转化为绝对路径之后，再根据规则获取指定路径的引用结果。

## 返回结果
### 空返回 和 error
当数据源为 mysql、clickhouse、es 等数据库时，如果 Find 或者 FindAll 查询的数据为空时，返回参数 isNil=true，否则，返回参数为 false，
而当数据源为 redis 时，只有 redis 返回 redigo: nil returned 错误时，才会使得 isNil = true，其他时候都是 isNil = false，
即便如下 ZRangeByScore 去查询一个不存在的有序集时，isNil 也是 false。

```go
func queryReturnNil(ctx context.Context) {
	var result Student
	where := horm.Where{"name": "noexist"}
	isNil, err := horm.NewQuery("student").Find(where).Exec(ctx, &result) // isNil = true

	var results []*Student
	where = horm.Where{"name": "noexists"}
	isNil, err = horm.NewQuery("student").FindAll(where).Exec(ctx, &results) // isNil = true

	// redis 中 GET 缓存
	var stu Student
	isNil, err = horm.NewQuery("redis_student").Get("noexists").Exec(ctx, &stu) // isNil = true

	// redis 中 ZRANGEBYSCORE
	rets := make([]*Student, 0)
	scores := make([]float64, 0)
	isNil, err = horm.NewQuery("redis_student"). // isNil = false ， rets 和 scores 为空数组
							ZRangeByScore("noexists", 70, 100, true).Exec(ctx, &rets, &scores)

	fmt.Println(isNil, err)
}
```

上面展示的是单执行单元的返回结果，在单执行单元中，is_nil、error 参数在 ResponseHeader 中返回客户端：
```protobuf
/* ResponseHeader 响应头 */
message ResponseHeader {
  ...
  Error err = 5;                     // 返回错误
  bool is_nil = 6;                   // 返回是否为空（针对单执行单元）
}
```

在并行查询中，一般系统返回，例如请求参数错误、解析失败、网络错误、权限错误等都会在 ResponseHeader 的 err 返回。
每个并行查询单元的 is_nil、error 结果则会在 ResponseHeader 中的 rsp_nils、rsp_errs 中返回给客户端，
这是一个 map，key是请求名(别名)，在 golang sdk 里面通过每个 Query 的 WithReceiver 来接收。
```protobuf
/* ResponseHeader 响应头 */
message ResponseHeader {
  ...
  Error err = 5;                     // 返回错误
  map<string, Error> rsp_errs = 7;   // 错误返回（针对多执行单元并发）
  map<string, bool> rsp_nils = 8;    // 是否为空返回（针对多执行单元并发）
}
```

```go
func queryReturnError(ctx context.Context) {
	data := map[string]interface{}{
		"no_field": nil,
	}

	var addErr, findErr error
	var addRet = proto.ModRet{}
	var student = Student{}

	err := horm.NewQuery("student(add)").
		Insert(data).WithReceiver(nil, &addErr, &addRet).
		Next("student(find)").
		Find(horm.Where{"no_field": "caohao"}).WithReceiver(nil, &findErr, &student).
		PExec(ctx)

	...
}
```

在复合查询中，请求参数错误、解析失败、网络错误、权限错误等依然在 ResponseHeader 的 err 中返回，
每个查询单元的 is_nil、error 则包含在结果里面。
```go
package proto // "github.com/horm-database/common/proto"

// CompResult 混合查询返回结果
type CompResult struct {
	RetBase             // 返回基础信息
	Data    interface{} `json:"data"` // 返回数据
}

// RetBase 混合查询返回结果基础信息
type RetBase struct {
	Error  *Error  `json:"error,omitempty"`  // 错误返回
	IsNil  bool    `json:"is_nil,omitempty"` // 是否为空
	Detail *Detail `json:"detail,omitempty"` // 查询细节信息
}
```

数据统一接入服务的错误结构如下，错误包含：错误类型，错误码，错误信息，异常查询语句组成（sql不仅指代sql语句，elastic语句、redis 命令也包含在内）
```protobuf
message Error {
  int32  type = 1; //错误类型
  int32  code = 2; //错误码
  string msg = 3;  //错误信息
  string sql = 4;  //异常sql语句
}
```
错误类型包含3大类，比如请求参数错误、解析失败、网络错误、权限错误等都属于系统错误，找不到插件、插件未注册、插件执行错误等都属于插件错误。
数据库执行报错都属于数据库错误。
```go
// EType 错误类型
type EType int8

const (
	ETypeSystem   EType = 0 //系统错误
	ETypePlugin   EType = 1 //插件错误
	ETypeDatabase EType = 2 //数据库错误
)
```

### 全部成功
这个函数用于在 Elastic 批量插入新数据时，由于 Elastic 支持部分成功，所以返回 `[]*proto.ModRet` 来接收每一条数据的插入结果，
可以用 IsAllSuccess 去判断数据是否全部插入成功， 我们可以遍历返回结果，`status` 为错误码，当 `status!=0` 则该条记录插入失败，
`reason` 为失败原因，这样，我们可以针对失败的记录 做特殊处理，比如重试。
```go
import (
    ...
    "github.com/horm-database/go-horm/horm"
    "github.com/horm-database/common/types"
)

func queryIsAllSuccess(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1967-08-27")

	datas := []*Student{
		{
			Id:       1,
			Identify: 2024061211,
			Gender:   1,
			Age:      67,
			Name:     "wigderson",
			Score:    98.3,
			Image:    []byte("IMAGE.PCG"),
			Article:  "enhanced human understanding of the role of randomness and pseudo-randomness in computing.",
			ExamTime: "14:30:00",
			Birthday: types.Time(birthday),
		},
		{
			Id:       2,
			Identify: 2024070733,
			Gender:   2,
			Age:      59,
			Name:     "liskov",
			Score:    99.1,
			Image:    []byte("IMAGE.PCG"),
			Article:  "practice and theory of programming language and systems design",
			ExamTime: "11:30:00",
			Birthday: types.Time(birthday),
		},
	}

	modRets := make([]*proto.ModRet, 0)
	_, err := horm.NewQuery("es_student").Insert(&datas).Exec(ctx, &modRets)
	if err != nil {
		fmt.Printf("batch insert student error: %v", err)
		return
	}

	if horm.IsAllSuccess(modRets) {
		fmt.Printf("batch insert success")
		return
	}
}
```

返回结果：
```json
[
  {
    "id": "Ay7DApQBdHFFOkFBRxKQ",
    "rows_affected": 1,
    "version": 1,
    "status": 0
  },
  {
    "id": "BC7DApQBdHFFOkFBRxKQ",
    "rows_affected": 1,
    "version": 1,
    "status": 0
  }
]
```

ModRet 的结构体如下：
```go
// ModRet 新增/更新返回信息
type ModRet struct {
	ID          ID                     `orm:"id,omitempty" json:"id,omitempty"`                       // id 主键，可能是 mysql 的最后自增id，last_insert_id 或 elastic 的 _id 等，类型可能是 int64、string
	RowAffected int64                  `orm:"rows_affected,omitempty" json:"rows_affected,omitempty"` // 影响行数
	Version     int64                  `orm:"version,omitempty" json:"version,omitempty"`             // 数据版本
	Status      int                    `orm:"status,omitempty" json:"status,omitempty"`               // 返回状态码
	Reason      string                 `orm:"reason,omitempty" json:"reason,omitempty"`               // mod 失败原因
	Extras      map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"`               // 更多详细信息
}

type ID string

func (id ID) String() string
func (id ID) Float64() float64
func (id ID) Int() int
func (id ID) Int64() int64
func (id ID) Uint()
func (id ID) Uint64()
```

上面语句在 es 插入了两条数据，如下，我们可以看到 updated_at 和 created_at 的时间格式，在未指定 time_fmt 的情况下，时间会被编码成
RFC3339 格式，如果希望修改格式，可以指定 time_fmt，但是struct的接收字段类型必须是 types.Time，否则在 Find 的时候，Receive 解析会异常。
```eslint
GET /es_student/_search
{
  "query": {
    "match_all": {}
  }
}
```
```json
{
  "took": 0,
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 2,
      "relation": "eq"
    },
    "max_score": 1,
    "hits": [
      {
        "_index": "es_student",
        "_type": "_doc",
        "_id": "z6SONpQBT1ym-Bx5C67P",
        "_score": 1,
        "_source": {
          "age": 67,
          "article": "enhanced human understanding of the role of randomness and pseudo-randomness in computing.",
          "birthday": "1967-08-27",
          "created_at": "2025-01-05T20:58:37.585539+08:00",
          "exam_time": "14:30:00",
          "gender": 1,
          "id": 1,
          "identify": 2024061211,
          "image": "SU1BR0UuUENH",
          "name": "wigderson",
          "score": 98.3,
          "updated_at": "2025-01-05T20:58:37.585526+08:00"
        }
      },
      {
        "_index": "es_student",
        "_type": "_doc",
        "_id": "0KSONpQBT1ym-Bx5C67P",
        "_score": 1,
        "_source": {
          "age": 59,
          "article": "practice and theory of programming language and systems design",
          "birthday": "1967-08-27",
          "created_at": "2025-01-05T20:58:37.585541+08:00",
          "exam_time": "11:30:00",
          "gender": 2,
          "id": 2,
          "identify": 2024070733,
          "image": "SU1BR0UuUENH",
          "name": "liskov",
          "score": 99.1,
          "updated_at": "2025-01-05T20:58:37.585543+08:00"
        }
      }
    ]
  }
}
```
# 查询语句
## 指定查询列
通过 `Column` 指定要查询的列。
示例1：
```go
func querySpecifiedColumn(ctx context.Context) {
	var results = []map[string]interface{}{}
	isNil, err := horm.NewQuery("student").
		Column("id", "identify", "gender", "age", "name").
		FindAll().Exec(ctx, &results)

	...
}
```
SQL语句：
```sql
 SELECT `identify` , `gender` , `age` , `name`  FROM `student`
```
返回结果：
```json
[
  {
    "identify": 2024080313,
    "name": "caohao",
    "gender": 1,
    "age": 23
  },
  {
    "identify": 2024070733,
    "name": "jerry",
    "gender": 1,
    "age": 17
  },
  {
    "identify": 2024080313,
    "name": "wigderson",
    "gender": 2,
    "age": 23
  }
]
```
示例2：
```go
func querySpecifiedColumn2(ctx context.Context) {
	var result = map[string]interface{}{}
	isNil, err := horm.NewQuery("student").
		Column("count(1) as cnt", "avg(age) as age", "sum(score) as score").
		Find().Exec(ctx, &result)

	...
}
```
SQL语句：
```sql
 SELECT count(1) as cnt , avg(age) as age , sum(score) as score  FROM `student` LIMIT 1
```
返回结果：
```json
{
    "cnt": 5,
    "age": 21,
    "score": 456.5
}
```

## 主键查询
### FindBy
- 示例1 mysql 主键查询：
```go
func queryFindBy(ctx context.Context) {
	var result = Student{}
	isNil, err := horm.NewQuery("student").FindBy("identify", 2024080313).Exec(ctx, &result)

	...
}
```
SQL语句：
```sql
 SELECT * FROM `student` WHERE  `identify` = 2024080313  LIMIT 1
```
返回结果：
```json
{
    "birthday": "1995-03-23T00:00:00+08:00",
    "created_at": "2024-11-30T20:53:57+08:00",
    "updated_at": "2024-12-12T19:30:37+08:00",
    "age": 23,
    "name": "caohao",
    "gender": 1,
    "score": 91.5,
    "image": "SU1BR0UuUENH",
    "article": "groundbreaking work in cryptography and complexity theory",
    "exam_time": "15:30:00",
    "id": 234047220842770433,
    "identify": 2024080313
}
```

- 示例2 mysql 主键查询：
```go
func queryFindAllBy(ctx context.Context) {
	var result = []*Student{}
	isNil, err := horm.NewQuery("student").
		FindAllBy("identify", []int{2024080313, 2024092316}).Exec(ctx, &result)

	...
}
```
SQL语句：
```sql
SELECT * FROM `student` WHERE `identify` IN (2024080313, 2024092316)  LIMIT 100
```

### Elastic主键
Elastic 默认在插入数据的时候会自动生成主键值，他的主键为 `_id`，按照 `_id` 批量插入有几种方式：
- orm 标签加上 `es_id` 属性会指定字段值作为 es 的主键，如下 `id` 字段：
```go
type Student struct {
	Id        uint64     `orm:"id,uint64,onuniqueid,es_id,omitempty" json:"id"`
	Identify  int64      `orm:"identify,int64,omitempty" json:"identify"`                 //学生编号
	Gender    int8       `orm:"gender,int8,omitinsertempty" json:"gender"`                //1-male 2-female
	Age       uint       `orm:"age,uint,omitreplaceempty" json:"age"`                     //年龄
	Name      string     `orm:"name,string,omitupdateempty" json:"name"`                  //名称
	Score     float64    `orm:"score,double,omitempty" json:"score"`                      //分数
	Image     []byte     `orm:"image,bytes,omitempty" json:"image"`                       //image
	Article   string     `orm:"article,string,omitempty" json:"article"`                  //publish article
	ExamTime  string     `orm:"exam_time,string,omitempty" json:"exam_time"`              //考试时间
	Birthday  types.Time `orm:"birthday,time,time_fmt='2006-01-02'" json:"birthday"`      //出生日期
	CreatedAt time.Time  `orm:"created_at,time,oninserttime,omitempty" json:"created_at"` //创建时间
	UpdatedAt time.Time  `orm:"updated_at,time,onupdatetime" json:"updated_at"`           //修改时间
}


func insertEsByID(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1976-08-27")

	datas := []*Student{
		{
			Identify: 2024061211,
			Gender:   2,
			Age:      39,
			Name:     "metcalfe",
			Score:    93.8,
			Image:    []byte("IMAGE.PCG"),
			Article:  "contribution to leading the public into the era of hyper-connectivity",
			ExamTime: "15:30:00",
			Birthday: types.Time(birthday),
		},
		{
			Identify: 2024070733,
			Gender:   2,
			Age:      36,
			Name:     "emerson",
			Score:    79.9,
			Image:    []byte("IMAGE.PCG"),
			Article:  "develop automated methods to detect design errors in computer hardware and software",
			ExamTime: "15:30:00",
			Birthday: types.Time(birthday),
		},
	}

	modRets := make([]*proto.ModRet, 0)
	_, err := horm.NewQuery("es_student").Insert(&datas).Exec(ctx, &modRets)
	if err != nil {
		fmt.Printf("batch insert student error: %v", err)
		return
	}

	if horm.IsAllSuccess(modRets) {
		fmt.Printf("batch insert success")
		return
	}
}
```
返回结果：
```json
[
  {
    "id": "234062949419855873",
    "rows_affected": 1,
    "version": 1
  },
  {
    "id": "234062949419855874",
    "rows_affected": 1,
    "version": 1
  }
]
```

- 也可以如下在 `ID()` 里面指定主键：
```go
func insertEsByID2(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1976-08-27")

	data := Student{
		Id:       888,
		Identify: 2024061211,
		Gender:   1,
		Age:      78,
		Name:     "Alen Joy",
		Score:    99.9,
		Image:    []byte("IMAGE.PCG"),
		Article:  "UNIX operating system and C programming language",
		ExamTime: "16:30:00",
		Birthday: types.Time(birthday),
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("es_student").ID(888).Insert(&data).Exec(ctx, &modRet)

	...
}
```
生成的 es 语句如下：
```eslint
PUT /es_student/_doc/888?op_type=create&refresh=false
{
    "age": 78,
    "article": "UNIX operating system and C programming language",
    "birthday": "1976-08-27",
    "created_at": "2025-01-09T15:37:54.219903+08:00",
    "exam_time": "16:30:00",
    "gender": 1,
    "id": 888,
    "identify": 2024061211,
    "image": "SU1BR0UuUENH",
    "name": "Alen Joy",
    "score": 99.9,
    "updated_at": "2025-01-09T15:37:54.219914+08:00"
}
```
返回结果如下：
```json
{
    "id": "888",
    "rows_affected": 1,
    "version": 1
}
```

- 也可以直接在 Student 结构体加个 `_id` 字段。作为主键：
```go
func insertEsByID3(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1976-08-27")

	type EsStudent struct {
		EsID uint64 `orm:"_id,uint64" json:"_id"`
		Student
	}

	data := EsStudent{
		EsID: 999,
		Student: Student{
			Id:       999,
			Identify: 2024070733,
			Gender:   1,
			Age:      46,
			Name:     "Alen Robot",
			Score:    88.3,
			Image:    []byte("IMAGE.PCG"),
			Article:  "Created tools and textbooks used by millions of programmers worldwide, advanced algorithms and theory underlying programming language implementation, and summarized these findings in influential books",
			ExamTime: "12:00:00",
			Birthday: types.Time(birthday),
		},
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("es_student").Insert(&data).Exec(ctx, &modRet)
	...
}

```

- Elastic 按照 _id 查询
```go
func queryByID(ctx context.Context) {
	var result = Student{}
	isNil, err := horm.NewQuery("es_student").ID(999).Find().Exec(ctx, &result)

	...
}
```
请求如下：
```json
[
    {
        "name": "es_student",
        "op": "find",
        "where": {
            "_id": 999
        }
    }
]
```

### snowflake 唯一ID
当我们将struct标签属性 onuniqueid 打开时，插入数据需要用到 snowflake.GenerateID() 自动生成ID，这时必须为每台机器设置不同的 machine_id，
否则生成的ID可能会有冲突。 利用该算法，我们可以生成一个全局唯一的 ID，而且 ID 里面包含了时间、机器和自增序列信息。虽然不含业务信息，
但是可以通过时间，我们可能会去找到对应的存储分区信息，冷热分离的时候，比如3天以上的数据存在冷分区，我们可以通过 ID 就知道该数据存在哪个分区。
```go
// github.com/horm-database/common/snowflake
func SetMachineID(machineID int)
func GenerateID() uint64
func ParseID(id uint64) (t time.Time, machineID int, sn uint64)
```

我们如果要用到 snowflake 自动生成主键 ID，那么务必在 orm.yaml 配置里面加上 machine_id 来防止碰撞。
```yaml
machine: server.access.gz003      # 本地机器名（容器名）
machine_id: 3                     # 本地机器编号（容器编号），当我们将struct标签属性 onuniqueid 打开时，插入数据需要用到 snowflake.GenerateID() 自动生成ID，这时必须为每台机器设置不同的 machine_id，否则生成的ID可能会有冲突
local_ip: 127.0.0.1               # 本地IP，容器内为容器ip，物理机或虚拟机为本机ip
...
```
## where 查询条件
### 操作符
```go
const ( // 操作符
	OPEqual          = "="   // 等于
	OPBetween        = "()"  // 在某个区间
	OPNotBetween     = "!()" // 不在某个区间
	OPGt             = ">"   // 大于
	OPGte            = ">="  // 大于等于
	OPLt             = "<"   // 小于
	OPLte            = "<="  // 小于等于
	OPNot            = "!"   // 去反
	OPLike           = "~"   // like语句，（或 es 的部分匹配）
	OPNotLike        = "!~"  // not like 语句，（或 es 的部分匹配排除）
	OPMatchPhrase    = "?"   // es 短语匹配 match_phrase
	OPNotMatchPhrase = "!?"  // es 短语匹配排除 must_not match_phrase
	OPMatch          = "*"   // es 全文搜索 match 语句
	OPNotMatch       = "!*"  // es 全文搜索排除 must_not match
)
```

### 基础用法
由于篇幅问题，下面所有用法都是用 mysql 举例，如果对应库类型为 elastic 则数据统一接入服务会生成对应的 es 请求。
```go
func queryWhere(ctx context.Context) {
	var result = []*Student{}

	where := horm.Where{}
	where["age"] = 29                           //`age` = 29
	where["age >"] = 29                         //`age` > 29
	where["age <="] = 39                        //`age` <= 29
	where["age !"] = 29                         //`age` != 29
	where["age ()"] = []int{20, 29}             //`age` BETWEEN 20 AND 29
	where["age !()"] = []int{35, 40}            // NOT ( `age` BETWEEN 35 AND 40)
	where["score"] = []int{60, 61, 62}          //`score` IN (60, 61, 62)
	where["score !"] = []int{70, 71, 72}        //`score` NOT IN (70, 71, 72)
	where["name"] = nil                         //`name` IS NULL
	where["name !"] = nil                       //`name` IS NOT NULL
	where["name ! #注释：排除smallhow"] = "smallhow" //`name` != 'smallhow'

	isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)
	
	...
}
```

Find 查询单条记录：
```go
func queryFind(ctx context.Context) {
	var result = Student{}
	where := horm.Where{"identify": 2024080313}
	isNil, err := horm.NewQuery("student").Find(where).Exec(ctx, &result)

	...
}
```

### 组合查询
针对快速构建 where 语句方式，我们也支持通过 "AND" 或者 "OR"、"NOT" 来组合更复杂的语句。

- 示例1：
```go
func queryCompWhere(ctx context.Context) {
	var result = []*Student{}

	var where = horm.Where{
		"age >":   18,
		"score >": 85,
		"OR": horm.Where{
			"id ()":  []int{1, 100},
			"gender": 1,
		},
	}

	isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

	//不设置 limit 时默认取100条
	// SELECT * FROM `student` WHERE  `age` > 18 AND `score` > 85  AND (( `id`  BETWEEN 1 AND 100)  OR `gender` = 1)  LIMIT 100
	
	...
}
```
上述语句如果转化为 elastic search 的 query 条件语句，则为（es 占用篇幅较大，后面都以 MySQL 为例）：
```json
{
  "from": 0,
  "query": {
    "bool": {
      "filter": [
        {
          "bool": {
            "should": [
              {
                "range": {
                  "id": {
                    "from": 1,
                    "include_lower": true,
                    "include_upper": true,
                    "to": 100
                  }
                }
              },
              {
                "terms": {
                  "gender": [
                    1
                  ]
                }
              }
            ]
          }
        },
        {
          "range": {
            "age": {
              "from": 18,
              "include_lower": false,
              "include_upper": true,
              "to": null
            }
          }
        },
        {
          "range": {
            "score": {
              "from": 85,
              "include_lower": false,
              "include_upper": true,
              "to": null
            }
          }
        }
      ]
    }
  },
  "size": 100
}
```
- 示例2，用注释来区分相同 key：
```go
func queryCompWhere2(ctx context.Context) {
	result := make([]*Student, 0)

	//注意：由于horm.Where是map参数，所以在下面的情况下，第一个 OR 会被覆盖。
	where := horm.Where{
		"OR": horm.Where{
			"id >":   3,
			"gender": 1,
		},
		"OR": horm.Where{
			"identify !": 0,
			"age >=":     20,
		},
	}

	// [X] SELECT * FROM `student` WHERE (`identify`!=0 OR `age`>=20)

	where = horm.Where{
		"OR #注释1": horm.Where{
			"id >":   3,
			"gender": 1,
		},
		"OR #注释2": horm.Where{
			"identify !": 0,
			"age >=":     20,
		},
	}

	// [√]  SELECT * FROM `student` WHERE (`id` > 3  OR `gender` = 1)  AND (`identify` != 0 OR `age` >= 20) LIMIT 100

	where = horm.Where{
		"OR": horm.Where{
			"article ~ #1": "%computer%",
			"article ~ #2": "%medical%",
			"article ~ #3": "%physic%",
		},
	}

	// [√]  SELECT * FROM `student` WHERE  (`article` LIKE '%computer%' OR `article` LIKE '%medical%' OR `article` LIKE '%physic%')  LIMIT 100

	// 上面 like 等同于
	where = horm.Where{
		"article ~": []string{
			"%computer%",
			"%medical%",
			"%physic%",
		}}
	
	isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

	...
}
```
- 示例3
```go
func queryCompWhere3(ctx context.Context) {
	result := make([]*Student, 0)

	where := horm.Where{
		"NOT": horm.Where{
			"id >":   3,
			"gender": 1,
		},
	}

	// SELECT * FROM `student` WHERE NOT (`id` > 3 AND `gender` = 1)  LIMIT 100

	isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

	...
}
```
- 示例 4， OR 下边的 map 数组，map里面各元素默认为 AND：
```go
func queryCompWhere4(ctx context.Context) {
	result := make([]*Student, 0)

	where := horm.Where{
		"OR": []horm.Where{
			{
				"id >":   3,
				"gender": 1,
			},
			{
				"id <":   100,
				"gender": 2,
			},
			{
				"id >=":  999,
				"gender": 2,
			},
		},
	}

	// SELECT * FROM `student` WHERE ((`id` > 3 AND `gender` = 1) OR (`gender` = 2 AND `id` < 100) OR (`id` >= 999 AND `gender` = 2)) LIMIT 100

	isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

	...
}
```
### 模糊查询
#### SQL LIKE 
在数据库引擎为 sql 相关系统时，`~` 操作符表示 LIKE。
- 示例1：
```go
func queryLike(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ~"] = "%cao%"                               // `name` LIKE '%cao%'
	where["birthday ~"] = []string{"2019-08%", "2020-01%"}  // (`birthday` LIKE '2019-08%' OR `birthday` LIKE '2020-01%')
	where["name !~"] = "%cao%"                              // `name` NOT LIKE '%cao%'
	where["birthday !~"] = []string{"2019-08%", "2020-01%"} // (`birthday` NOT LIKE '2019-08%' AND `birthday` NOT LIKE '2020-01%')  ## 注意他和 LIKE 的连接词不一样，NOT LIKE 是 AND，而 LIKE 是 OR

	isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

	fmt.Println(isNil, err)
}
```
- 示例2：
```go
func queryLike2(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ~"] = "Londo_"   // London, Londox, Londos...
	where["name ~"] = "[BCR]at"  // Bat, Cat, Rat
	where["name ~"] = "[!BCR]at" // Eat, Fat, Hat...

	isNil, err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

	...
}
```

#### Elastic 部分匹配
不同于 sql 相关数据库，在 elastic 中，`~` 操作符表示部分匹配。部分匹配分3种类型，prefix（默认）、wildcard、regexp

- prefix 前缀查询（默认）
<br><br>

类似 mysql 的 like 'jerry%'，以 jerry 开头的所有内容。
```go
func queryPrefix(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{"name ~": "jerry"} // jerry, jerrycao, jerrybao...
	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "prefix": {
          "name": "jerry"
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

- wildcard 通配符查询
<br><br>

如下我们为 `~` 操作符加上了 `wildcard` 属性，它使用标准的 shell 通配符查询： `?` 匹配任意字符， `*` 匹配 0 或多个字符。
```go
func queryWildcard(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ~(wildcard)"] = "j*r?y" // jerry, jriy, jasteriy...

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "wildcard": {
          "name": {
            "value": "j*r?y"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

- regexp 正则表达式查询
<br><br>

这个是正则查询，如下示例的正则表达式要求词必须以 W 开头，紧跟 0 至 9 之间的任何一个数字，然后接一或多个其他字符。
```go
func queryRegexp(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ~(regexp)"] = "W[0-9].+"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	
	...
}
```

生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "regexp": {
          "name": {
            "value": "W[0-9].+"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

- NOT 部分匹配排除
```go
func queryNotPrefix(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name !~"] = "cao" // 不以 cao 开头

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```

生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must_not": {
        "prefix": {
          "name": "cao"
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

### 短语匹配 match_phrase
在 Elastic Search 中， `match_phrase` 查询首先将查询字符串解析成一个`词项列表`，然后对这些词项进行搜索，
但只保留那些包含`全部搜索词项`，且`位置`与搜索词项相同的文档。在 horm ，我们用 `?` 操作符表示短语匹配。 `!?` 表示短语匹配排除。

```go
func queryMatchPhrase(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article ?"] = "programming"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match_phrase": {
          "article": {
            "query": "programming"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

#### 灵活度 slop
精确短语匹配或许是过于严格了。也许我们想要包含 “develop automated methods” 的文档也能够匹配 “develop methods”，
可以为`?`操作加上 `slop` 属性如下：
```go
func queryMatchPhraseSlop(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article ?(slop=1)"] = "develop methods"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match_phrase": {
          "article": {
            "query": "develop methods",
            "slop": 1
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

#### 提升权重
我们可以通过指定 `boost` 属性来控制任何查询语句的相对的权重，`boost` 的默认值为 `1` ，大于 `1` 会提升一个语句的相对权重。
如下，name 中包含"caohao"的话，权重更高。那么他可能会拥有更高的 `_score`评分。
```go
func queryMatchPhraseBoost(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ?(boost=3)"] = "caohao"
	where["article ?(boost=2)"] = "complexity"
	where["exam_time ?(boost=1)"] = "15"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Order("_score desc").Exec(ctx, &result)

	...
}

```
生成的 es 查询语句：
```json
{
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "match_phrase": {
            "exam_time": {
              "boost": 1,
              "query": "15"
            }
          }
        },
        {
          "match_phrase": {
            "name": {
              "boost": 3,
              "query": "caohao"
            }
          }
        },
        {
          "match_phrase": {
            "article": {
              "boost": 2,
              "query": "complexity"
            }
          }
        }
      ]
    }
  },
  "size": 100,
  "sort": [
    {
      "_score": {
        "order": "desc"
      }
    }
  ]
}
```

#### 多操作属性
一个 where 操作符可以拥有多个操作属性，通过逗号 `','` 来分隔，如下 article 有 slop 和 boost 两个条件属性。
```go
func queryMatchPhraseSlopBoost(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article ?(slop=2,boost=1)"] = "develop methods"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match_phrase": {
          "article": {
            "boost": 1,
            "query": "develop methods",
            "slop": 2
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

#### 短语匹配排除
操作符 `!?` 表示短语匹配排除
```go
func queryNotMatchPhrase(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article !?"] = "develop" //文章不包含 develop 的记录

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must_not": {
        "match_phrase": {
          "article": {
            "query": "develop"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
### 全文检索 match
当数据库为 Elastic 时，`*` 操作符表示对字段进行全文检索，在全文字段中搜索到最相关的文档。
```go
func queryMatch(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article *"] = "contribution to"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match": {
          "article": {
            "query": "contribution to"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

#### 提高精度 operator
上述例子，中文分词会将`contribution to`分为`contribution`、`to`， 用任意查询词项匹配文档可能会导致结果中出现不相关的长尾，这是种散弹式
搜索，可能我们只想搜索包含`所有词项`的文档，也就是说，不去匹配 `contribution OR to` ，而通过匹配 `contribution AND to`找到所有文档。
`*` 可以加上 operator 属性，默认情况下该属性是 `or`。我们可以将它修改成 `and` 让所有指定词项都必须匹配：
```go
func queryMatchOperator(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article *(and)"] = "contribution to"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
 ```json
{
  "query": {
    "bool": {
      "must": {
        "match": {
          "article": {
            "operator": "and",
            "query": "contribution to"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
#### 控制精度 minimum_should_match
在所有与任意间二选一有点过于非黑即白。如果用户给定 5 个查询词项，想查找只包含其中 4 个的文档，该如何处理？
在全文搜索的大多数应用场景下，我们既想包含那些可能相关的文档，同时又排除那些不太相关的。换句话说，我们想要处于中间某种结果。
`*` 查询支持 `minimum_should_match` 最小匹配属性，这让我们可以指定必须匹配的词项数用来表示一个文档是否相关。
我们可以将其设置为某个具体数字，更常用的做法是将其设置为一个百分数，因为我们无法控制用户搜索时输入的单词数量，如下，
我们设置最小匹配参数为 40%，即只需要命中至少2个词，则匹配文档。

```go
func queryMatchMinimumShouldMatch(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article *(minimum_should_match=40%)"] = "contribution to lead the public"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	fmt.Println(isNil, err)
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match": {
          "article": {
            "minimum_should_match": "40%",
            "query": "contribution to lead the public"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
#### 评分计算
`bool` 查询会为每个文档计算相关度评分 `_score`，再将所有匹配的 `must` 和 `should` 语句的分数 `_score` 求和，最后除以 `must` 和  
`should` 语句的总数。`must_not`  语句不会影响评分；它的作用只是将不相关的文档排除。

#### 提升权重
提升权重与 `match_phrase` 里的用法是一样的，也是通过指定 `boost` 来控制任何查询语句的相对的权重，`boost` 的默认值为 `1`，大于 `1` 会
提升一个语句的相对权重。如下，name 中包含"caohao"的话，权重更高。那么他可能会拥有更高的 `_score`评分。
```go
func queryMatchBoost(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ?(boost=3)"] = "caohao"
	where["article ?(boost=2)"] = "work in"
	where["exam_time ?(boost=1)"] = "15"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": [
        {
          "match_phrase": {
            "article": {
              "boost": 2,
              "query": "work in"
            }
          }
        },
        {
          "match_phrase": {
            "exam_time": {
              "boost": 1,
              "query": "15"
            }
          }
        },
        {
          "match_phrase": {
            "name": {
              "boost": 3,
              "query": "caohao"
            }
          }
        }
      ]
    }
  },
  "from": 0,
  "size": 100
}
```

#### 全文搜索排除
```go
func queryNotMatch(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["article !*"] = "contribution to"

	isNil, err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	...
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must_not": {
        "match": {
          "article": {
            "query": "contribution to"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

## 分组、聚合（暂未支持 elastic 的聚合）
### GROUP
通过 Group(group ...string) 支持分组查询。如下查看大于10岁的学生中，各个性别、年龄段分组的学生总数和平均分：
```go
func queryGroupBy(ctx context.Context) {
	result := make([]map[string]interface{}, 0)

	var where = horm.Where{"age >": 10}

	isNil, err := horm.NewQuery("student").
		Column("gender", "age", "count(1) as cnt", "avg(score) as score_avg").
		FindAll(where).Group("gender", "age").Exec(ctx, &result)

	// SELECT `gender`, `age`, count(1) as cnt, avg(score) as score_avg FROM `student` WHERE  `age` > 10  GROUP BY `gender`,`age` LIMIT 100
	...
}
```
查询结果：
```json
[
  {
    "gender": 1,
    "age": 23,
    "cnt": 1,
    "score_avg": 91.5
  },
  {
    "gender": 1,
    "age": 17,
    "cnt": 1,
    "score_avg": 82.5
  }
]
```

### HAVING
有些场景，我们需要在 group by 分组之后，根据分组聚合数据进行再次过滤，这时候我们需要用到 having，例如下面我们需要查询平均分大于90分的
年龄段、性别分组的学生总数和平均分，Having 函数的参数与 where 条件一样，解析规则也是一样。

```go
func queryGroupByHaving(ctx context.Context) {
	result := make([]map[string]interface{}, 0)

	isNil, err := horm.NewQuery("student").
		Column("gender", "age", "count(1) as cnt", "avg(score) as score_avg").
		FindAll().Group("gender", "age").
		Having(horm.Where{"score_avg >": 90}).Exec(ctx, &result)

	// SELECT `gender`, `age`, count(1) as cnt, avg(score) as score_avg FROM `student` GROUP BY `gender`,`age` HAVING  `score_avg` > 90  LIMIT 100
	...
}

```

## 排序与分页
### ORDER 排序
通过 `Order` 函数指定排序。
```go
func queryOrder(ctx context.Context) {
	var result = []*Student{}
	isNil, err := horm.NewQuery("student").FindAll().Order("+age", "-score").Exec(ctx, &result)       // ORDER BY age ASC, score DESC
	isNil, err = horm.NewQuery("student").FindAll().Order("age asc", "score desc").Exec(ctx, &result) // ORDER BY age ASC, score DESC
	isNil, err = horm.NewQuery("student").FindAll().Order("age").Exec(ctx, &result)                   // ORDER BY age

	...
}
```

Elastic 按照相关性评分排序：
```go
func queryOrder2(ctx context.Context) {
	var result = []*Student{}

	where := horm.Where{"article *": "contribution to"}
	isNil, err := horm.NewQuery("es_student").FindAll(where).Order("_score desc").Exec(ctx, &result)

	...
}
```
请求如下：
```json
{
    "query": {
        "bool": {
            "must": {
                "match": {
                    "article": {
                        "query": "contribution to"
                    }
                }
            }
        }
    },
    "from": 0,
    "size": 100,
    "sort": [
        {
            "_score": {
                "order": "desc"
            }
        }
    ]
}
```
返回如下，根据 _elastic._score 全文检索匹配相关性评分从高到底排序。
```json
[
    {
        "exam_time": "15:30:00",
        "age": 39,
        "article": "contribution to leading the public into the era of hyper-connectivity",
        "created_at": "2025-01-05T21:22:35.821669+08:00",
        "gender": 2,
        "score": 93.8,
        "_elastic": {
            "_score": 2.1302004,
            "_index": "es_student",
            "_id": "234062949419855873"
        },
        "birthday": "1976-08-27",
        "id": 234062949419855873,
        "identify": 2024061211,
        "image": "SU1BR0UuUENH",
        "name": "metcalfe",
        "updated_at": "2025-01-05T21:22:35.821654+08:00"
    },
    {
        "article": "contributions to deep learning in artificial intelligence",
        "exam_time": "15:30:00",
        "identify": 2024092316,
        "image": "SU1BR0UuUENH",
        "updated_at": "2025-01-05T20:33:33.041235+08:00",
        "gender": 1,
        "id": 234050606505930753,
        "name": "jerry",
        "_elastic": {
            "_score": 0.7857686,
            "_index": "es_student",
            "_id": "zqR0NpQBT1ym-Bx53K4b"
        },
        "age": 17,
        "created_at": "2025-01-05T20:33:33.04126+08:00",
        "score": 82.5,
        "birthday": "1995-03-24"
    },
    {
        "article": "develop automated methods to detect design errors in computer hardware and software",
        "_elastic": {
            "_score": 0.6358339,
            "_index": "es_student",
            "_id": "234062949419855874"
        },
        "created_at": "2025-01-05T21:22:35.82168+08:00",
        "exam_time": "15:30:00",
        "identify": 2024070733,
        "image": "SU1BR0UuUENH",
        "birthday": "1976-08-27",
        "gender": 2,
        "id": 234062949419855874,
        "age": 36,
        "name": "emerson",
        "score": 79.9,
        "updated_at": "2025-01-05T21:22:35.821675+08:00"
    }
]
```

### LIMIT、OFFSET
通过 `Limit` 函数去指定 limit 、offset 参数。
```go
func queryLimitSize(ctx context.Context) {
	var result = []*Student{}
	isNil, err := horm.NewQuery("student").FindAll().Limit(10).Exec(ctx, &result)    // LIMIT 10
	isNil, err = horm.NewQuery("student").FindAll().Limit(10, 30).Exec(ctx, &result) // LIMIT 10 OFFSET 30

	...
}
```

### 分页 PAGE
通过 Page(page, pageSize int) 函数请求分页数据，统一接入服务返回的分页数据结构如下，具体案例可以参考查询模式-单查询单元-分页返回章节。

```go
// PageResult 当 page > 1 时会返回分页结果
type PageResult struct {
	Detail *Detail       `orm:"detail,omitempty" json:"detail,omitempty"` // 查询细节信息
	Data   []interface{} `orm:"data,omitempty" json:"data,omitempty"`     // 分页结果
}

// Detail 其他查询细节信息，例如 分页信息、滚动翻页信息、其他信息等。
type Detail struct {
	Total     uint64                 `orm:"total" json:"total"`                               // 总数
	TotalPage uint32                 `orm:"total_page,omitempty" json:"total_page,omitempty"` // 总页数
	Page      int                    `orm:"page,omitempty" json:"page,omitempty"`             // 当前分页
	Size      int                    `orm:"size,omitempty" json:"size,omitempty"`             // 每页大小
	Scroll    *Scroll                `orm:"scroll,omitempty" json:"scroll,omitempty"`         // 滚动翻页信息
	Extras    map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"`         // 更多详细信息
}
```

## 返回结果高亮
在 Elastic Search 中，我们可以请求 es 将我们的检索结果中的关键词打上高亮标签返回，我们可以针对不同的字段打不同的标签，第四个参数 replace 
是一个可选参数，在我们不需要原字段返回，而只需要返回带标签的内容时，将 replace 置为 true，可以减少输出内容，避免返回过大，如下：
```go
func queryHighlight(ctx context.Context) {
	var where = horm.Where{}
	where["article *"] = "contribution to"
	where["exam_time *"] = "15"

	type HighLightStudent struct {
		Student
		HighlightArticle  []string `orm:"highlight_article,string,omitempty" json:"highlight_article"`     //publish article
		HighlightExamTime []string `orm:"highlight_exam_time,string,omitempty" json:"highlight_exam_time"` //考试时间
	}

	result := make([]*HighLightStudent, 0)
	isNil, err := horm.NewQuery("es_student").FindAll(where).
		HighLight("article", "<red>", "</red>", true).
		HighLight("exam_time", "<yellow>", "</yellow>").Exec(ctx, &result)

	...
}
```
生成的请求如下：
```json
{
  "name": "es_student",
  "op": "find_all",
  "where": {
    "article *": "contribution to",
    "exam_time *": "15"
  },
  "size": 100,
  "params": {
    "highlights": [
      {
        "field": "article",
        "pre_tag": "<red>",
        "post_tag": "</red>",
        "replace": true
      },
      {
        "field": "exam_time",
        "pre_tag": "<yellow>",
        "post_tag": "</yellow>"
      }
    ]
  }
}
```
在数据统一接入服务会被转化为如下 Elastic 请求：
```json
{
    "highlight": {
        "fields": {
            "article": {
                "post_tags": [
                    "</red>"
                ],
                "pre_tags": [
                    "<red>"
                ]
            },
            "exam_time": {
                "post_tags": [
                    "</yellow>"
                ],
                "pre_tags": [
                    "<yellow>"
                ]
            }
        }
    },
    "query": {
        "bool": {
            "must": [
                {
                    "match": {
                        "article": {
                            "query": "contribution to"
                        }
                    }
                },
                {
                    "match": {
                        "exam_time": {
                            "query": "15"
                        }
                    }
                }
            ]
        }
    },
  "from": 0,
    "size": 100
}
```

返回结果如下，我们会在高亮字段前面加 `highlight_` 表示该字段为高亮结果，他是一个字符串数组，在结果中我们可以看到，
原来的 article 并未返回，因为 replace 为 true：
```json
[
  {
    "_elastic": {
      "_score": 3.4667747,
      "_index": "es_student",
      "_id": "234062949419855873"
    },
    "image": "SU1BR0UuUENH",
    "age": 39,
    "created_at": "2025-01-05T21:22:35.821669+08:00",
    "gender": 2,
    "identify": 2024061211,
    "score": 93.8,
    "name": "metcalfe",
    "updated_at": "2025-01-05T21:22:35.821654+08:00",
    "birthday": "1976-08-27",
    "id": 234062949419855873,
    "exam_time": "15:30:00",
    "highlight_article": [
      "<red>contribution</red> <red>to</red> leading the public into the era of hyper-connectivity"
    ],
    "highlight_exam_time": [
      "<yellow>15</yellow>:30:00"
    ]
  },
  {
    "_elastic": {
      "_score": 1.8139606,
      "_index": "es_student",
      "_id": "zqR0NpQBT1ym-Bx53K4b"
    },
    "age": 17,
    "score": 82.5,
    "image": "SU1BR0UuUENH",
    "name": "jerry",
    "birthday": "1995-03-24",
    "id": 234050606505930753,
    "identify": 2024092316,
    "created_at": "2025-01-05T20:33:33.04126+08:00",
    "updated_at": "2025-01-05T20:33:33.041235+08:00",
    "gender": 1,
    "exam_time": "15:30:00",
    "highlight_article": [
      "contributions <red>to</red> deep learning in artificial intelligence"
    ],
    "highlight_exam_time": [
      "<yellow>15</yellow>:30:00"
    ]
  },
  {
    "_elastic": {
      "_score": 1.6168463,
      "_index": "es_student",
      "_id": "234062949419855874"
    },
    "age": 36,
    "birthday": "1976-08-27",
    "created_at": "2025-01-05T21:22:35.82168+08:00",
    "identify": 2024070733,
    "name": "emerson",
    "score": 79.9,
    "gender": 2,
    "id": 234062949419855874,
    "updated_at": "2025-01-05T21:22:35.821675+08:00",
    "image": "SU1BR0UuUENH",
    "exam_time": "15:30:00",
    "highlight_article": [
      "develop automated methods <red>to</red> detect design errors in computer hardware and software"
    ],
    "highlight_exam_time": [
      "<yellow>15</yellow>:30:00"
    ]
  }
]
```

# 数据维护
## INSERT 语句
Insert 函数可以插入各种类型的数据，比如 struct、map、struct数组、map数组。

### 插入 struct 数据
`Insert` 函数传参可以是 struct 结构体，详细的结构体及`orm`标签的解释可以参考章节 [结构体标签](#结构体标签)， 返回 `proto.ModRet`，如果不关心返回，可以不传：
```go
type Student struct {
	Id        uint64     `orm:"id,uint64,onuniqueid,omitempty" json:"id"`
	Identify  int64      `orm:"identify,int64,omitempty" json:"identify"`                 //学生编号
	Gender    int8       `orm:"gender,int8,omitinsertempty" json:"gender"`                //1-male 2-female
	Age       uint       `orm:"age,uint,omitreplaceempty" json:"age"`                     //年龄
	Name      string     `orm:"name,string,omitupdateempty" json:"name"`                  //名称
	Score     float64    `orm:"score,double,omitempty" json:"score"`                      //分数
	Image     []byte     `orm:"image,bytes,omitempty" json:"image"`                       //image
	Article   string     `orm:"article,string,omitempty" json:"article"`                  //publish article
	ExamTime  string     `orm:"exam_time,string,omitempty" json:"exam_time"`              //考试时间
	Birthday  types.Time `orm:"birthday,time,time_fmt='2006-01-02'" json:"birthday"`      //出生日期
	CreatedAt time.Time  `orm:"created_at,time,oninserttime,omitempty" json:"created_at"` //创建时间
	UpdatedAt time.Time  `orm:"updated_at,time,onupdatetime" json:"updated_at"`           //修改时间
}
```

- 示例 1：
```go
func insertStruct(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1976-08-27")

	data := Student{
		Identify: 2024061211,
		Gender:   2,
		Age:      39,
		Name:     "metcalfe",
		Score:    93.8,
		Image:    []byte("IMAGE.PCG"),
		Article:  "contribution to leading the public into the era of hyper-connectivity",
		ExamTime: "15:30:00",
		Birthday: types.Time(birthday),
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("student").Insert(&data).Exec(ctx, &modRet)
	
	...
}
```

返回结果：
```json
{
  "id": "235900923765862401",
  "rows_affected": 1
}
```

在 horm，我们用 `proto.ModRet` 来接收单条记录的插入结果，该结构体如下：
```go
// ModRet 新增/更新返回信息
type ModRet struct {
	ID          ID                     `orm:"id,omitempty" json:"id,omitempty"`                       // id 主键，可能是 mysql 的最后自增id，last_insert_id 或 elastic 的 _id 等，类型可能是 int64、string
	RowAffected int64                  `orm:"rows_affected,omitempty" json:"rows_affected,omitempty"` // 影响行数
	Version     int64                  `orm:"version,omitempty" json:"version,omitempty"`             // 数据版本
	Status      int                    `orm:"status,omitempty" json:"status,omitempty"`               // 返回状态码
	Reason      string                 `orm:"reason,omitempty" json:"reason,omitempty"`               // mod 失败原因
	Extras      map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"`               // 更多详细信息
}

type ID string

func (id ID) String() string
func (id ID) Float64() float64
func (id ID) Int() int
func (id ID) Int64() int64
func (id ID) Uint()
func (id ID) Uint64()
```

### 插入 map 数据
`Insert` 还可以传入 map 数据来插入单条数据：
- 示例 1：
```go
func insertMap(ctx context.Context) {
	data := horm.Map{
		"id":         235842198988402689,
		"identify":   2024061211,
		"name":       "metcalfe",
		"gender":     2,
		"age":        39,
		"score":      93.8,
		"article":    "contribution to leading the public into the era of hyper-connectivity",
		"exam_time":  "15:30:00",
		"image":      "SU1BR0UuUENH",
		"birthday":   "1976-08-27",
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("student").Insert(data).Exec(ctx, &modRet)

	...
}
```

返回数据如下：
```json
{
    "id": "235842198988402689",
    "rows_affected": 1
}
```

- 示例2，Elastic 通过 map 插入单条记录：
```go
func insertMapToElastic(ctx context.Context) {
	data := horm.Map{
		"_id":        66666, // Elastic主键
		"id":         66666,
		"identify":   2024061211,
		"name":       "metcalfe",
		"gender":     2,
		"age":        39,
		"score":      93.8,
		"article":    "contribution to leading the public into the era of hyper-connectivity",
		"exam_time":  "15:30:00",
		"image":      "SU1BR0UuUENH",
		"birthday":   "1976-08-27",
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("es_student").Insert(data).Exec(ctx, &modRet)

	...
}
```
返回结果：
```json
{
    "_id":"v03bpIEBL4QnOSO-YOvH",
    "version":1,
    "rows_affected":1
}
```

### 批量插入数据
`Insert` 函数接收的参数为map数组，或者 struct 数组的时候，就会执行批量插入，如果是插入 MySQL 等，则返回 `proto.ModRet`， 包含批量插入
最后一条记录的id和影响行数rows_affected，插入 Elastic 由于支持部分成功，返回的就是 `[]*proto.ModRet` 来接收每一条数据的插入结果。
- 示例 1：
```go
func insertMapArray(ctx context.Context) {
	datas := []horm.Map{
		{
			"id":         235842198988432689,
			"identify":   2024061291,
			"name":       "metcalfe",
			"gender":     2,
			"age":        39,
			"score":      93.8,
			"article":    "contribution to leading the public into the era of hyper-connectivity",
			"exam_time":  "15:30:00",
			"image":      "SU1BR0UuUENH",
			"birthday":   "1976-08-27",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
		{
			"id":         235842198988452699,
			"identify":   2024070746,
			"name":       "emerson",
			"gender":     2,
			"age":        36,
			"score":      79.9,
			"article":    "develop automated methods to detect design errors in computer hardware and software",
			"exam_time":  "15:30:00",
			"image":      "IMAGE.PCG",
			"birthday":   "1976-08-27",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("student").Insert(&datas).Exec(ctx, &modRet)

	...
}
```
返回结果：
```json
{
  "id": "235842198988452699",
  "rows_affected": 2
}
```

- Elastic 批量插入的具体案例可以参考章节 [全部成功](#全部成功)
- Elastic 插入数据的时候指定主键，可以参考章节 [Elastic主键](#Elastic主键)
- 除了 map 数组，批量插入同样也支持 struct 数组。


## REPLACE 语句
replace 和 insert 函数类似，只不过是把 sql 关键词 insert 替换为 replace，函数为 `Replace`，
参数同样支持 struct、map、struct数组、map数组，具体用法可以参考 insert。

`注意：elastic search 不支持 replace`

## UPDATE 语句
Update 函数可以更新各种类型的数据，比如 map、struct。
### 更新 map 数据
- 示例1：
```go
func updateMap(ctx context.Context) {
	data := horm.Map{
		"exam_time":  "09:00:00",
		"age":        49,
		"updated_at": time.Now(),
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("student").
		Update(data).Eq("id", 235842198988452699).Exec(ctx, &modRet)

	...
}
```
返回结果：
```json
{
  "rows_affected": 2
}
```

- 示例2：
```go
func updateMapByWhere(ctx context.Context) {
	data := horm.Map{
		"exam_time": "16:00:00",
		"updated_at": time.Now(),
	}

	where := horm.Where{"age >": 40}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("student").Update(data, where).Exec(ctx, &modRet)
	
	...
}
```

- 示例3（Elastic update by _id）：
```go
func updateMapElasticByID(ctx context.Context) {
	data := horm.Map{
		"exam_time":  "15:45:00",
		"updated_at": time.Now(),
	}

	where := horm.Where{"_id": 234062949419855873}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("es_student").Update(data, where).Exec(ctx, &modRet)

	...
}

```
生成的请求：
```eslint
{
    "script": {
        "params": {
            "exam_time": "15:45:00",
            "updated_at": "2025-01-11T22:24:29.168736+08:00"
        },
        "source": "ctx._source.exam_time=params.exam_time;ctx._source.updated_at=params.updated_at"
    }
}
```

返回结果：
```json
{
  "id": "234062949419855873",
  "rows_affected": 1,
  "version": 2
}
```

- 示例4（Elastic update by query）：
```go
func updateMapElastic(ctx context.Context) {
	data := horm.Map{
		"exam_time":  "16:00:00",
		"updated_at": time.Now(),
	}

	where := horm.Where{"age >": 60}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("es_student").Update(data, where).Exec(ctx, &modRet)

	...
}
```
生成的请求：
```eslint
POST /es_student/_update_by_query?refresh=false
{
  "query": {
    "bool": {
      "filter": {
        "range": {
          "age": {
            "from": 60,
            "include_lower": false,
            "include_upper": true,
            "to": null
          }
        }
      }
    }
  },
  "script": {
    "params": {
      "exam_time": "16:00:00",
      "updated_at": "2025-01-11T22:17:25.282884+08:00"
    },
    "source": "ctx._source.exam_time=params.exam_time;ctx._source.updated_at=params.updated_at"
  }
}
```
返回结果：
```json
{
  "rows_affected": 2
}
```

### 更新 struct 数据
omitupdateempty、omitempty、onupdatetime 标签会对 struct 的 update 操作作用。
```go
type Student struct {
	Id        uint64     `orm:"id,uint64,onuniqueid,omitempty" json:"id"`
	Identify  int64      `orm:"identify,int64,omitempty" json:"identify"`                 //学生编号
	Gender    int8       `orm:"gender,int8,omitinsertempty" json:"gender"`                //1-male 2-female
	Age       uint       `orm:"age,uint,omitreplaceempty" json:"age"`                     //年龄
	Name      string     `orm:"name,string,omitupdateempty" json:"name"`                  //名称
	Score     float64    `orm:"score,double,omitempty" json:"score"`                      //分数
	Image     []byte     `orm:"image,bytes,omitempty" json:"image"`                       //image
	Article   string     `orm:"article,string,omitempty" json:"article"`                  //publish article
	ExamTime  string     `orm:"exam_time,string,omitempty" json:"exam_time"`              //考试时间
	Birthday  types.Time `orm:"birthday,time,time_fmt='2006-01-02'" json:"birthday"`      //出生日期
	CreatedAt time.Time  `orm:"created_at,time,oninserttime,omitempty" json:"created_at"` //创建时间
	UpdatedAt time.Time  `orm:"updated_at,time,onupdatetime" json:"updated_at"`           //修改时间
}

func updateStruct(ctx context.Context) {
	data := Student{
		Gender:   2,
		Age:      39,
		Score:    93.8,
		Article:  "contribution to leading the public into the era of hyper-connectivity",
		ExamTime: "15:30:00",
	}

	modRet := proto.ModRet{}
	_, err := horm.NewQuery("student").Update(&data).Eq("id", 235842198988432689).Exec(ctx, &modRet)
	fmt.Println(err)
}
```

从下面请求我们可以看出id、identify、image、created_at有 omitempty 属性、name有 omitupdateempty 属性，所以在更新的时候该字段如果
未赋值，则会被忽略，而 birthday 没有 omitempty、omitupdateempty，那么在更新的时候，不会被忽略，会被置为零值，这可能会导致与 预期不一致的
结果，updated_at 虽然未被赋值，但是由于他有 onupdatetime 属性，所以在更新的时候，都会被horm 赋值为当前时间：
```json
{
  "name": "student",
  "op": "update",
  "where": {
    "id": 235842198988432689
  },
  "data": {
    "gender": 2,
    "age": 39,
    "score": 93.8,
    "article": "contribution to leading the public into the era of hyper-connectivity",
    "exam_time": "15:30:00",
    "updated_at": "2025-01-11T23:33:14.430691+08:00",
    "birthday": "0001-01-01"
  }
}
```

其他示例和更新 map 类似，我们主要是关注 omitupdateempty、omitempty、onupdatetime 标签属性对更新的影响。

## DELETE 数据
- 示例1（mysql）
```go
func deleteByWhere(ctx context.Context) {
	where := horm.Where{"name": "metcalfe"}

	ret := proto.ModRet{}
	_, err := horm.NewQuery("student").Delete(where).Exec(ctx, &ret)
	
	...
}
```
返回结果：
```json
{
  "rows_affected": 1
}
```

- 示例2（Elastic delete by query）

```go
func deleteByWhere(ctx context.Context) {
	where := horm.Where{"name": "metcalfe"}

	ret := proto.ModRet{}
	_, err := horm.NewQuery("es_student").Delete(where).Exec(ctx, &ret)
	
	...
}
```

生成的 es 请求：
```eslint
POST /es_student/_doc/_delete_by_query?refresh=false
{
    "query": {
        "bool": {
            "filter": {
                "terms": {
                    "name": ["metcalfe"]
                }
            }
        }
    }
}
```
返回结果：
```json
{
    "rows_affected":2
}
```

- 示例3（Elastic delete by _id）
```go
func deleteEsByID(ctx context.Context) {
	ret := proto.ModRet{}
	_, err := horm.NewQuery("es_student").Delete().ID(999).Exec(ctx, &ret)
	...
}
```

生成的 es 请求:
```eslint
DELETE /es_student/_doc/999?refresh=false
```

返回结果：
```json
{
    "_id":"999",
    "version":2,
    "rows_affected":1
}
```

## refresh 刷新
在 Elastic 通过 `Refresh(true)` 函数可以使数据在更新之后立即被刷新，当然，这个会导致 Elastic Search 的压力增大。
```go
func updateEsByIDRefresh(ctx context.Context) {
	data := horm.Map{
		"exam_time":  "15:45:00",
		"updated_at": time.Now(),
	}

	ret := proto.ModRet{}
	_, err := horm.NewQuery("es_student").Update(data).ID(234062949419855874).Refresh().Exec(ctx, &ret)
	...
}
```

生成的 es 请求：
```eslint
POST /es_student/_update/234062949419855874?refresh=true
{
    "script": {
        "params": {
            "exam_time": "15:45:00",
            "updated_at": "2025-01-12T09:19:57.578219+08:00"
        },
        "source": "ctx._source.exam_time=params.exam_time;ctx._source.updated_at=params.updated_at"
    }
}
```

# redis 协议
## Prefix（非常强烈建议使用）
`Prefix` 可以为我们的 key 加上前缀，如下案例真正的 key 就是 `student_2024080313`，（后面的案例，为了便捷，我都省略了 `Prefix`）。<br>

`强烈建议所有 key 都加上前缀，便于数据统一接入服务根据 Prefix 来对不同的对象的区分统计，如果没有加 Prefix，所有 key 都在一个库里面，
比如有 student_2024080313, teacher_Simon，那么我们如果希望统计每天查询 student 的请求量是多少，则无法统计，也无法更好的定位具体是哪个对象
发生了请求突发暴增的情况。`

```go
func testRedisPrefix(ctx context.Context) {
	var result int
	isNil, err := horm.NewQuery("redis_student").
		Prefix("student_").Get("2024080313").Exec(ctx, &result)

	// GET student_2024080313
	...
}

```

## 编码与解码
我们会有默认的编码器，例如下面的 Set 操作，默认编码器会将 data 编码为 `json` 字符串，然后 set 到 redis，在 Get 函数，
可以传入 `Student 结构体指针`，去接收解码后的返回结果。

```go
func testRedisCoder(ctx context.Context) {
	birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

	_, err := horm.NewQuery("redis_student").Prefix("student_").Set("2024061211", data).Exec(ctx)
	if err != nil {
		fmt.Println("set student to redis error: ", err)
		return
	}

	// SET student_2024061211 {\"age\":39,\"image\":\"SU1BR0UuUENH\",\"id\":236629372784619521,\"article\":\"contribution to leading the public into the era of hyper-connectivity\",\"updated_at\":\"2025-01-12T23:20:38.833523+08:00\",\"identify\":2024061211,\"gender\":2,\"created_at\":\"2025-01-12T23:20:38.833534+08:00\",\"name\":\"metcalfe\",\"score\":93.8,\"exam_time\":\"15:30:00\",\"birthday\":\"1976-08-27\"}

	result := Student{}
	isNil, err := horm.NewQuery("redis_student").Prefix("student_").Get("2024061211").Exec(ctx, &result)
	
	// GET student_2024061211
	...
}

```

horm 所有支持的 redis 操作，一共会返回8种类型的结构，在有编解码的情况下，这几种结构会以合理的方式写入接收对象，如果是复合查询，
则是会将原始结构映射到复合查询的大 json 结果字段里面去：

* 无返回，仅包含 error。这类操作包含 `EXPIRE` 、 `SET` 、  `SETEX` 、 `HSET` 、 `HMSET` ，仅返回 error，如果无 error 则执行成功。
* 返回 `[]byte`，这类操作包含 `GET` 、 `GETSET` 、 `HGET` 、 `LPOP` 、 `RPOP`。
* 返回 `bool`，这类操作包含 `EXISTS` 、 `SETNX` 、 `HEXISTS` 、 `HSETNX` 、 `SISMEMBER`。
* 返回 `int64`，这类操作包含 这类操作包含 `INCR` 、 `DECR` 、 `INCRBY` 、 `HINCRBY` 、 `TTL` 、 `DEL` 、 `HDEL` 、 `HLEN` 、 `LPUSH` 、 `RPUSH` 、 `LLEN` 、 `SADD` 、 `SREM` 、 `SCARD` 、 `SMOVE` 、 `ZADD` 、 `ZREM` 、 `ZREMRANGEBYSCORE` 、 `ZREMRANGEBYRANK` 、 `ZCARD` 、 `ZRANK` 、 `ZREVRANK` 、 `ZCOUNT`。
* 返回 `float64`，这类操作包含 `ZSCORE` 、 `ZINCRBY`。
* 返回 `[][]byte`，这类操作包含 `HKEYS` 、 `SMEMBERS` 、 `SRANDMEMBER` 、 `SPOP` 、 `ZRANGE` 、 `ZRANGEBYSCORE` 、 `ZREVRANGE` 、 `ZREVRANGEBYSCORE`。
* 返回 `map[string]string`，这类操作包含 `HGETALL` 、 `HMGET`。
* 返回 `map[string]float64`，这类操作包含 `ZPOPMIN`
* 返回 `member 和 score（类型为 [][]byte、[]float64）`，这类操作包含 `ZRANGE ... WITHSCORES` 、 `ZRANGEBYSCORE ... WITHSCORES` 、 `ZREVRANGE ... WITHSCORES` 、 `ZREVRANGEBYSCORE ... WITHSCORES`

## Redis 键
后面 redis 案例为了简便，都未加 `Prefix`。
- `EXPIRE`
  函数用法：
```go
// Expire 设置 key 的过期时间，key 过期后将不再可用。单位以秒计。
// param: key string
// param: int seconds 到期时间
Expire(key string, seconds int)
```
使用示例：
```go
_, err := horm.NewQuery("redis_student").Expire("student_2024061211", 3600).Exec(ctx)
```
- `TTL`
  函数用法：
```go
// TTL 以秒为单位返回 key 的剩余过期时间。
// param: string key
TTL(key string) 
```

使用示例：
```go
var ttl int // 当 key 不存在时，返回 -2 。 当 key 存在但没有设置剩余生存时间时，返回 -1 。 否则，以秒为单位，返回 key 的剩余生存时间。
_, err := horm.NewQuery("redis_student").TTL("student_2024061211").Exec(ctx, &ttl)
```

- `EXISTS`
  函数用法：
```go
// Exists 查看值是否存在 exists
// param: key string
Exists(key string)
```
使用示例：
```go
// 存在则 isNil=false，否则 isNil=true
isNil, err := horm.NewQuery("redis_student").Exists("student_2024061211").Exec(ctx)
```
- `DEL`
  函数用法：
```go
// Del 删除已存在的键。不存在的 key 会被忽略。
// param: key string
Del(key string)
```
使用示例：
```go
var affectNum int // 被删除 key 的数量。
isNil, err = horm.NewQuery("redis_student").Del("student_2024061211").Exec(ctx, &affectNum)
```
## 字符串
- `SET`
  函数用法：
```go
// Set 设置给定 key 的值。如果 key 已经存储其他值， Set 就覆写旧值。
// param: key string
// param: value interface{} 任意类型数据
// param: args ...interface{} set的其他参数
Set(key string, value interface{}, args ...interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

_, err = horm.NewQuery("redis_student").Set("test_bool", true).Exec(ctx)
_, err = horm.NewQuery("redis_student").Set("test_int", 78).Exec(ctx)
_, err = horm.NewQuery("redis_student").Set("test_float", 63.2567).Exec(ctx)
_, err = horm.NewQuery("redis_student").Set("test_string", "i am ok").Exec(ctx)
_, err = horm.NewQuery("redis_student").Set("test_struct", data).Exec(ctx)
```
- `SETEX`
  函数用法：
```go
// SetEX 指定的 key 设置值及其过期时间。如果 key 已经存在， SETEX 命令将会替换旧的值。
// param: key string
// param: v interface{} 任意类型数据
// param: ttl int 到期时间
func (s *Statement) SetEX(key string, v interface{}, ttl int)
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

_, err = horm.NewQuery("redis_student").SetEX("test_bool", true, 3600).Exec(ctx)
_, err = horm.NewQuery("redis_student").SetEX("test_int", 78, 3600).Exec(ctx)
_, err = horm.NewQuery("redis_student").SetEX("test_float", 63.2567, 3600).Exec(ctx)
_, err = horm.NewQuery("redis_student").SetEX("test_string", "i am ok", 86400).Exec(ctx)
_, err = horm.NewQuery("redis_student").SetEX("test_struct", data, 86400).Exec(ctx)
```
- `SETNX`
  函数用法：
```go
// SetNX redis.SetNX
// 指定的 key 不存在时，为 key 设置指定的值。
// param: key string
// param: v interface{} 任意类型数据
SetNX(key string, v interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var success bool // 设置成功，返回true 。设置失败，返回false
_, err = horm.NewQuery("redis_student").SetNX("test_struct", data).Exec(ctx, &success)
```
- `GET`
  函数用法：
```go
// Get 获取指定 key 的值。如果 key 不存在，返回 nil 。可用 IsNil(err) 判断是否key不存在，如果key储存的值不是字符串类型，返回一个错误。
// param: key string
Get(key string)
```
使用示例：
```go
var bval bool
isNil, err = horm.NewQuery("redis_student").Get("test_bool").Exec(ctx, &bval)
	
if horm.IsError(err) {
	return err
}

if horm.IsNil(err) { // err = nil returned
	fmt.Println("key not set")
}

var ival int
isNil, err = horm.NewQuery("redis_student").Get("test_int").Exec(ctx, &ival)

var fval float64
isNil, err = horm.NewQuery("redis_student").Get("test_float").Exec(ctx, &fval)

var strval string
isNil, err = horm.NewQuery("redis_student").Get("test_string").Exec(ctx, &strval)

var structVal Student
isNil, err = horm.NewQuery("redis_student").Get("test_struct").Exec(ctx, &structVal)
```
- `GETSET`
  函数用法：
```go
// GetSet 设置给定 key 的值。如果 key 已经存储其他值， GetSet 就覆写旧值，并返回原来的值。如果原来未设置值，则返回报错 nil returned
// param: key string
// param: v interface{} 任意类型数据
GetSet(key string, v interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var lastSet Student
isNil, err = horm.NewQuery("redis_student").GetSet("test_struct", data).Exec(ctx, &lastSet)

if horm.IsError(err) {
	return err
}

if horm.IsNil(err) { // err = nil returned
	fmt.Println("key not set")
}
```
- `INCR`
  函数用法：
```go
// Incr 将 key 中储存的数字值增一。 如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCR 操作。 如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
Incr(key string)
```
使用示例：
```go
var ret int
isNil, err = horm.NewQuery("redis_student").Incr("test_int").Exec(ctx, &ret)
```
- `DECR`
  函数用法：
```go
// Decr 将 key 中储存的数字值减一。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 DECR 操作。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
Decr(key string)
```
使用示例：
```go
var ret int
isNil, err = horm.NewQuery("redis_student").Decr("test_int").Exec(ctx, &ret)
```
- `INCRBY`
  函数用法：
```go
// IncrBy 将 key 中储存的数字加上指定的增量值。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCRBY 命令。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。
// param: key string
// param: n string 自增数量
IncrBy(key string, n int)
```
使用示例：
```go
var ret int
isNil, err = horm.NewQuery("redis_student").IncrBy("test_int", 15).Exec(ctx, &ret)
```
- `MSET`
  函数用法：
```go
// MSet 批量设置一个或多个 key-value 对
// param: values map[string]interface{} // value will marshal
// 注意，本接口 Prefix 一定要在 MSet 之前设置，这里所有的 key 都会被加上 Prefix
MSet(values map[string]interface{}) 
```
使用示例：
```go
mdata := map[string]interface{}{
	"a": 19,
	"b": 21,
	"c": 25,
}

isNil, err = horm.NewQuery("redis_student").Prefix("user_age_").MSet(mdata).Exec(ctx)
```
结果：
```bash
127.0.0.1:6379> get user_age_a
"19"
127.0.0.1:6379> get user_age_b
"21"
127.0.0.1:6379> get user_age_c
"25"
```

- `MGET`
  函数用法：
```go
// MGet 返回多个 key 的 value
// param: keys string
MGet(keys []string)
```
使用示例：
```go
var result []int
isNil, err = horm.NewQuery("redis_student").Prefix("user_age_").MGet("a", "b", "c").Exec(ctx, &result)
```

- `SETBIT`
  函数用法：
```go
// SetBit 设置或清除指定偏移量上的位
// param: key string
// param: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)
// param: value bool true:设置为1,false：设置为0
SetBit(key string, offset uint32, value bool)
```

- `GETBIT`
  函数用法：
```go
// GetBit 获取指定偏移量上的位
// param: key string
// param: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)
GetBit(key string, offset uint32)
```

- `BITCOUNT`
  函数用法：
```go
// BitCount 计算给定字符串中，被设置为 1 的比特位的数量
// param: key string
// param: start int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推
// param: end int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推
BitCount(key string, start, end int)
```

## 哈希
- `HSET`
  函数用法：
```go
// HSet 为哈希表中的字段赋值 。
// param: key string
// param: field interface{} 其中field建议为字符串,可以为整数，浮点数
// param: v interface{} 任意类型数据
// param: args ...interface{} 多条数据，按照filed,value 的格式，其中field建议为字符串,可以为整数，浮点数
HSet(key string, field, v interface{}, args ...interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

isNil, err = horm.NewQuery("redis_student").HSet("student_2024061211", "age", 39).Exec(ctx)
```
- `HSETNX`
  函数用法：
```go
// HSetNx 为哈希表中不存在的的字段赋值 。
// param: key string
// param: field string
// param: value interface{}
// Query返回: bool 设置成功，返回true 。 如果给定字段已经存在且没有操作被执行，返回 false。
HSetNx(key string, filed interface{}, value interface{})
```
使用示例：
```go
var success bool // 设置成功，返回true 。 如果给定字段已经存在且没有操作被执行，返回 false。
isNil, err = horm.NewQuery("redis_student").HSetNx("student_age", 13324, 22).Exec(ctx, &success)
```
- `HMSET`
  函数用法：
```go
// HmSet 把map数据设置到哈希表中，此命令会覆盖哈希表中已存在的字段。如果哈希表不存在，会创建一个空哈希表，并执行 HMSET 操作。
// param: key string
// param: v map[string]interface{} 任意类型map数据
HmSet(key string, v map[string]interface{})
```
使用示例1：
```go
data := map[string]interface{}{
		Identify: 2024061211,
		Gender:   2,
		Age:      39,
		Name:     "metcalfe",
		Score:    93.8,
		Image:    []byte("IMAGE.PCG"),
		Article:  "contribution to leading the public into the era of hyper-connectivity",
		ExamTime: "15:30:00",
}

isNil, err = horm.NewQuery("redis_student").Prefix("student_").HmSet("13324", data).Exec(ctx)
```
结果：
```bash
127.0.0.1:6379> hget student_2024061211 userid
"13324"
127.0.0.1:6379> hget student_2024061211 class_id
"1"
127.0.0.1:6379> hget student_2024061211 sex
"male"
127.0.0.1:6379> hget student_2024061211 age
"23"
127.0.0.1:6379> hget student_2024061211 name
"smallhow"
```

使用示例2：
```go
data := Student{
  Identify: 2024061211,
  Gender:   1,
  Age:      78,
  Name:     "Alen Joy",
  Score:    99.9,
  Image:    []byte("IMAGE.PCG"),
  Article:  "UNIX operating system and C programming language",
  ExamTime: "16:30:00",
  Birthday: types.Time(birthday),
}

isNil, err = horm.NewQuery("redis_student").Prefix("student_").HmSet("19827", data).Exec(ctx)
```
结果：
```bash
127.0.0.1:6379> hget student_19827 userid
"19827"
127.0.0.1:6379> hget student_19827 class_id
"2"
127.0.0.1:6379> hget student_19827 sex
"male"
127.0.0.1:6379> hget student_19827 age
"23"
127.0.0.1:6379> hget student_19827 name
"jerry"
```
- `HMGET`
  函数用法：
```go
// HmGet 返回哈希表中，一个或多个给定字段的值。
// param: key string
// param: fields string 需要返回的域
HmGet(key string, fields []string)
```
使用示例：
```go
fields := []string{"userid", "class_id", "sex", "age", "name"}

var result Student //可以通过一个 struct 指针接收结果
isNil, err = horm.NewQuery("redis_student").HmGet("student_19827", fields).Exec(ctx, &result)

var result map[string]interface{} //也可以通过一个 map 接收结果
isNil, err = horm.NewQuery("redis_student").HmGet("student_19827", fields).Exec(ctx, &result)
```
- `HGET`
  函数用法：
```go
// HGet 数据从redis hget 出来之后反序列化并赋值给 v
// param: key string
// param: field string
func (s *Statement) HGet(key string, field interface{})
```
使用示例：
HSet 的类型是什么，HGet 也用相同类型的指针去接收结果。
```go
isNil,err = horm.NewQuery("redis_student").HGet("test_hset", "xxxxxx").Exec(ctx, &bval)
if horm.IsError(err) {
	return err
}

if horm.IsNil(err) { // err = nil returned
	fmt.Println("key not set")
}

isNil, err = horm.NewQuery("redis_student").HGet("test_hset", "aa").Exec(ctx, &ival)
isNil, err = horm.NewQuery("redis_student").HGet("test_hset", "bb").Exec(ctx, &fval)
isNil, err = horm.NewQuery("redis_student").HGet("test_hset", "cc").Exec(ctx, &strval)
isNil, err = horm.NewQuery("redis_student").HGet("test_hset", "ee").Exec(ctx, &structVal)
```
- `HGETALL`
  函数用法：
```go
// HGetAll 返回哈希表中，所有的字段和值。
// param: key string
func (s *Statement) HGetAll(key string)
```
使用示例：
```go
var result Student //可以通过一个 struct 指针接收结果
isNil, err = horm.NewQuery("redis_student").HGETALL("student_19827", fields).Exec(ctx, &result)

var result map[string]interface{} //也可以通过一个 map 接收结果
isNil, err = horm.NewQuery("redis_student").HGETALL("student_19827", fields).Exec(ctx, &result)
```
- `HKEYS`
  函数用法：
```go
// HKeys 获取哈希表中的所有域（field）。
// param: key string
Hkeys(key string)
```
使用示例：
```go
var result []string
isNil, err = horm.NewQuery("redis_student").Hkeys("student_19827").Exec(ctx, &result)
```

- `HINCRBY`
  函数用法：
```go
// HIncrBy 为哈希表中的字段值加上指定增量值。
// param: key string
// param: field string
// param: n string 自增数量
HIncrBy(key string, field string, v int)
```
使用示例：
```go
var newAge int
isNil, err = horm.NewQuery("redis_student").HIncrBy("student_19827", "age", 5).Exec(ctx, &newAge)
```
- `HDEL`
  函数用法：
```go
// HDel 删除哈希表 key 中的一个或多个指定字段，不存在的字段将被忽略。
// param: keyfield interface{}，删除指定key的field数据，这里输入的第一参数为key，其他为多个field，至少得有一个field
HDel(key string, field ...interface{})
```
使用示例：
```go
var delNum int
isNil, err = horm.NewQuery("redis_student").HDel("student", 19827, 23312, 98322).Exec(ctx, &delNum)
```
- `HEXISTS`
  函数用法：
```go
// HExists 查看哈希表的指定字段是否存在。
// param: key string
// param: field string
HExists(key string, field string)
```
使用示例：
```go
// isNil=true 则不存在，isNil=false 则存在
isNil, err = horm.NewQuery("redis_student").HExists("student", 19827).Exec(ctx)
```
- `HLEN`
  函数用法：
```go
// HLen 获取哈希表中字段的数量。
// param: key string
HLen(key string)
```
使用示例：
```go
var num int
isNil, err = horm.NewQuery("redis_student").HLen("student").Exec(ctx, &num)
```
- `HSTRLEN`
  函数用法：
```go
// HStrLen 获取哈希表某个字段长度。
// param: key string
// param: field string
HStrLen(key string, field interface{})
```
使用示例：
```go
var l int
isNil, err = horm.NewQuery("redis_student").HStrLen("student_19827", "name").Exec(ctx, &l)
```
- `HINCRBYFLOAT`
  函数用法：
```go
// HIncrByFloat 为哈希表中的字段值加上指定增量浮点数。
// param: key string
// param: field string
// param: v float64 自增数量
HIncrByFloat(key string, field string, v float64)
```
使用示例：
```go
var score float64
isNil, err = horm.NewQuery("redis_student").HIncrByFloat("student_19827", "score", 0.7).Exec(ctx, &score)
```
- `HVALS`
  函数用法：
```go
// HVals 返回所有的 value
// param: key string
HVals(key string)
```
使用示例：
```go
var ages []int
isNil, err = horm.NewQuery("redis_student").HVals("student_age").Exec(ctx, &ages)
```
## 列表
- `LPUSH`
  函数用法：
```go
// LPush 将一个或多个值插入到列表头部。 如果 key 不存在，一个空列表会被创建并执行 LPUSH 操作。 当 key 存在但不是列表类型时，返回一个错误。
// param: key string
// param: v interface{} 任意类型数据
LPush(key string, v ...interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var n int // 执行 LPUSH 命令后，列表的长度。
isNil, err = horm.NewQuery("redis_student").LPush("student_list", data).Exec(ctx, &n)
```
- `RPUSH`
  函数用法：
```go
// RPush 将一个或多个值插入到列表的尾部(最右边)。如果列表不存在，一个空列表会被创建并执行 RPUSH 操作。 当列表存在但不是列表类型时，返回一个错误。
// param: key string
// param: v interface{} 任意类型数据
RPush(key string, v ...interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var n int // 执行 RPUSH 命令后，列表的长度。
isNil, err = horm.NewQuery("redis_student").RPush("student_list", data).Exec(ctx, &n)
```
- `LPOP`
  函数用法：
```go
// LPop 移除并返回列表的第一个元素。
// param: key string
LPop(key string)
```
使用示例：
```go
var result Student //用于接收返回的任意类型指针
isNil, err = horm.NewQuery("redis_student").LPop("student_list").Exec(ctx, &result)
```
- `RPOP`
  函数用法：
```go
// RPop 移除列表的最后一个元素，返回值为移除的元素。
// param: key string
RPop(key string)
```
使用示例：
```go
var result Student //用于接收返回的任意类型指针
isNil, err = horm.NewQuery("redis_student").RPop("student_list").Exec(ctx, &result)
```

- `LLEN`
  函数用法：
```go
// LLen 返回列表的长度。 如果列表 key 不存在，则 key 被解释为一个空列表，返回 0 。 如果 key 不是列表类型，返回一个错误。
// param: key string
LLen(key string)
```
使用示例：
```go
var n int
isNil, err = horm.NewQuery("redis_student").LLen("student_list").Exec(ctx, &n)
```

## 集合
- `SADD`
  函数用法：
```go
// SAdd 将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。
// param: key string
// param: v ...interface{} 任意类型的多条数据，但是务必确保各条数据的类型保持一致，
SAdd(key string, v ...interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var n int //被添加到集合中的新元素的数量，不包括被忽略的元素。
isNil, err = horm.NewQuery("redis_student").SAdd("student_set", data).Exec(ctx, &n)
```
- `SMEMBERS`
  函数用法：
```go
// SMembers 返回集合中的所有的成员。 不存在的集合 key 被视为空集合。
// param: key string
SMembers(key string)
```
使用示例：
```go
var result []*Student
isNil, err = horm.NewQuery("redis_student").SMembers("student_set").Exec(ctx, &result)
```
- `SREM`
  函数用法：
```go
// SRem 移除集合中的一个或多个成员元素，不存在的成员元素会被忽略
// param: key string
// param: v ...interface{} 任意类型的多条数据
SRem(key string, members ...interface{})
```
使用示例：
```go
var n int // 被成功移除的元素的数量，不包括被忽略的元素。
isNil, err = horm.NewQuery("redis_student").SRem("student_set", data).Exec(ctx, &n)
```
- `SCARD`
  函数用法：
```go
// SCard 返回集合中元素的数量。
// param: key string
SCard(key string)
```
使用示例：
```go
var n int // 被成功移除的元素的数量，不包括被忽略的元素。
isNil, err = horm.NewQuery("redis_student").SCard("student_set").Exec(ctx, &n)
```
- `SISMEMBER`
  函数用法：
```go
// SIsMember 判断成员元素是否是集合的成员。
// param: key string
// param: member interface{} 要检索的任意类型数据
SIsMember(key string, member interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var isMember bool
isNil, err = horm.NewQuery("redis_student").SIsMember("student_set", data).Exec(ctx, &isMember)
```
- `SRANDMEMBER`
  函数用法：
```go
// SRandMember 返回集合中的count个随机元素。
// param: key string
// param: count int 随机返回元素个数。
// 如果 count 为正数，且小于集合基数，那么命令返回一个包含 count 个元素的数组，数组中的元素各不相同。
// 如果 count 大于等于集合基数，那么返回整个集合。
// 如果 count 为负数，那么命令返回一个数组，数组中的元素可能会重复出现多次，而数组的长度为 count 的绝对值。
SRandMember(key string, count int)
```
使用示例：
```go
var result []*Student
isNil, err = horm.NewQuery("redis_student").SRandMember("student_set", 2).Exec(ctx, &result)
```
- `SPOP`
  函数用法：
```go
// SPop 移除集合中的指定 key 的一个或多个随机成员，移除后会返回移除的成员。
// param: key string
// param: int count
SPop(key string, count int)
```
使用示例：
```go
var result []*Student
isNil, err = horm.NewQuery("redis_student").SPop("student_set", 2).Exec(ctx, &result)
```
- `SMOVE`
  函数用法：
```go
// SMove 将指定成员 member 元素从 source 集合移动到 destination 集合。
// param: source string
// param: destination string
// param: member interface{} 要移动的成员，任意类型
SMove(source, destination string, member interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

isNil, err = horm.NewQuery("redis_student").SMOVE("student_set", "student_set2", &data).Exec(ctx)
```
## 有序集
- `ZADD`
  函数用法：
```go
// ZAdd redis.ZAdd
// 将成员元素及其分数值加入到有序集当中。如果某个成员已经是有序集的成员，那么更新这个成员的分数值，并通过重新插入这个成员元素，
// 来保证该成员在正确的位置上。分数值可以是整数值或双精度浮点数。
// param: key string
// param: args ...interface{} 添加更多成员，需要按照  member, score, member, score 依次排列
// 注意：⚠️ 与 redis 命令不一样，需要按照  member, score, member, score, 格式传入
ZAdd(key string, args ...interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data1 := Student{
  Identify: 2024061211,
  Gender:   2,
  Age:      39,
  Name:     "metcalfe",
  Score:    93.8,
  Image:    []byte("IMAGE.PCG"),
  Article:  "contribution to leading the public into the era of hyper-connectivity",
  ExamTime: "15:30:00",
  Birthday: types.Time(birthday),
}

data2 := Student{
  Identify: 2024061211,
  Gender:   1,
  Age:      78,
  Name:     "Alen Joy",
  Score:    99.9,
  Image:    []byte("IMAGE.PCG"),
  Article:  "UNIX operating system and C programming language",
  ExamTime: "16:30:00",
  Birthday: types.Time(birthday),
}

var n int // 被成功添加的新成员的数量，不包括那些被更新的、已经存在的成员。
isNil, err = horm.NewQuery("redis_student").ZAdd("student_zset", &data1, 97, &data2, 89).Exec(ctx, &n)
```
- `ZREM`
  函数用法：
```go
// ZRem 移除有序集中的一个或多个成员，不存在的成员将被忽略。
// param: key string
// param: members ...interface{} 任意类型的多条数据
ZRem(key string, members ...interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var n int // 被成功移除的成员的数量，不包括被忽略的成员。
isNil, err = horm.NewQuery("redis_student").ZRem("student_zset", &data).Exec(ctx, &n)
```
- `ZREMRANGEBYSCORE`
  函数用法：
```go
// ZRemRangeByScore 移除有序集中，指定分数（score）区间内的所有成员。
// param: key string
// param: interface{} min max 分数区间，类型为整数或者浮点数
ZRemRangeByScore(key string, min, max interface{})
```
使用示例：
```go
var n int // 被移除成员的数量。
isNil, err = horm.NewQuery("redis_student").ZRemRangeByScore("student_zset", 90, 100).Exec(ctx, &n)
```
- `ZREMRANGEBYRANK`
  函数用法：
```go
// ZRemRangeByRank 移除有序集中，指定排名(rank)区间内的所有成员。
// param: key string
// param: start stop int 排名区间
ZRemRangeByRank(key string, start, stop int)
```
使用示例：
```go
var n int // 被移除成员的数量。
isNil, err = horm.NewQuery("redis_student").ZRemRangeByRank("student_zset", 0, 2).Exec(ctx, &n)
```
- `ZCARD`
  函数用法：
```go
// ZCard 返回有序集成员个数
// param: key string
ZCard(key string)
```
使用示例：
```go
var n int
isNil, err = horm.NewQuery("redis_student").ZCard("student_zset").Exec(ctx, &n)
```
- `ZSCORE`
  函数用法：
```go
// ZScore 返回有序集中，成员的分数值。
// param: key string
// param: member interface{} 成员
ZScore(key string, member interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var score float64
isNil, err = horm.NewQuery("redis_student").ZScore("student_zset", &data).Exec(ctx, &score)
```
- `ZRANK`
  函数用法：
```go
// ZRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。
// param: key string
// param: member interface{} 成员，任意类型
ZRank(key string, member interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var rank int
isNil, err = horm.NewQuery("redis_student").ZRank("student_zset", &data).Exec(ctx, &rank)
```
- `ZREVRANK`
  函数用法：
```go
// ZRevRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从大到小)顺序排列。
// param: key string
// param: member interface{} 成员，任意类型
ZRevRank(key string, member interface{})
```
使用示例：
```go
birthday, _ := time.Parse("2006-01-02", "1976-08-27")
data := Student{
	Identify: 2024061211,
	Gender:   2,
	Age:      39,
	Name:     "metcalfe",
	Score:    93.8,
	Image:    []byte("IMAGE.PCG"),
	Article:  "contribution to leading the public into the era of hyper-connectivity",
	ExamTime: "15:30:00",
	Birthday: types.Time(birthday),
}

var rank int
isNil, err = horm.NewQuery("redis_student").ZRevRank("student_zset", &data).Exec(ctx, &rank)
```
- `ZCOUNT`
  函数用法：
```go
// ZCount 计算有序集合中指定分数区间的成员数量
// param: key string
// param: min interface{}
// param: max interface{}
ZCount(key string, min, max interface{})
```
使用示例：
```go
var n int
isNil, err = horm.NewQuery("redis_student").ZCount("student_zset", 70, 80).Exec(ctx, &n)
```
- `ZPOPMIN`
  函数用法：
```go
// ZPopMin 移除并弹出有序集合中分值最小的 count 个元素
// redis v5.0.0+
// param: key string
// param: count ...int64 不设置count参数时，弹出一个元素
ZPopMin(key string, count ...int64)
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZPopMin("student_zset", 2).Exec(ctx, &result, scores)
```
- `ZPOPMAX`
  函数用法：
```go
// ZPopMax 移除并弹出有序集合中分值最大的的 count 个元素
// redis v5.0.0+
// param: key string
// param: count ...int64 不设置count参数时，弹出一个元素
ZPopMax(key string, count ...int64)
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZPopMax("student_zset", 2).Exec(ctx, &result, &scores)
```
- `ZINCRBY`
  函数用法：
```go
// ZIncrBy 对有序集合中指定成员的分数加上增量 increment，可以通过传递一个负数值 increment ，
// 让分数减去相应的值，比如 ZINCRBY key -5 member ，
// 就是让 member 的 score 值减去 5 。当 key 不存在，或分数不是 key 的成员时，
// ZINCRBY key increment member 等同于 ZADD key
// increment member 。当 key 不是有序集类型时，返回一个错误。分数值可以是整数值或双精度浮点数。
// param: key string
// param: member interface{} 任意类型数据
// param: incr interface{} 增量值，可以为整数或双精度浮点
ZIncrBy(key string, member, incr interface{})
```
使用示例：
```go
var newScore float64 // member 成员的新分数值。
isNil, err = horm.NewQuery("redis_student").ZIncrBy("student_zset", &data, 9.5).Exec(ctx, &newScore)
```
- `ZRANGE`
  函数用法：
```go
// ZRange 返回有序集中，指定区间内的成员。其中成员的位置按分数值递增(从小到大)来排序。
// param: key string
// param: int start, stop 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
// param: withScore 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
ZRange(key string, start, stop int, withScore ...bool)
```
使用示例1：
```go
var result []*Student
isNil, err = horm.NewQuery("redis_student").ZRange("student_zset", 0, 9).Exec(ctx, &result)
```

使用示例2：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZRange("student_zset", 0, 9, true).Exec(ctx, &result, &scores)
```
返回结果：

- `ZRANGEBYSCORE`
  函数用法：
```go
// ZRangeByScore 根据分数返回有序集中的成员
// param: key string
// param: int min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
// param: withScores 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// param: LIMIT信息，包含 offset count
ZRangeByScore(key string, min, max interface{}, withScores bool, limit ...int64)
```
使用示例：
```go
var result []*Student // 80分~90分之间
isNil, err = horm.NewQuery("redis_student").ZRangeByScore("student_zset", 80, 90).Exec(ctx, &result)
```

使用示例2：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZRangeByScore("student_zset", 70, 100, true).Exec(ctx, &result, &scores)
```

使用示例3：
```go
var result []*Student // 80分~90分之间，OFFSET=0，LIMIT=10
isNil, err = horm.NewQuery("redis_student").ZRangeByScore("student_zset", 80, 90, false, 0, 10).Exec(ctx, &result)
```

使用示例4：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZRangeByScore("student_zset", 70, 100, true, 0, 10).Exec(ctx, &result, &scores)
```
返回结果：


- `ZREVRANGE`
  函数用法：
```go
// ZRevRange 返回有序集中指定区间的成员，其中成员的位置按分数值递减(从大到小)来排列。
// param: key string
// param: start, stop 排名区间，以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
// param: withScore 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
ZRevRange(key string, start, stop int, withScore ...bool)
```
使用示例：
```go
var result []*Student
isNil, err = horm.NewQuery("redis_student").ZRevRange("student_zset", 0, 9).Exec(ctx, &result)
```

使用示例2：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZRevRange("student_zset", 0, 9, true).Exec(ctx, &result, &scores)
```
返回结果：

- `ZREVRANGEBYSCORE`
  函数用法：
```go
// ZRevRangeByScore 返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。
// param: key string
// param: min , max interface{} 分数区间，类型为整数或双精度浮点数，但是 -inf +inf 表示负正无穷大
// param: withScore 是否返回有序集的分数， true - 返回，false - 不返回，默认不返回，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]
// param: LIMIT 信息，包含 offset count
ZRevRangeByScore(key string, min, max interface{}, withScore bool, limit ...int64)
```
使用示例1：
```go
var result []*Student // 80分~90分之间
isNil, err = horm.NewQuery("redis_student").ZRevRangeByScore("student_zset", 80, 90).Exec(ctx, &result)
```

使用示例2：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZRevRangeByScore("student_zset", 70, 100, true).Exec(ctx, &result, &scores)
```

使用示例3：
```go
var result []*Student // 80分~90分之间，OFFSET=0，LIMIT=10
isNil, err = horm.NewQuery("redis_student").ZRevRangeByScore("student_zset", 80, 90, false, 0, 10).Exec(ctx, &result)
```

使用示例4：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
isNil, err = horm.NewQuery("redis_student").ZRevRangeByScore("student_zset", 70, 100, true, 0, 10).Exec(ctx, &result, &scores)
```

返回结果：

# 事务
# 插件编排
## 插件配置
## 插件版本
### 版本回滚
# 结果编排
## 列表页详情页
### 并行执行单元结果 JOIN
### 字段提取
### 编排后再排序
## 数据可视化
### 柱状图
### 饼状图
### 折线图
### 同比环比
### 单页面多图