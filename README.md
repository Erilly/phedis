## Description

phedis是使用golang开发的PHP的redis代理连接池，兼容现有PHP的redis.so扩展。主要解决nginx+php-fpm架构下，面对高并发时redis连接数不够用，redis性能无法彻底释放的问题。

## Installation

`go get github.com/Erilly/phedis`

## Toml config
`vim phedis/conf/phedis_config.toml`

``` toml
listen=":8383"                                  #服务监听端口
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
  通过phedis/conf/phedis_config.toml约定配置的phredis_proxy_key，来设置需代理连接的真实redis配置。
  格式： host:port,password,database，如：127.0.0.1:6379,732de51677407fa6,3
  如果redis无密码，如：127.0.0.1:6379,,3
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
