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
建表语句：
```sql
CREATE TABLE `student` (
    `id` int NOT NULL AUTO_INCREMENT,
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
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学生表';

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

## golang 结构体：
```go
type Student struct {
  Id        int       `orm:"id,int,onuniqueid" json:"id"`
  Identify  int64     `orm:"identify,bigint" json:"identify"`                     //身份证件ID
  Gender    int8      `orm:"gender,tinyint,omitinsertempty" json:"gender"`        //1-male 2-female
  Age       uint      `orm:"age,int unsigned,omitreplaceempty" json:"age"`        //年龄
  Name      string    `orm:"name,varchar,omitupdateempty" json:"name"`            //名称
  Score     float64   `orm:"score,double,omitempty" json:"score"`                 //分数
  Image     []byte    `orm:"image,blob,omitempty" json:"image"`                   //image
  Article   string    `orm:"article,text,omitempty" json:"article"`               //publish article
  ExamTime  string    `orm:"exam_time,time,omitempty" json:"exam_time"`           //考试时间
  Birthday  time.Time `orm:"birthday,date" json:"birthday"`                       //出生日期
  CreatedAt time.Time `orm:"created_at,timestamp,oncreatetime" json:"created_at"` //创建时间
  UpdatedAt time.Time `orm:"updated_at,datetime,onupdatetime" json:"updated_at"`  //修改时间
}
```

```go
type StudentCourse struct {
  Id       uint64 `orm:"id,int,omitempty" json:"id"`
  Identify int64  `orm:"identify,bigint" json:"identify"` //学生编号
  Course   string `orm:"course,varchar" json:"course"`    //课程
  Hours    int    `orm:"hours,int" json:"hours"`          //课时数
}

type CourseInfo struct {
  Course  string `orm:"course,varchar" json:"course"`   //课程
  Teacher string `orm:"teacher,varchar" json:"teacher"` //课程老师
  Time    string `orm:"time,time" json:"time"`          //上课时间
}

type TeacherInfo struct {
  Teacher string `orm:"teacher,varchar" json:"teacher"` //课程老师
  Age     int    `orm:"age,int" json:"age"`             //年龄
}

type ScoreRankReward struct {
  Rank   int    `orm:"rank,int" json:"rank"`         //排名
  Reward string `orm:"reward,varchar" json:"reward"` //奖励
}
```

## 结构体标签
支持通过 golang 结构体标签来描述数据库表字段，标签以 orm 开头，第一个 field 为表字段名，第二个 field 为表类型，其他为属性，例如：
```golang
//示例结构体
type Student struct {
  Id        int       `orm:"id,int,onuniqueid" json:"id"`                         //onuniqueid 新增数据时候，如果字段为空值，而且类型为 uint64，则自动生成唯一 ID，记得务必在 orm.yaml 配置里面为每台机器设置不同的 machine_id，每个实例不一样。否则可能会有冲突
  Identify  int64     `orm:"identify,bigint" json:"identify"`                     
  Gender    int8      `orm:"gender,tinyint,omitinsertempty" json:"gender"`        //omitinsertempty 插入忽略零值
  Age       uint      `orm:"age,int unsigned,omitreplaceempty" json:"age"`        //omitreplaceempty 替换忽略零值
  Name      string    `orm:"name,varchar,omitupdateempty" json:"name"`            //omitupdateempty 更新忽略零值
  Score     float64   `orm:"score,double,omitempty" json:"score"`           
  Image     []byte    `orm:"image,blob,omitempty" json:"image"`                   
  Article   string    `orm:"article,text,omitempty" json:"article"`               //omitempty 忽略零值，= omitinsertempty + omitreplaceempty + omitupdateempty 表示插入、替换、更新都忽略零值，如 Auto Increment 需要
  ExamTime  string    `orm:"exam_time,time,omitempty" json:"exam_time"`         
  Birthday  time.Time `orm:"birthday,date" json:"birthday"`                      
  CreatedAt time.Time `orm:"created_at,timestamp,oncreatetime" json:"created_at"` //oncreatetime 插入时自动初始化为当前时间
  UpdatedAt time.Time `orm:"updated_at,datetime,onupdatetime" json:"updated_at"`  //onupdatetime 自动修改为当前时间
}
```

```go
//hrom 会将上面结构体的 orm 标签解析后，每个 field 结果存入下面的结构体中，并缓存到内存
// FieldSpec body 标签解析结果
type FieldSpec struct {
  Tag              string // tag
  Name             string // 字段名
  I                int    // 位置
  Index            []int
  Column           string // 对应数据库字段名
  Type             string // 表字段类型，bool、string、int、int8、int16、int32、int64、uint、uint8、uint16、uint32、uint64、float、float64、blob(bytes)、enum、json、date、datetime
  OmitEmpty        bool   // 忽略零值
  OmitInsertEmpty  bool   // INSERT 时忽略零值
  OmitReplaceEmpty bool   // REPLACE 时忽略零值
  OmitUpdateEmpty  bool   // UPDATE 时忽略零值
  OnCreateTime     bool   // INSERT/REPLACE 时初始化为当前时间，具体格式根据 Type 决定，如果是数字类型包括 int、int32、int64 等，则是时间戳，否则就是 time.Time 类型
  OnUpdateTime     bool   // 数据变更时修改为当前时间，具体格式根据 Type 决定，这里我推荐数据库自带的时间戳更新功能。
  OnUniqueID       bool   // 新增数据时候，如果字段为空值，而且类型为 uint64，则自动生成唯一 ID，记得务必在 orm.yaml 配置里面为每台机器设置不同的 machine_id，否则生成的ID可能会有冲突
}

```

# horm 客户端
为了访问数据统一接入服务，我们需要创建 Client 来与服务端建立连接，horm 提供了2种方式来指定 Query 语句使用的客户端。
一、为Query语句指定特定的客户端。二、配置全局客户端，在未指定特定客户端的情况下，所有 Query 都采用该全局客户端。

## 新建 horm client 连接
### 指定客户端
我们首先通过 horm.NewClient 创建一个客户端，该函数的第一个参数是允许传入一个 caller name，
他将读取配置文件 orm.yaml 里面的 server.caller.name 对应的数据统一接入服务 workspace_id、 encryption、token、target、appid、secret等信息，
然后用 WithClient 为 Query 指定该 Client。
我们可能会访问多个数据统一接入服务的不同数据（每个服务都有唯一的 workspace id）。

```go
package main
import (
	...
	"github.com/horm-database/go-horm/horm"
)

...

