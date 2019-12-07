### 3rd
<img src="https://camo.githubusercontent.com/20065c601e7a02381bee3a76007952d94dcb3f2d/68747470733a2f2f7472617669732d63692e6f72672f6f6c69766572652f656c61737469632e7376673f6272616e63683d72656c656173652d6272616e63682e7636" alt="Build Status" style="max-width:100%;"/>
<img src="https://camo.githubusercontent.com/2afd9bd1040ee3eca5d58a22b67cf1fd86057ffb/687474703a2f2f696d672e736869656c64732e696f2f62616467652f676f646f632d7265666572656e63652d626c75652e7376673f7374796c653d666c6174" alt="Godoc" style="max-width:100%;"/>
<img src="https://camo.githubusercontent.com/80a3f4387388a340de7d3b66176e2a53c56d2ea5/687474703a2f2f696d672e736869656c64732e696f2f62616467652f6c6963656e73652d4d49542d7265642e7376673f7374796c653d666c6174" alt="license" style="max-width:100%;"/>
<img src="https://camo.githubusercontent.com/98181c2cd08d758e1824b7466ebd326f5f81c1ac/68747470733a2f2f6170702e666f7373612e696f2f6170692f70726f6a656374732f6769742532426769746875622e636f6d2532466f6c6976657265253246656c61737469632e7376673f747970653d736869656c64" alt="FOSSA Status" style="max-width:100%;"/>


### feature
 * 文件操作 files
   * xml
     * []byte ---> xxx.xml
   * pdf 
     * []byte ---> xxx.pdf
     * url    ---> xxx.pdf
   * zip
     * []byte ---> xxx.zip
 * 日志
   * txt
 * 通知
   * email
 * 池
   * goroutine pool
 * 队列
   * 普通队列 t
 * 存储
   * LevelDB
   * ES7
   * Redis
 * 缓存
   * Redis
 * WebSocket
 * RPC 
 * HTTP
   * client
      * get
      * post
      * postForm
   * server
 * 工具包
   * uuid
   * snow flake id
   * 数值处理
   * 随机数/随机字符串
   * 时间处理
        
#### init
````shell
    export GOPROXY=https://goproxy.cn
    export GO111MODULE=on
    
    check list
    https://goproxy.io/github.com/go-redis/redis/@v/
                       -------------------------
                              package
                                 |
                               *.info
````

#### go mod
````shell
    //清缓存
    $ go clean -modcache

    go.mod：依赖列表和版本约束。
    go.sum：记录module文件hash值，用于安全校验
````

#### question
````md
    q: $GOPATH/go.mod exists but should not
    a: mod 不能与gopath共存
````
