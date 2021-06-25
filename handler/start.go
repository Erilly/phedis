package handler

import (
	"fmt"
	"git.btime.cn/btime_new/phedis/config"
	"github.com/tidwall/redcon"
	"log"
	"strings"
)

var (
	RemoteAddr     string		//客户端
	RedisProxyPool []*Connect	//redis代理池
	ClientPoolMap  map[string]string	//当前请求的客户端addr对应的redis唯一标识
	RedisPoolMap   map[string]*Connect	//redis唯一标识对应的redis连接
)

func Start() {
	config.Start()

	RedisProxyPool = make([]*Connect, 0, config.Configs.MinProxyPoolLength)
	ClientPoolMap = make(map[string]string)
	RedisPoolMap = make(map[string]*Connect)

	log.Printf("started server at %s", config.Configs.Listen)

	err := redcon.ListenAndServe(config.Configs.Listen,
		redServ,
		func(conn redcon.Conn) bool {
			// Use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)

			//归还redis连接进连接池
			GiveBack(RedisPoolMap[ClientPoolMap[RemoteAddr]])

			//删除已断开的客户端连接
			delete(ClientPoolMap, conn.RemoteAddr())
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}

func redServ(conn redcon.Conn, cmd redcon.Command) {

	//客户端RemoteAddr作为本次请求的唯一标识
	RemoteAddr = conn.RemoteAddr()

	//拿取自己的redis服务连接
	redisServer := RedisPoolMap[ClientPoolMap[RemoteAddr]]

	//查看key是否是请求代理连接的PhredisProxyKey
	phredisOptionsKey := string(cmd.Args[1])
	if redisServer == nil && config.Configs.PhredisProxyKey != phredisOptionsKey {
		conn.WriteError("ERR connection lost!")
		return
	}

	switch strings.ToLower(string(cmd.Args[0])) {
	case "set":
		if len(cmd.Args) > 3 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}

		//根据判断指定PhredisProxyKey代理连接指定redis
		if config.Configs.PhredisProxyKey == phredisOptionsKey {
			var setOptions []string

			//client端请求的 host:port,password,database
			//127.0.0.1:6379,732de51677407fa6,3
			setOptions = strings.Split(string(cmd.Args[2]), ",")
			if len(setOptions) != 3 {
				conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[2]) + "' command")
			}

			//连接redis所需的相关参数
			redOptions := &RedisOptions{setOptions[0], setOptions[1], setOptions[2]}
			fmt.Printf("redOptions --> %v <--\n",redOptions)

			//代理连接redis
			redisServer, err := NewRedisConnect(redOptions)

			if err != nil {
				conn.WriteError("ERR Connect for " + string(cmd.Args[2]))
				return
			}

			//redis唯一标识host:port_db
			//如：127.0.0.1:6379_3
			redisConnectIndex := fmt.Sprintf("%s_%s", setOptions[0], setOptions[2])

			//记录client端所连接的redis
			ClientPoolMap[RemoteAddr] = redisConnectIndex

			//将redis连接关联本次客户端
			RedisPoolMap[redisConnectIndex] = redisServer

			conn.WriteString("OK")
			return
		} else {
			err := redisServer.Write(cmd.Raw)
			if err != nil {
				conn.WriteError(err.Error())
				return
			}

			conn.WriteRaw(redisServer.Reply())
			return
		}
	default:
		err := redisServer.Write(cmd.Raw)
		if err != nil {
			conn.WriteError(err.Error())
			return
		}

		conn.WriteRaw(redisServer.Reply())
		return
	}
}