func Test(ctx context.Context) {
	//创建 db server 连接
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
package main
import (
	...
   "github.com/horm-database/go-horm/horm"
)

func Test(ctx context.Context) {
  cli := horm.NewClient("ws_test.app2.server2.service2",
    horm.WithAppID(10099),
    horm.WithSecret("S499721834"))

  var result = make([]*Student, 0)
  _, err := horm.NewQuery("student").FindAll().WithClient(cli).Exec(ctx, &result)
  
  ...
}
```

### 配置全局连接
配置全局变量之后，如果 Query 没有用 WithClient 指定客户端的话，就使用全局客户端
```go
package main
import (
	...
	"github.com/horm-database/go-horm/horm"
)

...

//init 配置全局Client
func init() {
  horm.SetGlobalClient("ws_test.app1.server1.service1")
}

func Test(ctx context.Context) {
  var result = make([]*Student, 0)
  _, err := horm.NewQuery("student").FindAll().Exec(ctx, &result)
  
  ...
}

```

# 执行模式
## 单执行单元
执行单条语句，`isNil`, `error` 直接通过 Exec 函数返回，当查询结果为空时，isNil=true，可以将 result 指针传入 Exec 第二个参数，接收返回结果。
```go
package main
import (
	...
	"github.com/horm-database/go-horm/horm"
)

...

func Test(ctx context.Context) {
  var result Student
  var where = horm.Where{"id": 13}
  isNil, err := horm.NewQuery("student").Find(where).Exec(ctx, &result)
	
  ...
}
```

有时候，可能需要返回多个结果，例如 redis 的 ZRangeByScore：
```go

birthday, _ := time.Parse("2006-01-02", "1987-08-27")

data := Student{
  Identify: 2024070733,
  Gender:   1,
  Age:      19,
  Name:     "smallhow",
  Score:    92.1,
  Image:    []byte("IMAGE.PCG"),
  Article:  "For pioneering work in the fields of cryptography and complex theory",
  ExamTime: "15:30:00",
  Birthday: birthday,
}

//horm 会对结构体参数自动编解码
_, err := horm.NewQuery("redis_student").
	ZAdd("student_score_rank", data, data.Score).Exec(ctx)

results := make([]*Student, 0)
scores := make([]float64, 0)
_, err = horm.NewQuery("redis_student").
	ZRangeByScore("student_score_rank", 70, 100, true).Exec(ctx, &results, &scores)

```

![image](https://github.com/horm-database/image/blob/master/%E5%8D%95%E6%89%A7%E8%A1%8C%E5%8D%95%E5%85%83-1.png)


## 并行执行多条语句
为了高效并发，我们可以用 `PExec` 函数将多个语句一同上传到数据统一接入服务，由数据统一接入服务并发执行，并返回结果，在 Query 语句里面，可以通过 `Next` 新建一个并发语句，然后通过 `WithReceiver` 传入对应指针来接收每个执行语句返回的 isNil、error 和结果。

`注意：如果并行执行访问同一个数据时，为了区别，可以像下面一样在括号里面加别名：redis_student(zadd) 和 redis_student(range)`

```go
birthday, _ := time.Parse("2006-01-02", "1987-08-27")
data := Student{
    Identify: 430602198702221111,
    Gender:   1,
    Age:      19,
    Name:     "smallhow",
    Score:    92.1,
    Image:    []byte("IMAGE.PCG"),
    Article:  "Artificial Intelligence",
    ExamTime: "15:30:00",
    Birthday: birthday,
}

var isNil bool
var zaddErr, rangeErr error
results := make([]*Student, 0)
scores := make([]float64, 0)

//下面的 query name 有加别名
err := horm.NewQuery("redis_student(zadd)").
            ZAdd("student_score_rank", &data, data.Score).WithReceiver(nil, &zaddErr).
            Next("redis_student(range)").
            ZRangeByScore("student_score_rank", 70, 100, true).WithReceiver(&isNil, &rangeErr, &results, &scores).
            PExec(ctx)
```

![image](https://github.com/horm-database/image/blob/master/%E5%8D%95%E6%89%A7%E8%A1%8C%E5%8D%95%E5%85%83-1.png)

## 复合执行
```json
[
  {
    "name": "student",
    "op": "find_all",
    "size": 100,
    "sub": [
      {
        "name": "student_course(sc)",
        "op": "find_all",
        "where": {
          "@identify": "identify"
        },
        "size": 100
      },
      {
        "name": "redis_student(rank)",
        "op": "zrank",
        "key": "student_score_rank",
        "args": [
          "@{identify}"
        ]
      }
    ]
  },
  {
    "name": "course_info",
    "op": "find_all",
    "where": {
      "@course": "/student/sc.course"
    },
    "size": 100
  }
]
```
## 返回结果
### IsNil
当数据源为 mysql、clickhouse、es 等数据库时，如果 Find或者 FindAll 查询的数据为空时，返回参数 isNil=true， 否则，返回参数为 false，
而当数据源为 redis 时，只有 redis 返回 redigo: nil returned 错误时，才会使得 isNil = true，其他时候都是 isNil = false，即便如下 ZRangeByScore 去查询一个不存在的有序集时。
```go
// 查询单条 mysql 数据
var result Student
where := horm.Where{"name": "noexist"}
isNil, err := horm.NewQuery("student").Find(where).Exec(ctx, &result) // isNil = true

// 查询多条 mysql 数据
var results []*Student
where = horm.Where{"name": "noexists"}
isNil, err = horm.NewQuery("student").FindAll(where).Exec(ctx, &results) // isNil = true

// redis 中 GET 缓存
var stu Student
isNil, err = horm.NewQuery("redis_student").Get("noexists").Exec(ctx, &stu) // isNil = true

// redis ZRangeByScore
rets := make([]*Student, 0)
scores := make([]float64, 0)
isNil, err = horm.NewQuery("redis_student"). // isNil = false ， rets 和 scores 为空数组
ZRangeByScore("noexists", 70, 100, true).Exec(ctx, &rets, &scores)
```

### IsError
`horm.IsError(err)` 可以判断是否执行失败，如果是 nil returned 错误，不是真正的错误，而是数据为空，或者 redis key 不存在。

```go
func Test(ctx context.Context) {
	var result = make([]*Students, 0)
	err := horm.NewQuery("student").FindAll().Exec(ctx, &result)
	
	if horm.IsError(err) { //可以判断是否执行失败，如果是 nil returned 错误，不是真正的错误，而是数据为空，或者 redis key 不存在。
		fmt.Printf("find student error: %v", err)
		return
	}

	if horm.IsNil(err) { // err = nil returned，所有空数据都返回这个错误。
		fmt.Printf("not fine student")
	}
	
	
	...
}
```

### IsAllSuccess
这个函数仅用于 Elastic 批量插入新数据时候，返回 `[]*horm.EsResult`，可以用 IsAllSuccess 去判断数据是否全部插入成功，当只有部分成功的时候，我们可以遍历返回结果，`status` 为错误码，当 `status!=0` 则该条记录插入失败，`reason`为失败原因，这样，我们可以针对失败的记录做特殊处理，或者重试。
```go
func Test(ctx context.Context) {
	datas := []*Student{
		{
			ClassId: 1,
			Sex:     "male",
			Age:     22,
			Name:    "smallhow",
		},
		{
			ClassId: 2,
			Sex:     "female",
			Age:     19,
			Name:    "jerry",
		},
	}

	result := make([]*horm.EsResult, 0)
	err := horm.NewQuery("es_student").InsertStructs(&datas).Exec(ctx, &result)

	if horm.IsError(err) {
		fmt.Printf("batch insert student error: %v", err)
		return
	}

	if horm.IsAllSuccess(result) {
		fmt.Printf("batch insert success")
		return
	}
	...
}
```

返回结果：
```json
[
    {
        "_id":"wU3spIEBL4QnOSO-F-tV",
        "version":1,
        "rows_affected":1,
        "status":0,
        "reason":""
    },
    {
        "_id":"wk3spIEBL4QnOSO-F-tV",
        "version":1,
        "rows_affected":1,
        "status":0,
        "reason":""
    }
]
```

# 数据名称（执行单元名）
每个执行语句都需要指定一个数据名称，也叫执行单元名，表明我需要访问哪个数据，如下的 `horm.NewQuery("student(senior)") 中的 student` ，我们在 DB Proxy 可以通过 `数据名称` 找到对应的表/索引/redis 配置信息、及其数据库信息，然后执行 sql 语句、elastic 请求或 redis 请求。
```go
package main
import (
	...
	"github.com/horm-database/go-horm/horm"
)

...

func Test(ctx context.Context) {
	var result = make([]*Students, 0)
	err := horm.NewQuery("student").FindAll().Exec(ctx, &result)
	...
}
```

## 别名
在并发执行、复合执行场景下，同一层级的多条语句如果访问同一张表，为了结果的正常，我们最好在括号里加上别名，如下代码的`horm.NewQuery("skynet(zadd)")` 和 `Next("skynet(range_by_score)")` ，我们都是访问 BDB 库 skynet。
```go
func Test(ctx context.Context) {
	student := Student{
		Sex:    "male",
		Age:    19,
		Name:   "smallhowcao",
		Status: 1,
	}
	
	var err1, err2 error
	ret := make([]*Student, 0)
	scores := make([]float64, 0)
	
	//下面操作有加别名
	err := horm.NewQuery("skynet(zadd)").ZAdd("student_score_rank", student, 65).WithReceiver(&err1).
		Next("skynet(range_by_score)").ZRangeByScoreWithScore("student_score_rank", 70, 100).WithReceiver(&err2, &ret, &scores).
		PExec(ctx)
}
```

## 分片、分表、分库

在 DB Proxy 服务端，可以配置 4 种分表策略。
* 0 - 无分表，直接返回 DB Proxy 配置 table_name 。以及所属的数据库信息。
* 1 - 取自客户端送的 shard 字段，通过 `Shard` 函数指定分表。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)
	err := horm.NewQuery("student").Shard("student_61").FindAll().Exec(ctx, &result)
}
```
如果是 es 获取多个索引的数据，我们可以用逗号分隔，例如：
```go
func Test(ctx context.Context) {
	result := make([]*EsArticleLog, 0)
	err := horm.NewQuery("es_article_log").Shard("es_article_log_202205,es_article_log_202206").
			FindAll().Exec(ctx, &result)
}
```
* 2 - 字段取模，table_name + "_" + shard % shard_num
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)
	err := horm.NewQuery("student").Shard(userID).FindAll().Exec(ctx, &result)
}
```

* 3 - shard 函数，遇到上面都无法满足的特殊分表、分库需求，我们可以配置 shard func 模式，然后在 DB Proxy 服务端代码里面嵌入分表函数：


# 查询语句
## 指定查询列
通过 `Column` 指定要查询的列。
示例1：
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)
	err := horm.NewQuery("student").Column("userid", "class_id", "created_at").FindAll().Exec(ctx, &result)
}
```

示例2：
```go
func Test(ctx context.Context) {
	result := make([]map[string]interface{}, 0)
	err := horm.NewQuery("student").
		Column("userid as id, max(age) as max_age").
		GroupBy("userid").FindAll().Exec(ctx, &result)

	if horm.IsNil(err) {
		fmt.Println("not fine student")
	}
}
```

## 主键搜索（elastic 则是 _id ）
- 示例1 mysql 主键查询：
```go
func Test(ctx context.Context) {
	result := Student{}

	err := horm.NewQuery("student").Eq("userid", 3099).FindOne().Exec(ctx, &result)
	// SELECT * FROM `student` WHERE `userid` = 3099
	...
}
```
- 示例2 mysql 主键查询：
```go
func Test(ctx context.Context) {
	result := []*Student{}

	err := horm.NewQuery("student").Eq("userid", []int{3099, 6348, 9713}).FindAll().Exec(ctx, &result)
	// SELECT * FROM `student` WHERE `userid` IN (3099, 6348, 9713) LIMIT 100
	...
}
```

- 示例3 elastic 按照 _id 批量插入 ：
```go
func Test(ctx context.Context) {
	datas := []*Student{}

	datas = append(datas, &Student{
		Userid: 2338,
		Sex:    "male",
		Name:   "smallhowcao",
	})

	datas = append(datas, &Student{
		Userid: 1650,
		Sex:    "male",
		Name:   "smallhowcao",
	})

	ids := []interface{}{2338, 1650}

	result := make([]*horm.EsResult, 0)
	err := horm.NewQuery("es_student").Eq("_id", ids).InsertStructs(datas).Exec(ctx, &result)
	...
}
```

