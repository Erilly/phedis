## Description

PHP 的redis 代理连接池

## Installation

`go get github.com/Erilly/phedis`

## Toml config
`vim phedis/conf/phedis_config.toml`

``` toml
listen=":8383"  #服务监听端口
phredis_proxy_key = "#phedisProxy_options#"     #PHP端要连接的真实redis配置key
timeout = 500000000                             #连接redis的超时时间5s
min_proxy_pool_length = 10                      #初始化池中连接个数
max_proxy_pool_length = 20                      #池中最大连接个数
```

## Start

`go build && ./phedis`

## PHP example

``` php
$redis = new \Redis();
//连接phedis
$redis->connect('127.0.0.1',8383);
 
/*
  通过phedis中指定的phredis_proxy_key，来设置需代理连接的真实redis配置。
  格式： host:port,password,database，如：127.0.0.1:6379,732de51677407fa6,3
*/
$bool = $redis->set('#phedisProxy_options#','127.0.0.1:6379,732de51677407fa6,11');

if ($bool){
  $redis->set();
  $redis->get();
  $redis->hset();
  $redis->hgset();

  ...
}

```
