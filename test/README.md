# 简单的测试步骤

## 配置文件
```
manager_config:
  serve_addr: "localhost:1234"

process_configs:
  process_name1:
    soft_limit:
      limit_file_descriptor: 65536
    uid_gid: ""
    envs:
      - http_proxy=http://127.0.0.1:8888/
      - https_proxy=http://127.0.0.1:8888/
      - PATH=/python_env_path_bin:$PATH
    command: "python"
    command_args:
      - "/tmp/hello.py"
    stderr_path: "/tmp/err.log"
    stdout_path: "/tmp/out.log"
```

## 启动管理程序
```
> ./daemon_manager -c config.yaml
```

## 操作子进程
```
# list process
> curl -s localhost:1234/process |python -m json.tool
RUNNING状态

# stop process
> curl -s localhost:1234/process/stop?name=process_name1 | python -m json.tool
EXIT状态

# start process
> curl -s localhost:1234/process/start?name=process_name1 |python -m json.tool
RUNNING状态

# signal
> curl -s "localhost:1234/process/signal?name=process_name1\&signal=15" | python -m json.tool
退出后又拉起来
```