返回结果：

- 示例 4 elastic 按照 _id 查询
```go
func Test(ctx context.Context) {
	ids := []interface{}{2338, 1650}

	result := []*Student{}
	err := horm.NewQuery("es_student").Eq("_id", ids).FindAll().Exec(ctx, &result)
	...
}
```

## where 查询条件
### 操作符
```
const ( // 操作符
	OPEqual          = "="  // 等于
	OPBetween        = "()" // 在某个区间
	OPNotBetween     = "><" // 不在某个区间
	OPGt             = ">"  // 大于
	OPGte            = ">=" // 大于等于
	OPLt             = "<"  // 小于
	OPLte            = "<=" // 小于等于
	OPNot            = "!"  // 去反
	OPLike           = "~"  // like语句，（或 es 的部分匹配）
	OPNotLike        = "!~" // not like 语句，（或 es 的部分匹配排除）
	OPMatchPhrase    = "?"  // es 短语匹配 match_phrase
	OPNotMatchPhrase = "!?" // es 短语匹配排除 must_not match_phrase
	OPMatch          = "*"  // es 全文搜索 match 语句
	OPNotMatch       = "!*" // es 全文搜索排除 must_not match
)
```

### 基础用法
由于篇幅问题，下面所有用法都是用 mysql 举例，如果对应库类型为 elastic 则 DB Proxy 会生成对应的 es 请求。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	where := horm.Where{}
	where["age"] = 29                    //`age` = 29
	where["age >"] = 29                  //`age` > 29
	where["age >="] = 29                 //`age` >= 29
	where["age !"] = 29                  //`age` != 29
	where["age ()"] = []int{20, 29}      //`age` BETWEEN 20 AND 29
	where["age ><"] = []int{35, 40}      // NOT ( `age` BETWEEN 35 AND 40)
	where["score"] = []int{60, 61, 62}   //`score` IN (60, 61, 62)
	where["score !"] = []int{70, 71, 72} //`score` NOT IN (70, 71, 72)
	where["name"] = nil                  //`name` IS NULL
	where["name !"] = nil                //`name` IS NOT NULL
	where["name ! #注释：排除smallhow"] = "smallhow"                //`name` != 'smallhow'

	err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)

	if horm.IsError(err) { // 判断是否执行失败，如果是 nil returned 错误，不是真正的错误，而是空数据。
		fmt.Printf("find student error: %v", err)
		return
	}

	if horm.IsNil(err) { // err = nil returned，所有返回空数据都报这个错误。
		fmt.Printf("not fine student")
	}
}
```

FindOne 查询单条记录：
```go
func Test(ctx context.Context) {
	result := Student{}
	
	where := horm.Where{"userid": 399883}
	err := horm.NewQuery("student").FindOne(where).Exec(ctx, &result)

	if horm.IsError(err) { // 判断是否执行失败，如果是 nil returned 错误，不是真正的错误，而是空数据。
		fmt.Printf("find student error: %v", err)
		return
	}

	if horm.IsNil(err) { // err = nil returned。
		fmt.Printf("not fine student userid=399883")
	}
}
```

### 组合查询
针对快速构建 where 语句方式，我们也支持通过 "AND" 或者 "OR"、"NOT" 来组合更复杂的语句。

- 示例1：
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{
		"age":     36,
		"score >": 97,
		"OR": horm.Where{
			"id ()": []int{10, 25},
			"sex":   "male",
		},
	}
	
	err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)
	
	//不设置 limit 默认取100条
	// SELECT * FROM `student` WHERE `age` = 36 AND `score` > 97 AND ((`id` BETWEEN 10 AND 25) OR `sex` = 'male')  LIMIT 100
	...
}
```
上述语句如果转化为 elastic search 的 query 条件语句，则为（es 占用篇幅较大，后面都以 MySQL 为例）：
```json
{
    "bool":{
        "filter":[
            {
                "terms":{
                    "age":[
                        36
                    ]
                }
            },
            {
                "range":{
                    "score":{
                        "from":97,
                        "include_lower":false,
                        "include_upper":true,
                        "to":null
                    }
                }
            },
            {
                "bool":{
                    "should":[
                        {
                            "range":{
                                "id":{
                                    "from":10,
                                    "include_lower":true,
                                    "include_upper":true,
                                    "to":25
                                }
                            }
                        },
                        {
                            "terms":{
                                "sex":[
                                    "male"
                                ]
                            }
                        }
                    ]
                }
            }
        ]
    }
}
```
- 示例2：
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	//注意：由于mysql使用map参数，所以在下面的情况下，第一个 OR 会被覆盖。
	where := horm.Where{
		"OR": horm.Where{
			"id >": 3,
			"sex":  "male",
		},
		"OR": horm.Where{
			"uid !":     3,
			"height >=": 170,
		},
	}

	// [X] SELECT * FROM `student` WHERE (`uid`!=3 OR `height`>=170)
	
	where := horm.Where{
		"OR #注释1": horm.Where{
			"id >": 3,
			"sex":  "male",
		},
		"OR #注释2": horm.Where{
			"uid !":     3,
			"height >=": 170,
		},
	}

	// [√] SELECT * FROM `student` WHERE (`id`>3 OR `sex`='male') AND (`uid`!= 3 OR `height`>=170) LIMIT 100

	err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)
	...
}
```
- 示例3
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	where := horm.Where{
		"NOT": horm.Where{
			"id >": 3,
			"sex":  "male",
		},
	}

	// SELECT * FROM `student` WHERE NOT (`id` > 3 AND `sex` = 'male') LIMIT 100

	err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)
	...
}
```
### 模糊匹配
注意：在 elastic 中 LIKE 操作符用法有些不同，详细可以看下一个章节。
- 示例1：
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ~"] = "%ide%"                              //`name` LIKE '%ide%'
	where["addtime ~"] = []string{"2019-08%", "2020-01%"}  //( `addtime` LIKE '2019-08%' OR `addtime` LIKE '2020-01%')
	where["name !~"] = "%ide%"                             //`name` NOT LIKE '%ide%'
	where["addtime !~"] = []string{"2019-08%", "2020-01%"} //( `addtime` NOT LIKE '2019-08%' AND `addtime` NOT LIKE '2020-01%')  ## 注意他和 LIKE 的连接词不一样，NOT LIKE 是 AND，而 LIKE 是 OR

	err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)
	...
}
```
- 示例2：
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ~"] = "Londo_"   // London, Londox, Londos...
	where["name ~"] = "[BCR]at"  // Bat, Cat, Rat
	where["name ~"] = "[!BCR]at" // Eat, Fat, Hat...
	
	err := horm.NewQuery("student").FindAll(where).Exec(ctx, &result)
	...
}
```

### 部分匹配（prefix、wildcard、regexp）（elastic 特有）
对于 elastic ，horm 把 `~` 操作符用于表示部分匹配。部分匹配分3中类型，prefix（默认）、wildcard、regexp

#### prefix 前缀查询（默认）
类似 mysql 的 like 'cao%'，以 cao 为前缀的所有内容。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{"name ~": "cao"}  // caohao, caocao, caoxueqin...
	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)

	if horm.IsError(err) { // 判断是否执行失败，如果是 nil returned 错误，不是真正的错误，而是空数据。
		fmt.Printf("find student error: %v", err)
		return
	}

	if horm.IsNil(err) { // err = nil returned，所有返回空数据都报这个错误。
		fmt.Printf("not fine student")
	}
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "prefix":{
                "name":"cao"
            }
        }
    }
}
```

#### wildcard 通配符查询
它使用标准的 shell 通配符查询： `?` 匹配任意字符， `*` 匹配 0 或多个字符。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["name ~(type=wildcard)"] = "ca*h?o" // cao hao, ca li hoo, ca huo...

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "wildcard":{
                "name":{
                    "value":"ca*h?o"
                }
            }
        }
    }
}
```

#### regexp 正则表达式查询
这个是正则查询，如下示例的正则表达式要求词必须以 W 开头，紧跟 0 至 9 之间的任何一个数字，然后接一或多个其他字符。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title ~(type=regexp)"] = "W[0-9].+"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```

生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "regexp":{
                "title":{
                    "value":"W[0-9].+"
                }
            }
        }
    }
}
```

#### NOT 部分匹配排除
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title !~"] = "smallhow" // 不以smallhow开头

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```

生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must_not":{
            "prefix":{
                "title":"smallhow"
            }
        }
    }
}
```

### 短语匹配 match_phrase（elastic 特有）
`match_phrase`  查询首先将查询字符串解析成一个`词项列表`，然后对这些词项进行搜索，但只保留那些包含`全部 搜索词项`，且`位置`与搜索词项相同的文档。在 horm ，我们用 `?` 操作符表示短语匹配。 `!?` 表示短语匹配排除。

```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title ?"] = "smallhow"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "match_phrase":{
                "title":{
                    "query":"smallhow"
                }
            }
        }
    }
}
```

#### 灵活度 slop
精确短语匹配 或许是过于严格了。也许我们想要包含 “quick brown fox” 的文档也能够匹配 “quick fox” ，如下：
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title ?(slop=1)"] = "quick fox"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "match_phrase":{
                "title":{
                    "query":"quick fox",
                    "slop":1
                }
            }
        }
    }
}
```

