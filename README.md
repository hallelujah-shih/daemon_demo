# daemon manager demo
```
一个demo程序
```

## 简单运行
```
> cd $PROJECT_DIR
> dep ensure
> cd cmd/daemon_manager
> go build
> ./daemon_manager -h

配置文件模板在$PROJECT_DIR/config/config.yaml中
```

## TODO
```
0. 单元测试
1. uidgid设置问题（导致权限不够）
2. 资源限制相关
3. 动态更新配置文件
4. restart service
5. api部分找个能自描述的
6. 进程信息细化
7. 管理程序自身的信号处理，以及restart等细节问题
...
```
