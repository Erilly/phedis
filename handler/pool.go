package handler

import (
	"fmt"
	"git.btime.cn/btime_new/phedis/config"
	"sync"
)

//池中获取redis连接
func NewRedisConnect(opt *RedisOptions) (*Connect,error){
	var mu sync.RWMutex
	var connect *Connect
	var err error

	mu.Lock()

	if len(RedisProxyPool)==0{
		//初始化池中连接个数
		for i:=0;i<config.Configs.MinProxyPoolLength;i++{
			connect,err = ConnetRedis(opt)

			if err!=nil{
				break
			}else{
				RedisProxyPool = append(RedisProxyPool,connect)
			}

		}
	}

	if len(RedisProxyPool)>0{
		connect = RedisProxyPool[0]
		RedisProxyPool = RedisProxyPool[1:]
	}

	mu.Unlock()
	fmt.Printf("RedisProxyPool:%v, length %v\n",RedisProxyPool,len(RedisProxyPool))
	return connect,err
}

//连接归还给池中
func GiveBack(conn *Connect)bool{
	r:=false

	if conn != nil{
		var mu sync.RWMutex
		mu.Lock()
		if len(RedisProxyPool)>=config.Configs.MaxProxyPoolLength{
			conn.Close()
			r=false
		}else{
			RedisProxyPool = append(RedisProxyPool,conn)
			r=true
		}

		mu.Unlock()
	}
	return r
}