#### 提升权重
我们可以通过指定  `boost`  来控制任何查询语句的相对的权重，  `boost`  的默认值为  `1`  ，大于  `1`  会提升一个语句的相对权重。如下，title 中包含"smallhow"的话，权重更高。那么他可能会拥有更高的 `_score`评分。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title ?(boost=3)"] = "smallhow"
	where["abstract ?(boost=2)"] = "smallhow"
	where["content ?(boost=1)"] = "smallhow"

	err := horm.NewQuery("es_student").FindAll(where).Order("_score", true).Exec(ctx, &result)
	...
}
```
生成的 es 查询语句：
```json
{
    "query":{
        "bool":{
            "must":[
                {
                    "match_phrase":{
                        "abstract":{
                            "boost":2,
                            "query":"smallhow"
                        }
                    }
                },
                {
                    "match_phrase":{
                        "content":{
                            "boost":1,
                            "query":"smallhow"
                        }
                    }
                },
                {
                    "match_phrase":{
                        "title":{
                            "boost":3,
                            "query":"smallhow"
                        }
                    }
                }
            ]
        }
    },
    "from":0,
    "size":100,
    "sort":[
        {
            "_score":{
                "order":"desc"
            }
        }
    ]
}
```

#### 多个属性
一个where 条件可以同时拥有多个数据，通过 `&` 来分隔，如下 title 有 slop 和 boost 两个条件属性。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title ?(slop=2,boost=1)"] = "quick fox"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "match_phrase":{
                "title":{
                    "query":"quick fox",
                    "slop":2,
                    "boost":1
                }
            }
        }
    }
}
```

#### NOT 短语匹配排除
操作符 `!?` 表示短语匹配排除
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title !?"] = "smallhow" //标题中不包含smallhow的记录

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must_not":{
            "match_phrase":{
                "title":{
                    "query":"smallhow"
                }
            }
        }
    }
}
```
### 全文搜索 match（elastic 特有）
对于 elastic ，horm 把 `*` 操作符用于表示全文检索，在全文字段中搜索到最相关的文档。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title *"] = "国产芯片"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "match":{
                "title":{
                    "query":"国产芯片"
                }
            }
        }
    }
}
```

#### 提高精度 operator
上述例子，中文分词会将`国产芯片`分为`国产`、`芯片`， 用 任意查询词项匹配文档可能会导致结果中出现不相关的长尾。这是种散弹式搜索。可能我们只想搜索包含`所有词项`的文档，也就是说，不去匹配  `国产 OR 芯片` ，而通过匹配  `国产 AND 芯片`  找到所有文档。

`match`  查询还可以接受  `operator`  操作符作为输入参数，默认情况下该操作符是  `or`  。我们可以将它修改成  `and`  让所有指定词项都必须匹配：
 ```json
 func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title *(operator=and)"] = "国产芯片"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
 ```
生成的 elastic query 条件语句 ：
 ```json
 {
    "bool":{
        "must":{
            "match":{
                "title":{
                    "query":"国产芯片",
                    "operator":"and"
                }
            }
        }
    }
}
```
#### 控制精度 minimum_should_match
在 所有 与 任意 间二选一有点过于非黑即白。如果用户给定 5 个查询词项，想查找只包含其中 4 个的文档，该如何处理？

在全文搜索的大多数应用场景下，我们既想包含那些可能相关的文档，同时又排除那些不太相关的。换句话说，我们想要处于中间某种结果。

`match`  查询支持  `minimum_should_match`  最小匹配参数，这让我们可以指定必须匹配的词项数用来表示一个文档是否相关。我们可以将其设置为某个具体数字，更常用的做法是将其设置为一个百分数，因为我们无法控制用户搜索时输入的单词数量，如下，我们设置最小匹配参数为 60%，即只需要命中至少 3个词，则匹配文档。

```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title *(minimum_should_match=60%)"] = "smallhow is stockholder of jingdong"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must":{
            "match":{
                "title":{
                    "query":"smallhow is stockholder of jingdong",
                    "minimum_should_match":"60%"
                }
            }
        }
    }
}
```
#### 评分计算
`bool`  查询会为每个文档计算相关度评分  `_score`  ，再将所有匹配的  `must`  和  `should`  语句的分数  `_score`  求和，最后除以  `must`  和  `should`  语句的总数。
`must_not`  语句不会影响评分；它的作用只是将不相关的文档排除。

#### 提升权重
提升权重与 `match_phrase` 的用法是一样的，也是通过指定  `boost`  来控制任何查询语句的相对的权重，  `boost`  的默认值为  `1`  ，大于  `1`  会提升一个语句的相对权重。
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title *(boost=3)"] = "smallhow"
	where["abstract *(boost=2)"] = "smallhow"
	where["content *(boost=1)"] = "smallhow"

	err := horm.NewQuery("es_student").FindAll(where).Order("_score", true).Exec(ctx, &result)
	...
}
```

#### NOT 全文搜索排除
```go
func Test(ctx context.Context) {
	result := make([]*Student, 0)

	var where = horm.Where{}
	where["title !*"] = "国产芯片"

	err := horm.NewQuery("es_student").FindAll(where).Exec(ctx, &result)
	...
}
```
生成的 elastic query 条件语句 ：
```json
{
    "bool":{
        "must_not":{
            "match":{
                "title":{
                    "query":"国产芯片"
                }
            }
        }
    }
}
```

## 分组、聚合（暂未支持 elastic 的聚合）
### GROUP BY
通过 GroupBy("column" ...) 加上 GROUP BY 语句，支持一个或多个参数。
```go
func Test(ctx context.Context) {
	result := make([]map[string]interface{}, 0)

	var where = horm.Where{"age>": 20}

	err := horm.NewQuery("student").Column("sex,age,count(1) as cnt").
		FindAll(where).GroupBy("sex", "age").Exec(ctx, &result)

	// SELECT sex,age,count(1) as cnt FROM `student` WHERE `age` > 20  GROUP BY `sex`,`age` LIMIT 100
	...
}
```
查询结果：
```json
[
    {
        "sex":"male",
        "age":23,
        "cnt":3
    },
    {
        "sex":"male",
        "age":24,
        "cnt":1
    },
    {
        "sex":"male",
        "age":22,
        "cnt":3
    },
    {
        "sex":"male",
        "age":25,
        "cnt":1
    }
]
```

### HAVING
Having 函数参数与 where 条件一样 ，两者生成 SQL 语法一致。

```go
func Test(ctx context.Context) {
	result := make([]map[string]interface{}, 0)

	var where = horm.Where{"age>": 20}
	var having = horm.Where{"cnt>": 2}

	err := horm.NewQuery("student").Column("sex,age,count(1) as cnt").
		FindAll(where).GroupBy("sex", "age").Having(having).Exec(ctx, &result)
	// SELECT sex,age,count(1) as cnt FROM `student` WHERE `age` > 20 GROUP BY `sex`,`age` HAVING `cnt` > 2  LIMIT 100
	...
}
```

查询结果：<br>

## 排序与分页
### ORDER 排序
通过 `Order` 函数指定排序。
```go
horm.NewQuery("student").FindAll().Order("age") //ORDER BY age ASC

horm.NewQuery("student").FindAll().Order("age", true) //ORDER BY age DESC

horm.NewQuery("student").FindAll().Order("age DESC", "score ASC") //ORDER BY age DESC, score ASC
```

elastic 按照相关性评分排序
```go
horm.NewQuery("es_student").FindAll().Order("_score", true) //ORDER BY _score DESC
```

### LIMIT、OFFSET
通过 `Limit` 函数去指定 limit 、offset 参数。
```go
horm.NewQuery("student").FindAll().Limit(20) // LIMIT 20

