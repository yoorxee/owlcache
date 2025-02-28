package network

import (
	"fmt"

	"github.com/xssed/owlcache/cache"
	owlconfig "github.com/xssed/owlcache/config"
	"github.com/xssed/owlcache/counter"
	"github.com/xssed/owlcache/group"
	owllog "github.com/xssed/owlcache/log"
	"github.com/xssed/owlcache/network/httpclient"
	"github.com/xssed/owlcache/network/memcacheclient"
	"github.com/xssed/owlcache/network/redisclient"
)

//创建一个全局的缓存DB
var BaseCacheDB *cache.BaseCache

//创建一个全局的身份认证缓存
var BaseAuth *cache.BaseCache

//创建一个全局的服务器集群信息存储列表
var ServerGroupList *group.Servergroup

//创建一个全局的HttpClient客户端
var HttpClient *httpclient.OwlClient

//创建一个全局的HttpClient错误请求控制计数器
var HttpClientRequestErrorCounter *counter.Counter

//创建一个全局的MemcacheClient错误请求控制计数器
var MemcacheClientRequestErrorCounter *counter.Counter

//创建一个全局的RedisClient错误请求控制计数器
var RedisClientRequestErrorCounter *counter.Counter

func BaseCacheDBInit() {

	//执行步骤信息
	fmt.Println("owlcache  database running...")

	//创建DB
	BaseCacheDB = cache.NewCache("owlcache")

	//加载之前缓存本地的DB数据
	BaseCacheDB.LoadFromFile(owlconfig.OwlConfigModel.DBfile + "owlcache.db")

	//身份认证数据,所有客户端身份认证都在这里有效期60分钟
	//存储内容: key:tonken  value:"uuid"
	BaseAuth = cache.NewCache("Auth")

	//加载之前缓存本地的DB数据
	BaseAuth.LoadFromFile(owlconfig.OwlConfigModel.DBfile + "auth.db")

	//初始化服务器集群信息存储列表
	ServerGroupList = group.NewServergroup()

	//加载之前缓存本地的服务器集群信息
	ServerGroupList.LoadFromFile(owlconfig.OwlConfigModel.DBfile, "server_group_config.json")

	//初始化HttpClient客户端
	HttpClient = httpclient.NewOwlClient()

	//初始化HttpClient错误请求控制计数器
	HttpClientRequestErrorCounter = counter.NewCounter()

	//初始化MemcacheClient错误请求控制计数器
	MemcacheClientRequestErrorCounter = counter.NewCounter()

	//初始化RedisClient错误请求控制计数器
	RedisClientRequestErrorCounter = counter.NewCounter()

	//启动tcp服务
	//检查是否开启TCP服务。默认为开启。
	if owlconfig.OwlConfigModel.CloseTcp == "1" {
		fmt.Println("owlcache  tcp server running...")
		go startTCP()
	} else if owlconfig.OwlConfigModel.CloseTcp == "0" {
		owllog.OwlLogRun.Info("The configuration file does not open the TCP service.")
	} else {
		//检测到配置书写异常强制退出
		owllog.OwlLogRun.Fatal(ErrorCloseTcp)
	}

	//启动http服务
	fmt.Println("owlcache  http server running...")
	go startHTTP()

	//启动gossip数据最终一致服务
	//检查是否开启gossip服务。默认为关闭。
	if owlconfig.OwlConfigModel.GroupWorkMode == "gossip" {
		fmt.Println("owlcache  final consistency service running...")
		go startGossip()
	} else if owlconfig.OwlConfigModel.GroupWorkMode == "owlcache" {
		//什么也不做，集群方式默认开启owlcache
	} else {
		//检测到配置书写异常强制退出
		owllog.OwlLogRun.Fatal(ErrorGroupWorkMode)
	}

	//启动是否启动从memcache中查询数据
	if owlconfig.OwlConfigModel.Get_data_from_memcache == "1" {
		fmt.Println("owlcache  memcache client service running...")
		go memcacheclient.Start()
	}

	//启动是否启动从memcache中查询数据
	if owlconfig.OwlConfigModel.Get_data_from_redis == "1" {
		fmt.Println("owlcache  redis client service running...")
		go redisclient.Start()
	}

}