horm.NewQuery("student").FindAll().Limit(20, 50) // LIMIT 20 OFFSET 50
```

### 分页 PAGE
为了方便分页数据的返回，提供了分页函数 `Page`，第一个参数为`请求页数`，从1开始，第二个参数为`每页大小`。

分页返回，必须得重新定义返回结构，加入分页返回信息，将列表数据放入 data 域，如下：
```go
func Test(ctx context.Context) {
	type pageStudents struct {
		Total     uint64     `json:"total"`      //总数
		TotalPage uint32     `json:"total_page"` //总页数
		Page      uint32     `json:"page"`       //当前分页
		PageSize  uint32     `json:"page_size"`  //每页大小
		Data      []*Student `json:"data"`       //数据
	}

	result := pageStudents{}

	err := horm.NewQuery("student").FindAll().Page(2, 5).Exec(ctx, &result)
	...
}
```


## 返回结果高亮（elastic 特有）
- 示例1，用高亮结果替换原内容，多个高亮结果用 `</br>` 分隔：
```go
func Test(ctx context.Context) {
	var where = horm.Where{}
	where["title *"] = "elastic"

	highLight := horm.HighLight{
		Fields:   []string{"title"},
		PreTags:  "<highlight>",
		PostTags: "</highlight>",
	}

	result := make([]*Student, 0)
	err := horm.NewQuery("es_student").FindAll(where).HighLight(&highLight).Exec(ctx, &result)
	...
}
```
查询结果：
```json
[
    {
        "userid":1552,
        "class_id":3,
        "sex":"male",
        "age":28,
        "name":"smallhowcao",
        "title":"在他找工作的过程中，为了给妻子构建一个食谱的搜索引擎，他开始构建一个早期版本的 <highlight>elastic</highlight> 。</br>直接基于Lucene工作会比较困难，所以Shay开始抽象 <highlight>elastic</highlight> 代码以便lava程序员可以在应用中添加搜索功能。他发布了他的第一个开源项目，叫做“Compass”。</br>然后他决定重写Compass库使其成为一个独立的服务叫做 <highlight>elastic</highlight> search。</br>第一个公开版本出现在2010年2月，在那之后 <highlight>elastic</highlight> 已经成为Github上最受欢迎的项目之一，代码贡献者超过300人。</br>一家主营 <highlight>elastic</highlight> 的公司就此成立，他们一边提供商业支持一边开发新功能，不过 <highlight>elastic</highlight> 将永远开源且对所有人可用。\nShay的妻子依旧等待着她的食谱搜索……</br>"
    },
    {
        "userid":658,
        "class_id":1,
        "sex":"male",
        "age":28,
        "name":"smallhowcao",
        "title":"以搜索引擎闻名世界的开源软件提供商 <highlight>elastic</highlight> 。\n2012年成立，总部位于美国的山景城。\n2018年10月上市，目前市值近50亿美金。</br><highlight>elastic</highlight> 公司致力于结构化和非结构化数据的分布式实时全文搜索及分析，典型应用场景 包括日志管理、分析、系统指标分析、安全分析、企业搜索、网站搜索、应用搜索、应用性能管理APM等。</br><highlight>elastic</highlight> 公司产品包括享誉业界的 <highlight>elastic</highlight> stack(ES，ELK Stack)、具备多种高级特性的商业扩展插件、云服务 <highlight>elastic</highlight> cloud 等。</br>"
    }
]
```

- 示例2：高亮结果单独返回：
  我们会将字段以 `key_highlight` 的方式追加到 map 里面，如下 `title` 的高亮结果存于 `title_highlight` 中。
```go
func Test(ctx context.Context) {
	var where = horm.Where{}
	where["title *"] = "elastic"

	highLight := horm.HighLight{
		Fields:          []string{"title"},
		PreTags:         "<highlight>",
		PostTags:        "</highlight>",
		ReturnHighLight: horm.ReturnHighLightAlone,
	}
	
	type highLightStudent struct {
		Student
		HighLight []string `json:"title_highlight"`
	}
	highLightResult := make([]*highLightStudent, 0)

	err := horm.NewQuery("es_student").FindAll(where).HighLight(&highLight).Exec(ctx, &highLightResult)
	...
}
```

查询结果：
```json
[
    {
        "userid":1552,
        "class_id":3,
        "sex":"male",
        "age":28,
        "name":"smallhowcao",
        "title":"多年前，一个叫做Shay Banon的刚结婚不久的失业开发者，由于妻子要去伦敦学习厨师，他便跟着也去了。在他找工作的过程中，为了给妻子构建一个食谱的搜索引擎，他开始构建一个早期版本的 elastic 。\n直接基于Lucene工作会比较困难，所以Shay开始抽象 elastic 代码以便lava程序员可以在应用中添加搜索功能。他发布了他的第一个开源项目，叫做“Compass”。\n后来Shay找到一份工作，这份工作处在高性能和内存数据网格的分布式环境中，因此高性能的、实时的、分布式的搜索引擎也是理所当然需要的。然后他决定重写Compass库使其成为一个独立的服务叫做 elastic search。\n第一个公开版本出现在2010年2月，在那之后 elastic 已经成为Github上最受欢迎的项目之一，代码贡献者超过300人。一家主营 elastic 的公司就此成立，他们一边提供商业支持一边开发新功能，不过 elastic 将永远开源且对所有人可用。\nShay的妻子依旧等待着她的食谱搜索……",
        "title_highlight":[
            "在他找工作的过程中，为了给妻子构建一个食谱的搜索引擎，他开始构建一个早期版本的 <highlight>elastic</highlight> 。",
            "直接基于Lucene工作会比较困难，所以Shay开始抽象 <highlight>elastic</highlight> 代码以便lava程序员可以在应用中添加搜索功能。他发布了他的第一个开源项目，叫做“Compass”。",
            "然后他决定重写Compass库使其成为一个独立的服务叫做 <highlight>elastic</highlight> search。",
            "第一个公开版本出现在2010年2月，在那之后 <highlight>elastic</highlight> 已经成为Github上最受欢迎的项目之一，代码贡献者超过300人。",
            "一家主营 <highlight>elastic</highlight> 的公司就此成立，他们一边提供商业支持一边开发新功能，不过 <highlight>elastic</highlight> 将永远开源且对所有人可用。\nShay的妻子依旧等待着她的食谱搜索……"
        ]
    },
    {
        "userid":658,
        "class_id":1,
        "sex":"male",
        "age":28,
        "name":"smallhowcao",
        "title":"以搜索引擎闻名世界的开源软件提供商 elastic 。\n2012年成立，总部位于美国的山景城。\n2018年10月上市，目前市值近50亿美金。\n公司员工1000+人分布在全世界80多个国家，以最快的效率服务于我们的客户。\n100000+社区参与者，3亿5千万下载使用量，5000+商业客户持续订阅。\nelastic 公司致力于结构化和非结构化数据的分布式实时全文搜索及分析，典型应用场景 包括日志管理、分析、系统指标分析、安全分析、企业搜索、网站搜索、应用搜索、应用性能管理APM等。\nelastic 公司产品包括享誉业界的 elastic stack(ES，ELK Stack)、具备多种高级特性的商业扩展插件、云服务 elastic cloud 等。",
        "title_highlight":[
            "以搜索引擎闻名世界的开源软件提供商 <highlight>elastic</highlight> 。\n2012年成立，总部位于美国的山景城。\n2018年10月上市，目前市值近50亿美金。",
            "<highlight>elastic</highlight> 公司致力于结构化和非结构化数据的分布式实时全文搜索及分析，典型应用场景 包括日志管理、分析、系统指标分析、安全分析、企业搜索、网站搜索、应用搜索、应用性能管理APM等。",
            "<highlight>elastic</highlight> 公司产品包括享誉业界的 <highlight>elastic</highlight> stack(ES，ELK Stack)、具备多种高级特性的商业扩展插件、云服务 <highlight>elastic</highlight> cloud 等。"
        ]
    }
]
```

# 数据维护
## INSERT 语句
### 插入 map 数据
我们可以通过 `InsertMap` 传入 map 数据，插入单条数据。
- 示例 1，MySQL 插入新数据，返回`horm.AffectedInfo`，如果不关心返回，可以不传 result：
```go
func Test(ctx context.Context) {
	data := horm.SetMap{
		"class_id": 1,
		"name":     "smallhowcao",
		"sex":      "male",
		"age":      33,
		"status":   1,
	}

	result := horm.AffectedInfo{}
	err := horm.NewQuery("student").InsertMap(data).Exec(ctx, &result)

	if horm.IsError(err) {
		fmt.Printf("insert student error: %v", err)
		return
	}
	...
}
```
- 示例2，Elastic 插入新数据，返回 `horm.EsResult`，如果不关心返回，可以不传 result：
```go
func Test(ctx context.Context) {
	data := horm.SetMap{
		"class_id": 1,
		"name":     "smallhowcao",
		"sex":      "male",
		"age":      33,
		"status":   1,
	}

	result := horm.EsResult{}
	err := horm.NewQuery("es_student").InsertMap(data).Exec(ctx, &result)
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
### 插入 struct 数据
`InsertStruct`函数用于插入单条数据。

如下代码， 我们在 struct 结构体里面加入了自己的标签 orm，用于将字段与数据库表字段对应上，如果没有该标签，字段将会是struct的field name，如果在标签第3个分割位加上加上 omitinsertempty 可以让插入数据的时候忽略该字段，例如 Auto Increment 的字段就很需要这个，关于orm标签详情请看[结构体标签](#结构体标签)
- 示例 1，MySQL 插入新数据，返回 `horm.AffectedInfo` ，如果不关心返回，可以不传 result：
```go
type Userinfo struct {
	//omitempty = omitinsertempty + omitreplaceempty + omitupdateempty 表示插入、替换、更新都忽略零值，如 Auto Increment 需要
	Id         int         `orm:"id,int,omitempty"`
	Status     bool        `orm:"status,int8,omitinsertempty"` //omitinsertempty 插入忽略零值
	Height     int         `orm:"height,int,omitreplaceempty"`    //omitreplaceempty 替换忽略零值
	Score      float64     `orm:"score,double,omitupdateempty"`   //omitupdateempty 更新忽略零值
	Name       string      `orm:"name,varchar"`
	Sex        string      `orm:"sex,varchar"`
	Work       string      `orm:"work,varchar"`
	Buyed      string      `orm:"buyed,varchar"`
	Age        int         `orm:"age,int"`
	Bid        int         `orm:"bid,int"`
	Addtime    time.Time   `orm:"addtime,timestamp,oncreatetime"`    //oncreatetime 插入时自动初始化为当前时间
	Updatetime time.Time   `orm:"updatetime,timestamp,onupdatetime"` //onupdate 自动修改为当前时间
}

func Test(ctx context.Context) {
	data := Userinfo{
		Height: 170,
		Status: true,
		Name:   "smallhowcao",
		Sex:    "male",
		Age:    33,
		Bid:    1004,
	}

	result := horm.AffectedInfo{}
	err := horm.NewQuery("student").InsertStruct(&data).Exec(ctx, &result)
	...
}
```
返回结果：
```json
{
    "last_insert_id":1324,
    "rows_affected":1
}
```

- 示例2，Elastic 插入新数据，返回 `horm.EsResult`，如果不关心返回，可以不传 result：
```go
func Test(ctx context.Context) {
	data := Userinfo{
		Height: 170,
		Status: true,
		Name:   "smallhowcao",
		Sex:    "male",
		Age:    33,
		Bid:    1004,
	}

	result := horm.EsResult{}
	err := horm.NewQuery("es_student").InsertStruct(&data).Exec(ctx, &result)
	
	if horm.IsError(err) {
		fmt.Printf("insert student error: %v", err)
		return
	}
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
`InsertStructs`函数用于插入多条数据。
- 示例 1，MySQL 插入新数据，返回 `horm.AffectedInfo` ，如果不关心返回，可以不传 result：
```go
func Test(ctx context.Context) {
	datas := []*Student{
		{
			ClassId: 1,
			Sex:     "male",
			Age:     22,
			Name:    "smallhow",
		},
		{
			ClassId: 2,
			Sex:     "female",
			Age:     19,
			Name:    "jerry",
		},
	}

	result := horm.AffectedInfo{}
	err := horm.NewQuery("student").InsertStructs(&datas).Exec(ctx, &result)
	...
}
```
返回结果：
```json
{
    "last_insert_id":39923455,
    "rows_affected":2
}
```

- 示例2，Elastic 插入新数据，返回 `[]*horm.EsResult`，如果不关心返回，可以不传 result：
```go
func Test(ctx context.Context) {
	datas := []*Student{
		{
			ClassId: 1,
			Sex:     "male",
			Age:     22,
			Name:    "smallhow",
		},
		{
			ClassId: 2,
			Sex:     "female",
			Age:     19,
			Name:    "jerry",
		},
	}

	result := make([]*horm.EsResult, 0)
	err := horm.NewQuery("es_student").InsertStructs(&datas).Exec(ctx, &result)

	if horm.IsError(err) {
		fmt.Printf("batch insert student error: %v", err)
		return
	}

	if horm.IsAllSuccess(result) {
		fmt.Printf("batch insert success")
		return
	}
	...
}
```

返回结果：

可以通过 `horm.IsAllSuccess` 判断是否全部数据都插入成功，当只有部分成功的时候，我们可以遍历返回结果，`status` 为错误码，当 `status!=0` 则该条记录插入失败，`reason`为失败原因：
```json
[
    {
        "_id":"wU3spIEBL4QnOSO-F-tV",
        "version":1,
        "rows_affected":1,
        "status":0,
        "reason":""
    },
    {
        "_id":"wk3spIEBL4QnOSO-F-tV",
        "version":1,
        "rows_affected":1,
        "status":0,
        "reason":""
    }
]
```

- 示例3，Elastic 根据 `_id` 插入新数据，返回 `[]*horm.EsResult`，如果不关心返回，可以不传 result：
```go
func Test(ctx context.Context) {
	datas := []*Student{
		{
			Userid:  2233455,
			ClassId: 1,
			Sex:     "male",
			Age:     22,
			Name:    "smallhow",
		},
		{
			Userid:  2233456,
			ClassId: 2,
			Sex:     "female",
			Age:     19,
			Name:    "jerry",
		},
	}

	ids := []int{2233455, 2233456}

	result := make([]*horm.EsResult, 0)
	err := horm.NewQuery("es_student").InsertStructs(&datas).Eq("_id", ids).Exec(ctx, &result)
	
	if horm.IsError(err) {
		fmt.Printf("batch insert student error: %v", err)
		return
	}

	if horm.IsAllSuccess(result) {
		fmt.Printf("batch insert success")
		return
	}
	...
}
```

返回结果：

可以通过 `horm.IsAllSuccess` 判断是否全部数据都插入成功，当只有部分成功的时候，我们可以遍历返回结果，`status` 为错误码，当 `status!=0` 则该条记录插入失败，`reason`为失败原因：
```json
[
    {
        "_id":"2233455",
        "version":1,
        "rows_affected":1,
        "status":0,
        "reason":""
    },
    {
        "_id":"2233456",
        "version":1,
        "rows_affected":1,
        "status":0,
        "reason":""
    }
]
```
## REPLACE 语句
replace 和 insert 函数类似，只不过是把 sql 关键词 insert 替换为 replace，可以参考 insert 的写法。也是支持 `ReplaceMap`、`ReplaceStruct`、`ReplaceStructs` 三个函数。

`注意：elastic search 不支持 replace`

## UPDATE 语句
### 更新 map 数据
- 示例1（主键更新）：
```go
func Test(ctx context.Context) {
	data := horm.SetMap{
		"class_id": 2,
		"name":     "jerry",
		"sex":      "male",
		"age":      21,
	}

	result := horm.AffectedInfo{}
	err := horm.NewQuery("student").UpdateMap(data).Eq("userid", 9713).Exec(ctx, &result)
	...
}
```
返回结果：
```json
{
    "last_insert_id":0,
    "rows_affected":1
}
```
- 示例2（where 条件更新）：
```go
func Test(ctx context.Context) {
	data := horm.SetMap{
		"class_id": 2,
		"name":     "jerry",
		"sex":      "male",
		"age":      21,
	}

	where := horm.Where{"userid": 2233456}
	
	result := horm.AffectedInfo{}
	err := horm.NewQuery("student").UpdateMap(data, where).Exec(ctx, &result)
	...
}
```
- 示例3（elastic by query 更新）：
```go
func Test(ctx context.Context) {
	data := horm.SetMap{
		"class_id": 2,
		"name":     "jerry",
		"sex":      "male",
		"age":      21,
	}

	where := horm.Where{"userid": 2233456}
	
	result := horm.EsResult{}
	err := horm.NewQuery("es_student").UpdateMap(data, where).Exec(ctx, &result)
	...
}
```
生成的请求：
```json
{
    "query":{
        "bool":{
            "filter":{
                "terms":{
                    "age":[
                        19
                    ]
                }
            }
        }
    },
    "script":{
        "source":"ctx._source.age=21;ctx._source.class_id=2;ctx._source.name='jerry';ctx._source.sex='male'"
    }
}
```
返回结果：
```json
{
    "_id":"",
    "version":0,
    "rows_affected":1
}
```

- 示例3（elastic by _id 更新）：
```go
func Test(ctx context.Context) {
	data := horm.SetMap{
		"class_id": 2,
		"name":     "jerry",
		"sex":      "male",
		"age":      22,
	}

	result := horm.EsResult{}
	err := horm.NewQuery("es_student").UpdateMap(data).Eq("_id", 2233456).Exec(ctx, &result)
	...
}
```
生成的请求：
```json
/student/_update/2233456?refresh=false
{
    "doc":{
        "age":21,
        "class_id":2,
        "name":"jerry",
        "sex":"male"
    }
}
```

返回结果：
```json
{
    "_id":"2233456",
    "version":5,
    "rows_affected":1
}
```

### 更新 struct 数据
omitupdateempty、omitempty、onupdatetime 标签对本函数生效。
```go
func Test(ctx context.Context) {
	data := Student{
		ClassId: 1,
		Sex:     "male",
		Age:     31,
		Name:    "smallhow",
	}

	result := horm.AffectedInfo{}
	err := horm.NewQuery("student").UpdateStruct(data).Eq("userid", 6348).Exec(ctx, &result)
	...
}
```
其他示例与 更新 map 类似，就是将 UpdateStruct 替换为 UpdateMap 。

## DELETE 数据
delete 比较简单，就只需要加上 where 条件即可。
- 示例1（mysql）
```go
func Test(ctx context.Context) {
	where := horm.Where{"age": 33}

	result := horm.AffectedInfo{}
	err := horm.NewQuery("student").Delete(where).Exec(ctx, &result)
	...
}
```
返回结果：
```json
{
    "last_insert_id":0,
    "rows_affected":2
}
```

- 示例2（elastic by query）
```go
func Test(ctx context.Context) {
	where := horm.Where{"age": 22}

	result := horm.EsResult{}
	err := horm.NewQuery("es_student").Delete(where).Exec(ctx, &result)
	...
}
```
返回结果：
```json
{
    "rows_affected":5,
}
```

- 示例3（elastic by _id）
```go
func Test(ctx context.Context) {
	result := horm.EsResult{}
	err := horm.NewQuery("es_student").Delete().Eq("_id", "w001qIEBL4QnOSO-k-s5").Exec(ctx, &result)
	...
}
```
返回结果：
```json
{
    "_id":"w001qIEBL4QnOSO-k-s5",
    "version":2,
    "rows_affected":1,
    "status":0,
    "reason":""
}
```


## refresh 刷新（elastic 特有）
通过 `Refresh(true)` 函数可以使 elastic 在更新数据之后立即刷新，当然，这个会导致 elastic search 的压力增大。
```go
func Test(ctx context.Context) {
	data := horm.SetMap{
		"class_id": 2,
		"name":     "jerry",
		"sex":      "male",
		"age":      22,
	}

	result := horm.EsResult{}
	err := horm.NewQuery("es_student").UpdateMap(data).Eq("_id", 2233456).
		Refresh().        // 更新数据立即刷新
		Exec(ctx, &result)
	...
}
```

# redis 协议
## Prefix（强烈建议使用）
`Prefix` 可以为我们的 key 加上前缀，如下真正的 key 就是 `student_13324` ，`强烈建议所有 key 都加上前缀，便于服务端对具体业务数据的统计与数据看盘，能够更好的定位具体的业务量级、请求暴增等情况。`，

```go
func Test(ctx context.Context) {
	result := Student{}
	err := horm.NewQuery("skynet").Prefix("student_").Get("13324").Exec(ctx, &result)
	if horm.IsError(err) {
		fmt.Println("find student error: %v", err)
		return
	}

	if horm.IsNil(err) { // err = nil returned
		fmt.Println("student not exists")
	}
}
```

## 编码与解码
通过 `WithCoder` ，我们可以为 redis 自定义编解码器。如果未使用，则采用默认编解码器，
```go
// WithCoder 更换编解码器
func (s *Statement) WithCoder(coder RCodec) *Statement {
	s.Coder = coder
	return s
}

// GetCoder 获取编解码器
func (s *Statement) GetCoder() RCodec {
	if s.Coder != nil {
		return s.Coder
	}
	//返回默认编码器
	return RCodecJSON
}
```
如下示例，我们在 Set 的时候，会采用默认编解码器将 data 编码为 `json` 字符串，然后 set 到 redis，在 Get 函数，我们传入 `Student 结构体指针`，去接收解码后的返回结果。
```go
func Test(ctx context.Context) {
	data := Student{
		Userid:  13324,
		ClassId: 1,
		Sex:     "male",
		Age:     23,
		Name:    "smallhow",
		Status:  1,
	}

	err := horm.NewQuery("skynet").Prefix("student_").Set("13324", data).Exec(ctx)
	if horm.IsError(err) {
		fmt.Println("find student error: %v", err)
		return
	}

	result := Student{}
	err = horm.NewQuery("skynet").Prefix("student_").Get("13324").Exec(ctx, &result)
	if horm.IsError(err) {
		fmt.Println("find student error: %v", err)
		return
	}

	if horm.IsNil(err) { // err = nil returned
		fmt.Println("student not exists")
	}
}
```

在没有编解码的情况下，比如复合查询模式下，返回的就是原始的 redis 数据，他返回8种类型的结构：
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
下面示例为了便捷，我都省略 `Prefix`。
- `EXPIRE`
  函数用法：
```go
// Expire 设置 key 的过期时间，key 过期后将不再可用。单位以秒计。
// param: key string
// param: int ttl 到期时间，ttl秒
Expire(key string, ttl int)
```
使用示例：
```go
err := horm.NewQuery("skynet").Expire("student_13324", 3600).Exec(ctx)
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
err := horm.NewQuery("skynet").TTL("student_13324").Exec(ctx, &ttl)
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
var exists bool // 存在返回true，否则返回false
err := horm.NewQuery("skynet").Exists("student_13324").Exec(ctx, &exists)
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
err := horm.NewQuery("skynet").Del("student_13324").Exec(ctx, &affectNum)
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
data := Student{
	Userid:  13324,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

err := horm.NewQuery("skynet").Set("test_bool", true).Exec(ctx)
err = horm.NewQuery("skynet").Set("test_int", 78).Exec(ctx)
err = horm.NewQuery("skynet").Set("test_float", 63.2567).Exec(ctx)
err = horm.NewQuery("skynet").Set("test_string", "i am ok").Exec(ctx)
err = horm.NewQuery("skynet").Set("test_struct", data).Exec(ctx)
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
data := Student{
	Userid:  13324,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

err := horm.NewQuery("skynet").SetEX("test_bool", true, 3600).Exec(ctx)
err = horm.NewQuery("skynet").SetEX("test_int", 78, 3600).Exec(ctx)
err = horm.NewQuery("skynet").SetEX("test_float", 63.2567, 3600).Exec(ctx)
err = horm.NewQuery("skynet").SetEX("test_string", "i am ok", 86400).Exec(ctx)
err = horm.NewQuery("skynet").SetEX("test_struct", data, 86400).Exec(ctx)
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
data := Student{
	Userid:  13324,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var success bool // 设置成功，返回true 。设置失败，返回false
err := horm.NewQuery("skynet").SetNX("test_struct", data).Exec(ctx, &success)
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
err := horm.NewQuery("skynet").Get("test_bool").Exec(ctx, &bval)
	
if horm.IsError(err) {
	return err
}

if horm.IsNil(err) { // err = nil returned
	fmt.Println("key not set")
}

var ival int
err = horm.NewQuery("skynet").Get("test_int").Exec(ctx, &ival)

var fval float64
err = horm.NewQuery("skynet").Get("test_float").Exec(ctx, &fval)

var strval string
err = horm.NewQuery("skynet").Get("test_string").Exec(ctx, &strval)

var structVal Student
err = horm.NewQuery("skynet").Get("test_struct").Exec(ctx, &structVal)
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
data := Student{
	Userid:  13324,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var lastSet Student
err := horm.NewQuery("skynet").GetSet("test_struct", data).Exec(ctx, &lastSet)

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
err := horm.NewQuery("skynet").Incr("test_int").Exec(ctx, &ret)
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
err := horm.NewQuery("skynet").Decr("test_int").Exec(ctx, &ret)
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
err := horm.NewQuery("skynet").IncrBy("test_int", 15).Exec(ctx, &ret)
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

err := horm.NewQuery("skynet").Prefix("user_age_").MSet(mdata).Exec(ctx)
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
keys := []string{"a", "b", "c"}
err := horm.NewQuery("skynet").Prefix("user_age_").MGet(keys).Exec(ctx, &result)
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
data := Student{
	Userid:  13324,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

err := horm.NewQuery("skynet").HSet("student", "13324", data).Exec(ctx)
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
err := horm.NewQuery("skynet").HSetNx("student_age", 13324, 22).Exec(ctx, &success)
```
- `HMSET`
  函数用法：
```go
// HmSet 把map数据设置到哈希表中。
// param: key string
// param: v map[string]interface{} 任意类型map数据
HmSet(key string, v map[string]interface{})
```
使用示例：
```go
data := map[string]interface{}{
	"userid":   13324,
	"class_id": 1,
	"sex":      "male",
	"age":      23,
	"name":     "smallhow",
}

err := horm.NewQuery("skynet").Prefix("student_").HmSet("13324", data).Exec(ctx)
```
结果：
```bash
127.0.0.1:6379> hget student_13324 userid
"13324"
127.0.0.1:6379> hget student_13324 class_id
"1"
127.0.0.1:6379> hget student_13324 sex
"male"
127.0.0.1:6379> hget student_13324 age
"23"
127.0.0.1:6379> hget student_13324 name
"smallhow"
```
- `HmSetStruct`
  函数用法：
```go
// HmSetStruct 把struct结构体数据写入哈希表中，此命令会覆盖哈希表中已存在的字段。如果哈希表不存在，会创建一个空哈希表，并执行 HMSET 操作。
// param: key string
// param: v interface{} 结构体数据 map or slice slice is key,value...结构
HmSetStruct(key string, v interface{})
```
使用示例：
```go
data := Student{
	Userid:  19827,
	ClassId: 2,
	Sex:     "male",
	Age:     23,
	Name:    "jerry",
	Status:  1,
}

err := horm.NewQuery("skynet").Prefix("student_").HmSetStruct("19827", data).Exec(ctx)
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
err := horm.NewQuery("skynet").HmGet("student_19827", fields).Exec(ctx, &result)

var result map[string]interface{} //也可以通过一个 map 接收结果
err := horm.NewQuery("skynet").HmGet("student_19827", fields).Exec(ctx, &result)
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
err = horm.NewQuery("skynet").HGet("test_hset", "xxxxxx").Exec(ctx, &bval)
if horm.IsError(err) {
	return err
}

if horm.IsNil(err) { // err = nil returned
	fmt.Println("key not set")
}

err = horm.NewQuery("skynet").HGet("test_hset", "aa").Exec(ctx, &ival)
err = horm.NewQuery("skynet").HGet("test_hset", "bb").Exec(ctx, &fval)
err = horm.NewQuery("skynet").HGet("test_hset", "cc").Exec(ctx, &strval)
err = horm.NewQuery("skynet").HGet("test_hset", "ee").Exec(ctx, &structVal)
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
err := horm.NewQuery("skynet").HGETALL("student_19827", fields).Exec(ctx, &result)

var result map[string]interface{} //也可以通过一个 map 接收结果
err := horm.NewQuery("skynet").HGETALL("student_19827", fields).Exec(ctx, &result)
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
err := horm.NewQuery("skynet").Hkeys("student_19827").Exec(ctx, &result)
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
err := horm.NewQuery("skynet").HIncrBy("student_19827", "age", 5).Exec(ctx, &newAge)
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
err := horm.NewQuery("skynet").HDel("student", 19827, 23312, 98322).Exec(ctx, &delNum)
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
var exist bool // field 是否存在，存在
err := horm.NewQuery("skynet").HExists("student", 19827).Exec(ctx, &exist)
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
err := horm.NewQuery("skynet").HLen("student").Exec(ctx, &num)
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
err := horm.NewQuery("skynet").HStrLen("student_19827", "name").Exec(ctx, &l)
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
err := horm.NewQuery("skynet").HIncrByFloat("student_19827", "score", 0.7).Exec(ctx, &score)
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
err := horm.NewQuery("skynet").HVals("student_age").Exec(ctx, &ages)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var n int // 执行 LPUSH 命令后，列表的长度。
err := horm.NewQuery("skynet").LPush("student_list", data).Exec(ctx, &n)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var n int // 执行 RPUSH 命令后，列表的长度。
err := horm.NewQuery("skynet").RPush("student_list", data).Exec(ctx, &n)
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
err := horm.NewQuery("skynet").LPop("student_list").Exec(ctx, &result)
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
err := horm.NewQuery("skynet").RPop("student_list").Exec(ctx, &result)
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
err := horm.NewQuery("skynet").LLen("student_list").Exec(ctx, &n)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var n int //被添加到集合中的新元素的数量，不包括被忽略的元素。
err := horm.NewQuery("skynet").SAdd("student_set", data).Exec(ctx, &n)
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
err := horm.NewQuery("skynet").SMembers("student_set").Exec(ctx, &result)
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
err := horm.NewQuery("skynet").SRem("student_set", data).Exec(ctx, &n)
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
err := horm.NewQuery("skynet").SCard("student_set").Exec(ctx, &n)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var isMember bool
err := horm.NewQuery("skynet").SIsMember("student_set", data).Exec(ctx, &isMember)
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
err := horm.NewQuery("skynet").SRandMember("student_set", 2).Exec(ctx, &result)
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
err := horm.NewQuery("skynet").SPop("student_set", 2).Exec(ctx, &result)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

err := horm.NewQuery("skynet").SMOVE("student_set", "student_set2", &data).Exec(ctx)
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
data1 := Student{
	Userid:  16883,
	ClassId: 1,
	Sex:     "male",
	Age:     19,
	Name:    "evan",
	Status:  1,
}

data2 := Student{
	Userid:  18723,
	ClassId: 2,
	Sex:     "female",
	Age:     18,
	Name:    "yolanda",
	Status:  1,
}

var n int // 被成功添加的新成员的数量，不包括那些被更新的、已经存在的成员。
err := horm.NewQuery("skynet").ZAdd("student_zset", &data1, 97, &data2, 89).Exec(ctx, &n)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var n int // 被成功移除的成员的数量，不包括被忽略的成员。
err := horm.NewQuery("skynet").ZRem("student_zset", &data).Exec(ctx, &n)
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
err := horm.NewQuery("skynet").ZRemRangeByScore("student_zset", 90, 100).Exec(ctx, &n)
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
err := horm.NewQuery("skynet").ZRemRangeByRank("student_zset", 0, 2).Exec(ctx, &n)
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
err := horm.NewQuery("skynet").ZCard("student_zset").Exec(ctx, &n)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var score float64
err := horm.NewQuery("skynet").ZScore("student_zset", &data).Exec(ctx, &score)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var rank int
err := horm.NewQuery("skynet").ZRank("student_zset", &data).Exec(ctx, &rank)
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
data := Student{
	Userid:  19827,
	ClassId: 1,
	Sex:     "male",
	Age:     23,
	Name:    "smallhow",
	Status:  1,
}

var rank int
err := horm.NewQuery("skynet").ZRevRank("student_zset", &data).Exec(ctx, &rank)
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
err := horm.NewQuery("skynet").ZCount("student_zset", 70, 80).Exec(ctx, &n)
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
err = horm.NewQuery("skynet").ZPopMin("student_zset", 2).Exec(ctx, &result, scores)
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
err = horm.NewQuery("skynet").ZPopMax("student_zset", 2).Exec(ctx, &result, &scores)
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
err = horm.NewQuery("skynet").ZIncrBy("student_zset", &data, 9.5).Exec(ctx, &newScore)
```
- `ZRANGE`
  函数用法：
```go
// ZRange 返回有序集中，指定区间内的成员。其中成员的位置按分数值递增(从小到大)来排序。
// param: key string
// param: int start, stop 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
ZRange(key string, start, stop int)
```
使用示例：
```go
var result []*Student
err = horm.NewQuery("skynet").ZRange("student_zset", 0, 9).Exec(ctx, &result)
```
- `ZRANGE WITHSCORES`
  函数用法：
```go
// ZRangeWithScore 返回有序集中指定区间的成员(从小到大)和他们的分数。
// param: key string
// param: int start, stop 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
ZRangeWithScore(key string, start, stop int)
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
err = horm.NewQuery("skynet").ZRangeWithScore("student_zset", 0, 9).Exec(ctx, &result, &scores)
```
返回结果：

- `ZRANGEBYSCORE`
  函数用法：
```go
// ZRangeByScore 根据分数返回有序集中的成员
// param: key string
// param: int min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
ZRangeByScore(key string, min, max interface{})
```
使用示例：
```go
var result []*Student // 80分~90分之间
err = horm.NewQuery("skynet").ZRangeByScore("student_zset", 80, 90).Exec(ctx, &result)
```
- `ZRangeByScoreWithLimit`
  函数用法：
```go
// ZRangeByScoreWithLimit 根据分数返回有序集中指定区间的成员(从小到大)
// param: key string
// param: interface{} min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
// param: LIMIT offset count 游标
ZRangeByScoreWithLimit(key string, min, max interface{}, offset, count int64)
```
使用示例：
```go
var result []*Student // 80分~90分之间，OFFSET=0，LIMIT=10
err = horm.NewQuery("skynet").ZRangeByScoreWithLimit("student_zset", 80, 90, 0, 10).Exec(ctx, &result)
```
- `ZRangeByScoreWithScoreWithLimit`
  函数用法：
```go
// ZRangeByScoreWithScoreWithLimit 根据分数返回有序集中指定区间的成员和他们的分数 分开在两个数组存储，但是数组下标是一一对应的。
// param: key string
// param: interface{} min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
// param: LIMIT offset count 游标
ZRangeByScoreWithScoreWithLimit(key string, min, max interface{}, offset, count int64)
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
err = horm.NewQuery("skynet").ZRangeByScoreWithScoreWithLimit("student_zset", 70, 100, 0, 10).Exec(ctx, &result, &scores)
```
返回结果：

- `ZRangeByScoreWithScore`
  函数用法：
```go
// ZRangeByScoreWithScore 根据分数返回有序集中的成员和他们分数（从小大到大排列），分开在两个数组存储，但是数组下标是一一对应的，
// param: key string
// param: interface{} min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
ZRangeByScoreWithScore(key string, min, max interface{})
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
err = horm.NewQuery("skynet").ZRangeByScoreWithScore("student_zset", 70, 100).Exec(ctx, &result, &scores)
```
- `ZREVRANGE`
  函数用法：
```go
// ZRevRange 返回有序集中指定区间的成员，其中成员的位置按分数值递减(从大到小)来排列。
// param: key string
// param: start, stop 排名区间，以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
ZRevRange(key string, start, stop int)
```
使用示例：
```go
var result []*Student
err = horm.NewQuery("skynet").ZRevRange("student_zset", 0, 9).Exec(ctx, &result)
```
- `ZREVRANGE WITHSCORES`
  函数用法：
```go
// ZRevRangeWithScore 返回有序集中指定区间的成员和他们的分数，分开在两个数组存储，但是数组下标是一一对应的。
// param: key string
// param: int start, stop 排名区间，以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，
// 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
ZRevRangeWithScore(key string, start, stop int)
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
err = horm.NewQuery("skynet").ZRevRangeWithScore("student_zset", 0, 9).Exec(ctx, &result, &scores)
```
返回结果：
- `ZREVRANGEBYSCORE`
  函数用法：
```go
// ZRevRangeByScore 返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。
// param: key string
// param: min , max interface{} 分数区间，类型为整数或双精度浮点数，但是 -inf +inf 表示负正无穷大
ZRevRangeByScore(key string, min, max interface{})
```
使用示例：
```go
var result []*Student // 80分~90分之间
err = horm.NewQuery("skynet").ZRevRangeByScore("student_zset", 80, 90).Exec(ctx, &result)
```
- `ZRevRangeByScoreWithLimit`
  函数用法：
```go
// ZRevRangeByScoreWithLimit 返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。
// param: key string
// param: min , max interface{} 分数区间，类型为整数或双精度浮点数，但是 -inf +inf 表示负正无穷大
// param: LIMIT offset count 游标
ZRevRangeByScoreWithLimit(key string, min, max interface{}, offset, count int64)
```
使用示例：
```go
var result []*Student // 80分~90分之间，OFFSET=0，LIMIT=10
err = horm.NewQuery("skynet").ZRevRangeByScoreWithLimit("student_zset", 80, 90, 0, 10).Exec(ctx, &result)
```
- `ZRevRangeByScoreWithScoreWithLimit`
  函数用法：
```go
// ZRevRangeByScoreWithScoreWithLimit 根据分数返回有序集中的成员和他们分数，分开在两个数组存储，但是数组下标是一一对应的，
// param: key string
// param: min , max interface{} 分数区间，类型为整数或双精度浮点数，但是 -inf +inf 表示负正无穷大
// param: LIMIT offset count 游标
ZRevRangeByScoreWithScoreWithLimit(key string, min, max interface{}, offset, count int64)
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
err = horm.NewQuery("skynet").ZRevRangeByScoreWithScoreWithLimit("student_zset", 70, 100, 0, 10).Exec(ctx, &result, &scores)
```

返回结果：


- `ZRevRangeByScoreWithScore`
  函数用法：
```go
// ZRevRangeByScoreWithScore 根据分数（从打到小排列）返回有序集中的成员和他们分数，分开在两个数组存储，但是数组下标是一一对应的，
// param: key string
// param: min, max interface{} 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大
ZRevRangeByScoreWithScore(key string, min, max interface{})
```
使用示例：
```go
var result []*Student
var scores []float64 // 返回成员的分数，与 result 数组顺序一一对应。
err = horm.NewQuery("skynet").ZRevRangeByScoreWithScore("student_zset", 70, 100).Exec(ctx, &result, &scores)
```