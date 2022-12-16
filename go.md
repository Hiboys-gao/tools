# GO 总结

## golang

### 编译

```
go build [-o 输出名] [-i] [编译标记] [包名]

//windows --> linux
SET CGO_ENABLED=0
SET GOARCH=amd64
SET GOOS=linux
go build main.go

//linux --> windows
SET CGO_ENABLED=1
SET GOOS=windows
SET GOARCH=amd64
go build main.go
(CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go)
```

### 协程

提前声明**defer recover**函数，这样可以保证协程内部崩溃，不会将整个进程崩溃掉。

1. WaitGroup

   ```go
   package main
   
   import "sync"
   var wg = sync.WaitGroup{}
   
   func main() {
   	wg.Add(1)
   	go say("Hello World")
   	wg.Wait()
   }
   
   func say(s string) {
   	println(s)
   	wg.Done()
   }
   ```

   

2. channel

   通道是一种允许一个`goroutine`将数据发送到另一个`goroutine`的技术。默认情况下，通道是双向的，这意味着goroutine可以通过同一通道发送或接收数据

   `chan string`这样的写法能够使用读写功能双向管道外，还可以创建出单向管道，如`<-chan string`只能从管道中读取数据，而`chan<- string`只能够向管道中写入数据。

3. 协程的超时处理

   ctx. Done() time. After()  time. Ticket()

   - time. After():

   `Timer` 不会被 GC 回收直到它被触发，如果需要考虑效率的话，`Timer` 不再被需要时，需要主动调用 `Timer.Stop`。

   ```go
   func AsyncCallWithTimeout2() {
       ctx, cancel := context.WithCancel(context.Background())
       //ctx, cancel := context.WithTimeOut(context.Background(), time.Duration())
   
       go func() {
           defer cancel()
           // 模拟请求调用
           time.Sleep(200 * time.Millisecond)
       }()
   
       timer := time.NewTimer(3 * time.Second)
       defer timer.Stop()
       select {
       case <-ctx.Done():
           // fmt.Println("call successfully!!!")
           return
       case <-timer.C:
           // fmt.Println("timeout!!!")
           return
       }
   }
   ```

   

   ### gorm
   
   ```go
   type User struct {
     gorm.Model
     Name string
   }
    
   type Profile struct {
     gorm.Model
     Name      string
     User      User `gorm:"foreignkey:UserRefer"` // use UserRefer as foreign key
     UserRefer uint
   }
   
   var profile Profile
   user := &User{ ID: 111 }
   db.Model(&user).Related(&profile)
   // SELECT * FROM profiles WHERE user_id = 111;
   
   //求和
   var amountReal, amountPlanned float64
   // _ = m.db.Model(&model_table.FinancialIncome{}).Where("`year` = ? AND `month` = ?", year, month).Pluck("COALESCE(SUM(planned_income), 0) as amount_planned", &amountPlanned).Error
   	tx := m.db.Model(&model_table.FinancialIncome{}).Where("`year` = ? AND `month` = ?", year, month)
   	row := tx.Select("COALESCE(SUM(real_income), 0) as amount_real", "COALESCE(SUM(planned_income), 0) as amount_planned").Row()
   	err = row.Scan(&amountReal, &amountPlanned)
   
   //删除关联关系
   db.Debug().Model(&d3).Association("GirlGODs").Delete(&g2)
   //清空关联关系
   db.Model(&authority).Association("SysBaseMenus").Clear()
   //替换关联关系
   db.Model(&d3).Association("GirlGODs").Replace(&g2)
   ```
   
   ### flag
   
   1. 控制台 help功能的实现
   
   ```go
   // 实际中应该用更好的变量名
   var (
   	h, v, V, t, T bool
   	s, p, c, g    string
   	q             *bool
   )
   
   func init() {
   	flag.BoolVar(&h, "h", false, "this help")
   	flag.BoolVar(&v, "v", false, "show version and exit")
   	flag.BoolVar(&V, "V", false, "show version and configure options then exit")
   	flag.BoolVar(&t, "t", false, "test configuration and exit")
   	flag.BoolVar(&T, "T", false, "test configuration, dump it and exit")
   	// 另一种绑定方式
   	q = flag.Bool("q", false, "suppress non-error messages during configuration testing")
   	// 注意 `signal`。默认是 -s string，有了 `signal` 之后，变为 -s signal
   	flag.StringVar(&s, "s", "", "send `signal` to a master process: stop, quit, reopen, reload")
   	flag.StringVar(&p, "p", "/usr/local/nginx/", "set `prefix` path")
   	flag.StringVar(&c, "c", "conf/nginx.conf", "set configuration `file`")
   	flag.StringVar(&g, "g", "conf/nginx.conf", "set global `directives` out of configuration file")
   	// 改变默认的 Usage
   	flag.Usage = usage
   }
   
   func main() {
   	flag.Parse()
   	if h {
   		flag.Usage()
   	}
   }
   
   func usage() {
   	fmt.Fprintf(os.Stderr, `nginx version: nginx/1.10.0
   
   Options:
   `)
   	flag.PrintDefaults()
   }
   ```
   
   2. 自定义参数类型
   
   ```go
   type interval []time.Duration
   
   // 实现String接口
   func (i *interval) String() string {
   	return fmt.Sprintf("%v", *i)
   }
   
   // 实现Set接口,Set接口决定了如何解析flag的值
   func (i *interval) Set(value string) error {
   	//此处决定命令行是否可以设置多次-deltaT
   	if len(*i) > 0 {
   		return errors.New("interval flag already set")
   	}
   	for _, dt := range strings.Split(value, ",") {
   		duration, err := time.ParseDuration(dt)
   		if err != nil {
   			return err
   		}
   		*i = append(*i, duration)
   	}
   	return nil
   }
   
   var intervalFlag interval
   
   func init() {
   	flag.Var(&intervalFlag, "deltaT", "comma-separated list of intervals to use between events")
   }
   
   func main() {
   	flag.Parse()
   	fmt.Println(intervalFlag)
   }
   ```
   
   

 ## LINUX

```
df -h  //磁盘使用状况
netstat -ntpl  //端口使用状况
fuser -n tcp 80 //查看80端口的占用
```



## NGINX

```
创建目录/var/temp/nginx/   mkdir /var/temp/nginx -p

下载后解压到 /usr/local
到解压后的nginx文件夹中执行 ./configure \  （可以make）
						make
						make install



```

```
######Nginx配置文件nginx.conf中文详解#####

#定义Nginx运行的用户和用户组
user www www;

#nginx进程数，建议设置为等于CPU总核心数。
worker_processes 8;
 
#全局错误日志定义类型，[ debug | info | notice | warn | error | crit ]
error_log /usr/local/nginx/logs/error.log info;

#进程pid文件
pid /usr/local/nginx/logs/nginx.pid;

#指定进程可以打开的最大描述符：数目
#工作模式与连接数上限
#这个指令是指当一个nginx进程打开的最多文件描述符数目，理论值应该是最多打开文件数（ulimit -n）与nginx进程数相除，但是nginx分配请求并不是那么均匀，所以最好与ulimit -n 的值保持一致。
#现在在linux 2.6内核下开启文件打开数为65535，worker_rlimit_nofile就相应应该填写65535。
#这是因为nginx调度时分配请求到进程并不是那么的均衡，所以假如填写10240，总并发量达到3-4万时就有进程可能超过10240了，这时会返回502错误。
worker_rlimit_nofile 65535;


events
{
    #参考事件模型，use [ kqueue | rtsig | epoll | /dev/poll | select | poll ]; epoll模型
    #是Linux 2.6以上版本内核中的高性能网络I/O模型，linux建议epoll，如果跑在FreeBSD上面，就用kqueue模型。
    #补充说明：
    #与apache相类，nginx针对不同的操作系统，有不同的事件模型
    #A）标准事件模型
    #Select、poll属于标准事件模型，如果当前系统不存在更有效的方法，nginx会选择select或poll
    #B）高效事件模型
    #Kqueue：使用于FreeBSD 4.1+, OpenBSD 2.9+, NetBSD 2.0 和 MacOS X.使用双处理器的MacOS X系统使用kqueue可能会造成内核崩溃。
    #Epoll：使用于Linux内核2.6版本及以后的系统。
    #/dev/poll：使用于Solaris 7 11/99+，HP/UX 11.22+ (eventport)，IRIX 6.5.15+ 和 Tru64 UNIX 5.1A+。
    #Eventport：使用于Solaris 10。 为了防止出现内核崩溃的问题， 有必要安装安全补丁。
    use epoll;

    #单个进程最大连接数（最大连接数=连接数*进程数）
    #根据硬件调整，和前面工作进程配合起来用，尽量大，但是别把cpu跑到100%就行。每个进程允许的最多连接数，理论上每台nginx服务器的最大连接数为。
    worker_connections 65535;

    #keepalive超时时间。
    keepalive_timeout 60;

    #客户端请求头部的缓冲区大小。这个可以根据你的系统分页大小来设置，一般一个请求头的大小不会超过1k，不过由于一般系统分页都要大于1k，所以这里设置为分页大小。
    #分页大小可以用命令getconf PAGESIZE 取得。
    #[root@web001 ~]# getconf PAGESIZE
    #4096
    #但也有client_header_buffer_size超过4k的情况，但是client_header_buffer_size该值必须设置为“系统分页大小”的整倍数。
    client_header_buffer_size 4k;

    #这个将为打开文件指定缓存，默认是没有启用的，max指定缓存数量，建议和打开文件数一致，inactive是指经过多长时间文件没被请求后删除缓存。
    open_file_cache max=65535 inactive=60s;

    #这个是指多长时间检查一次缓存的有效信息。
    #语法:open_file_cache_valid time 默认值:open_file_cache_valid 60 使用字段:http, server, location 这个指令指定了何时需要检查open_file_cache中缓存项目的有效信息.
    open_file_cache_valid 80s;

    #open_file_cache指令中的inactive参数时间内文件的最少使用次数，如果超过这个数字，文件描述符一直是在缓存中打开的，如上例，如果有一个文件在inactive时间内一次没被使用，它将被移除。
    #语法:open_file_cache_min_uses number 默认值:open_file_cache_min_uses 1 使用字段:http, server, location  这个指令指定了在open_file_cache指令无效的参数中一定的时间范围内可以使用的最小文件数,如果使用更大的值,文件描述符在cache中总是打开状态.
    open_file_cache_min_uses 1;
    
    #语法:open_file_cache_errors on | off 默认值:open_file_cache_errors off 使用字段:http, server, location 这个指令指定是否在搜索一个文件时记录cache错误.
    open_file_cache_errors on;
}
 
 
 
#设定http服务器，利用它的反向代理功能提供负载均衡支持
http
{
    #文件扩展名与文件类型映射表
    include mime.types;

    #默认文件类型
    default_type application/octet-stream;

    #默认编码
    #charset utf-8;

    #服务器名字的hash表大小
    #保存服务器名字的hash表是由指令server_names_hash_max_size 和server_names_hash_bucket_size所控制的。参数hash bucket size总是等于hash表的大小，并且是一路处理器缓存大小的倍数。在减少了在内存中的存取次数后，使在处理器中加速查找hash表键值成为可能。如果hash bucket size等于一路处理器缓存的大小，那么在查找键的时候，最坏的情况下在内存中查找的次数为2。第一次是确定存储单元的地址，第二次是在存储单元中查找键 值。因此，如果Nginx给出需要增大hash max size 或 hash bucket size的提示，那么首要的是增大前一个参数的大小.
    server_names_hash_bucket_size 128;

    #客户端请求头部的缓冲区大小。这个可以根据你的系统分页大小来设置，一般一个请求的头部大小不会超过1k，不过由于一般系统分页都要大于1k，所以这里设置为分页大小。分页大小可以用命令getconf PAGESIZE取得。
    client_header_buffer_size 32k;

    #客户请求头缓冲大小。nginx默认会用client_header_buffer_size这个buffer来读取header值，如果header过大，它会使用large_client_header_buffers来读取。
    large_client_header_buffers 4 64k;

    #设定通过nginx上传文件的大小
    client_max_body_size 8m;

    #开启高效文件传输模式，sendfile指令指定nginx是否调用sendfile函数来输出文件，对于普通应用设为 on，如果用来进行下载等应用磁盘IO重负载应用，可设置为off，以平衡磁盘与网络I/O处理速度，降低系统的负载。注意：如果图片显示不正常把这个改成off。
    #sendfile指令指定 nginx 是否调用sendfile 函数（zero copy 方式）来输出文件，对于普通应用，必须设为on。如果用来进行下载等应用磁盘IO重负载应用，可设置为off，以平衡磁盘与网络IO处理速度，降低系统uptime。
    sendfile on;

    #开启目录列表访问，合适下载服务器，默认关闭。
    autoindex on;

    #此选项允许或禁止使用socke的TCP_CORK的选项，此选项仅在使用sendfile的时候使用
    tcp_nopush on;
     
    tcp_nodelay on;

    #长连接超时时间，单位是秒
    keepalive_timeout 120;

    #FastCGI相关参数是为了改善网站的性能：减少资源占用，提高访问速度。下面参数看字面意思都能理解。
    fastcgi_connect_timeout 300;
    fastcgi_send_timeout 300;
    fastcgi_read_timeout 300;
    fastcgi_buffer_size 64k;
    fastcgi_buffers 4 64k;
    fastcgi_busy_buffers_size 128k;
    fastcgi_temp_file_write_size 128k;

    #gzip模块设置
    gzip on; #开启gzip压缩输出
    gzip_min_length 1k;    #最小压缩文件大小
    gzip_buffers 4 16k;    #压缩缓冲区
    gzip_http_version 1.0;    #压缩版本（默认1.1，前端如果是squid2.5请使用1.0）
    gzip_comp_level 2;    #压缩等级
    gzip_types text/plain application/x-javascript text/css application/xml;    #压缩类型，默认就已经包含textml，所以下面就不用再写了，写上去也不会有问题，但是会有一个warn。
    gzip_vary on;

    #开启限制IP连接数的时候需要使用
    #limit_zone crawler $binary_remote_addr 10m;



    #负载均衡配置
    upstream jh.w3cschool.cn {
     
        #upstream的负载均衡，weight是权重，可以根据机器配置定义权重。weigth参数表示权值，权值越高被分配到的几率越大。
        server 192.168.80.121:80 weight=3;
        server 192.168.80.122:80 weight=2;
        server 192.168.80.123:80 weight=3;

        #nginx的upstream目前支持4种方式的分配
        #1、轮询（默认）
        #每个请求按时间顺序逐一分配到不同的后端服务器，如果后端服务器down掉，能自动剔除。
        #2、weight
        #指定轮询几率，weight和访问比率成正比，用于后端服务器性能不均的情况。
        #例如：
        #upstream bakend {
        #    server 192.168.0.14 weight=10;
        #    server 192.168.0.15 weight=10;
        #}
        #2、ip_hash
        #每个请求按访问ip的hash结果分配，这样每个访客固定访问一个后端服务器，可以解决session的问题。
        #例如：
        #upstream bakend {
        #    ip_hash;
        #    server 192.168.0.14:88;
        #    server 192.168.0.15:80;
        #}
        #3、fair（第三方）
        #按后端服务器的响应时间来分配请求，响应时间短的优先分配。
        #upstream backend {
        #    server server1;
        #    server server2;
        #    fair;
        #}
        #4、url_hash（第三方）
        #按访问url的hash结果来分配请求，使每个url定向到同一个后端服务器，后端服务器为缓存时比较有效。
        #例：在upstream中加入hash语句，server语句中不能写入weight等其他的参数，hash_method是使用的hash算法
        #upstream backend {
        #    server squid1:3128;
        #    server squid2:3128;
        #    hash $request_uri;
        #    hash_method crc32;
        #}

        #tips:
        #upstream bakend{#定义负载均衡设备的Ip及设备状态}{
        #    ip_hash;
        #    server 127.0.0.1:9090 down;
        #    server 127.0.0.1:8080 weight=2;
        #    server 127.0.0.1:6060;
        #    server 127.0.0.1:7070 backup;
        #}
        #在需要使用负载均衡的server中增加 proxy_pass http://bakend/;

        #每个设备的状态设置为:
        #1.down表示单前的server暂时不参与负载
        #2.weight为weight越大，负载的权重就越大。
        #3.max_fails：允许请求失败的次数默认为1.当超过最大次数时，返回proxy_next_upstream模块定义的错误
        #4.fail_timeout:max_fails次失败后，暂停的时间。
        #5.backup： 其它所有的非backup机器down或者忙的时候，请求backup机器。所以这台机器压力会最轻。

        #nginx支持同时设置多组的负载均衡，用来给不用的server来使用。
        #client_body_in_file_only设置为On 可以讲client post过来的数据记录到文件中用来做debug
        #client_body_temp_path设置记录文件的目录 可以设置最多3层目录
        #location对URL进行匹配.可以进行重定向或者进行新的代理 负载均衡
    }
     
     
     
    #虚拟主机的配置
    server
    {
        #监听端口
        listen 80;

        #域名可以有多个，用空格隔开
        server_name www.w3cschool.cn w3cschool.cn;
        index index.html index.htm index.php;
        root /data/www/w3cschool;

        #对******进行负载均衡
        location ~ .*.(php|php5)?$
        {
            fastcgi_pass 127.0.0.1:9000;
            fastcgi_index index.php;
            include fastcgi.conf;
        }
         
        #图片缓存时间设置
        location ~ .*.(gif|jpg|jpeg|png|bmp|swf)$
        {
            expires 10d;
        }
         
        #JS和CSS缓存时间设置
        location ~ .*.(js|css)?$
        {
            expires 1h;
        }
         
        #日志格式设定
        #$remote_addr与$http_x_forwarded_for用以记录客户端的ip地址；
        #$remote_user：用来记录客户端用户名称；
        #$time_local： 用来记录访问时间与时区；
        #$request： 用来记录请求的url与http协议；
        #$status： 用来记录请求状态；成功是200，
        #$body_bytes_sent ：记录发送给客户端文件主体内容大小；
        #$http_referer：用来记录从那个页面链接访问过来的；
        #$http_user_agent：记录客户浏览器的相关信息；
        #通常web服务器放在反向代理的后面，这样就不能获取到客户的IP地址了，通过$remote_add拿到的IP地址是反向代理服务器的iP地址。反向代理服务器在转发请求的http头信息中，可以增加x_forwarded_for信息，用以记录原有客户端的IP地址和原来客户端的请求的服务器地址。
        log_format access '$remote_addr - $remote_user [$time_local] "$request" '
        '$status $body_bytes_sent "$http_referer" '
        '"$http_user_agent" $http_x_forwarded_for';
         
        #定义本虚拟主机的访问日志
        access_log  /usr/local/nginx/logs/host.access.log  main;
        access_log  /usr/local/nginx/logs/host.access.404.log  log404;
         
        #对 "/" 启用反向代理
        location / {
            proxy_pass http://127.0.0.1:88;
            proxy_redirect off;
            proxy_set_header X-Real-IP $remote_addr;
             
            #后端的Web服务器可以通过X-Forwarded-For获取用户真实IP
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
             
            #以下是一些反向代理的配置，可选。
            proxy_set_header Host $host;

            #允许客户端请求的最大单文件字节数
            client_max_body_size 10m;

            #缓冲区代理缓冲用户端请求的最大字节数，
            #如果把它设置为比较大的数值，例如256k，那么，无论使用firefox还是IE浏览器，来提交任意小于256k的图片，都很正常。如果注释该指令，使用默认的client_body_buffer_size设置，也就是操作系统页面大小的两倍，8k或者16k，问题就出现了。
            #无论使用firefox4.0还是IE8.0，提交一个比较大，200k左右的图片，都返回500 Internal Server Error错误
            client_body_buffer_size 128k;

            #表示使nginx阻止HTTP应答代码为400或者更高的应答。
            proxy_intercept_errors on;

            #后端服务器连接的超时时间_发起握手等候响应超时时间
            #nginx跟后端服务器连接超时时间(代理连接超时)
            proxy_connect_timeout 90;

            #后端服务器数据回传时间(代理发送超时)
            #后端服务器数据回传时间_就是在规定时间之内后端服务器必须传完所有的数据
            proxy_send_timeout 90;

            #连接成功后，后端服务器响应时间(代理接收超时)
            #连接成功后_等候后端服务器响应时间_其实已经进入后端的排队之中等候处理（也可以说是后端服务器处理请求的时间）
            proxy_read_timeout 90;

            #设置代理服务器（nginx）保存用户头信息的缓冲区大小
            #设置从被代理服务器读取的第一部分应答的缓冲区大小，通常情况下这部分应答中包含一个小的应答头，默认情况下这个值的大小为指令proxy_buffers中指定的一个缓冲区的大小，不过可以将其设置为更小
            proxy_buffer_size 4k;

            #proxy_buffers缓冲区，网页平均在32k以下的设置
            #设置用于读取应答（来自被代理服务器）的缓冲区数目和大小，默认情况也为分页大小，根据操作系统的不同可能是4k或者8k
            proxy_buffers 4 32k;

            #高负荷下缓冲大小（proxy_buffers*2）
            proxy_busy_buffers_size 64k;

            #设置在写入proxy_temp_path时数据的大小，预防一个工作进程在传递文件时阻塞太长
            #设定缓存文件夹大小，大于这个值，将从upstream服务器传
            proxy_temp_file_write_size 64k;
        }
         
         
        #设定查看Nginx状态的地址
        location /NginxStatus {
            stub_status on;
            access_log on;
            auth_basic "NginxStatus";
            auth_basic_user_file confpasswd;
            #htpasswd文件的内容可以用apache提供的htpasswd工具来产生。
        }
         
        #本地动静分离反向代理配置
        #所有jsp的页面均交由tomcat或resin处理
        location ~ .(jsp|jspx|do)?$ {
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_pass http://127.0.0.1:8080;
        }
         
        #所有静态文件由nginx直接读取不经过tomcat或resin
        location ~ .*.(htm|html|gif|jpg|jpeg|png|bmp|swf|ioc|rar|zip|txt|flv|mid|doc|ppt|
        pdf|xls|mp3|wma)$
        {
            expires 15d; 
        }
         
        location ~ .*.(js|css)?$
        {
            expires 1h;
        }
    }
}
######Nginx配置文件nginx.conf中文详解#####

```

### gcc 升级（centos）

```
查看当前gcc版本
[root@centos ~]# gcc -v 
```

``` 
安装gcc(gcc-10.1.0)
[root@centos install-package]# wget -P /home/common/install-package/ https://mirrors.aliyun.com/gnu/gcc/gcc-10.1.0/gcc-10.1.0.tar.gz

[root@centos install-package]# ls
gcc-10.1.0.tar.gz
[root@centos install-package]# tar -xvf gcc-10.1.0.tar.gz -C /opt
[root@centos install-package]# cd /opt/gcc-10.1.0
[root@centos gcc-10.1.0]# mkdir build/
[root@centos gcc-10.1.0]# cd build/
[root@centos build]#../configure --prefix=/opt/gcc-10.1.0/ --enable-checking=release --enable-languages=c,c++ --disable-multilib

```

```
从日志中可以看出有如下报错，故下面每个都安装
configure: error: Building GCC requires GMP 4.2+, MPFR 3.1.0+ and MPC 0.8.0+.

解决报错的问题 
1  安装gmp
[root@centos install-package]# wget -P /home/common/install-package/ https://gcc.gnu.org/pub/gcc/infrastructure/gmp-6.2.1.tar.bz2
[root@centos install-package]# tar -vxf gmp-6.2.1.tar.bz2 -C /opt
[root@centos install-package]# cd /opt/gmp-6.2.1
[root@centos gmp-6.2.1]# ./configure --prefix=/opt/gmp-6.2.1
......
checking whether sscanf needs writable input... no
checking for struct pst_processor.psp_iticksperclktick... no
......
[root@centos gmp-6.2.1]# make
[root@centos gmp-6.2.1]# make install

2  MPFR编译
[root@centos install-package]# wget -P /home/common/install-package/ https://gcc.gnu.org/pub/gcc/infrastructure/mpfr-4.1.0.tar.bz2
[root@centos install-package]# tar -vxf mpfr-4.1.0.tar.bz2 -C /opt

[root@centos install-package]# cd /opt/mpfr-4.1.0/
[root@centos mpfr-4.1.0]#./configure --prefix=/opt/mpfr-4.1.0 --with-gmp=/opt/gmp-6.2.1
[root@centos mpfr-4.1.0]# make
[root@centos mpfr-4.1.0]# make install

3  MPC编译
[root@centos install-package]# wget -P /home/common/install-package/ https://gcc.gnu.org/pub/gcc/infrastructure/mpc-1.2.1.tar.gz
[root@centos install-package]# tar -zvxf mpc-1.2.1.tar.gz -C /opt
[root@centos install-package]# cd /opt/mpc-1.2.1
[root@centos mpc-1.2.1]# ./configure --prefix=/opt/mpc-1.2.1 --with-gmp=/opt/gmp-6.2.1 --with-mpfr=/opt/mpfr-4.1.0
[root@centos mpc-1.2.1]# make
[root@centos mpc-1.2.1]# make install
```

![](C:\Users\Lenovo\Desktop\gcc_error.png)





![image-20220923134620033](C:\Users\Lenovo\AppData\Roaming\Typora\typora-user-images\image-20220923134620033.png)



```
gcc配置
[root@centos install-package]# cd /opt/gcc-10.1.0
[root@centos gcc-10.1.0]# cd build

[root@centos build]# ../configure --prefix=/opt/gcc-10.1.0/ --enable-checking=release --enable-languages=c,c++ --disable-multilib --with-gmp=/opt/gmp-6.2.1 --with-mpfr=/opt/mpfr-4.1.0 --with-mpc=/opt/mpc-1.2.1

# 编译  这里执行完make -j4 会报一个错误 见下面
[root@VM-16-13-centos build]# make -j4 # 时间很长很长 耐心等待 也可以使用make -j8
[root@VM-16-13-centos build]# make install

执行make -j4 时间会很长很长 大概1个半小时到2个小时之间的样子，而且执行完后会报一个下图的错误:
error while loading shared libraries: libmpfr.so.6: cannot open shared object file
```

```
解决错误 error while loading shared libraries: libmpfr.so.6: cannot open shared object file
[root@centos install-package]# wget -P /home/common/install-package/ https://distrib-coffee.ipsl.jussieu.fr/pub/linux/altlinux/p10/branch/x86_64/RPMS.classic/libmpfr6-4.1.0-alt1.x86_64.rpm
[root@centos install-package]# rpm2cpio libmpfr6-4.1.0-alt1.x86_64.rpm | cpio -div
[root@centos install-package]# rpm2cpio libmpfr6-4.1.0-alt1.x86_64.rpm | cpio -div
./usr/lib64/libmpfr.so.6
./usr/lib64/libmpfr.so.6.1.0
./usr/share/doc/mpfr-4.1.0
./usr/share/doc/mpfr-4.1.0/AUTHORS
./usr/share/doc/mpfr-4.1.0/BUGS
./usr/share/doc/mpfr-4.1.0/NEWS
5494 blocks

[root@centos install-package]# ls
libmpfr6-4.1.0-alt1.x86_64.rpm usr

[root@centos install-package]# mv  ./usr/lib64/libmpfr.so.6 /usr/lib64/
[root@centos install-package]# mv  ./usr/lib64/libmpfr.so.6.1.0 /usr/lib64/
[root@centos install-package]# cd /opt/gcc-10.1.0
[root@centos gcc-10.1.0]# cd build
[root@centos build]# make -j4 # 时间很长很长 耐心等待 也可以使用make -j8
[root@centos build]# make install
```

```
gcc 版本更新
[root@centos install-package]# mv /usr/bin/gcc /usr/bin/gcc485
[root@centos install-package]# mv /usr/bin/g++ /usr/bin/g++485
[root@centos install-package]# mv /usr/bin/c++ /usr/bin/c++485
[root@centos install-package]# mv /usr/bin/cc /usr/bin/cc485


[root@centos install-package]# ln -s /opt/gcc-10.1.0/bin/gcc /usr/bin/gcc
[root@centos install-package]# ln -s /opt/gcc-10.1.0/bin/g++ /usr/bin/g++
[root@centos install-package]# ln -s /opt/gcc-10.1.0/bin/c++ /usr/bin/c++
[root@centos install-package]# ln -s /opt/gcc-10.1.0/bin/gcc /usr/bin/cc


[root@centos install-package]# mv /usr/lib64/libstdc++.so.6 /usr/lib64/libstdc++.so.6.bak
[root@centos install-package]# ln -s /opt/gcc-10.1.0/lib64/libstdc++.so.6.0.28 /usr/lib64/libstdc++.so.6


脚本执行成功之后就可以查看当前使用的gcc版本了  查看的命令：gcc -v
[root@centos install-package]# gcc -v
Using built-in specs.
COLLECT_GCC=gcc
COLLECT_LTO_WRAPPER=/opt/gcc-10.1.0/libexec/gcc/x86_64-pc-linux-gnu/10.1.0/lto-wrapper
Target: x86_64-pc-linux-gnu
Configured with: ../configure --prefix=/opt/gcc-10.1.0/ --enable-checking=release --enable-languages=c,c++ --disable-multilib --with-gmp=/opt/gmp-6.2.1 --with-mpfr=/opt/mpfr-4.1.0 --with-mpc=/opt/mpc-1.2.1
Thread model: posix
Supported LTO compression algorithms: zlib
gcc version 10.1.0 (GCC)

```



## SQL

### 一、MySQL的数据类型

主要包括以下五大类：

整数类型：BIT、BOOL、TINY INT、SMALL INT、MEDIUM INT、 INT、 BIG INT

浮点数类型：FLOAT、DOUBLE、DECIMAL

字符串类型：CHAR、VARCHAR、TINY TEXT、TEXT、MEDIUM TEXT、LONGTEXT、TINY BLOB、BLOB、MEDIUM BLOB、LONG BLOB

日期类型：Date、DateTime、TimeStamp、Time、Year

其他数据类型：BINARY、VARBINARY、ENUM、SET、Geometry、Point、MultiPoint、LineString、MultiLineString、Polygon、GeometryCollection等

```
1、整型

MySQL数据类型	含义（有符号）
tinyint(m)	1个字节  范围(-128~127)
smallint(m)	2个字节  范围(-32768~32767)
mediumint(m)	3个字节  范围(-8388608~8388607)
int(m)	4个字节  范围(-2147483648~2147483647)
bigint(m)	8个字节  范围(+-9.22*10的18次方)
取值范围如果加了unsigned，则最大值翻倍，如tinyint unsigned的取值范围为(0~256)。
 int(m)里的m是表示SELECT查询结果集中的显示宽度，并不影响实际的取值范围，没有影响到显示的宽度，不知道这个m有什么用。
```

```
2、浮点型(float和double)

MySQL数据类型	含义
float(m,d)	单精度浮点型    8位精度(4字节)     m总个数，d小数位
double(m,d)	双精度浮点型    16位精度(8字节)    m总个数，d小数位
设一个字段定义为float(6,3)，如果插入一个数123.45678,实际数据库里存的是123.457，但总个数还以实际为准，即6位。整数部分最大是3位，如果插入数12.123456，存储的是12.1234，如果插入12.12，存储的是12.1200.
```

```
3、字符串(char,varchar,_text)

MySQL数据类型	含义
char(n)	固定长度，最多255个字符
varchar(n)	固定长度，最多65535个字符
tinytext	可变长度，最多255个字符
text	可变长度，最多65535个字符
mediumtext	可变长度，最多2的24次方-1个字符
longtext	可变长度，最多2的32次方-1个字符
char和varchar：
	1.char(n) 若存入字符数小于n，则以空格补于其后，查询之时再将空格去掉。所以char类型存储的字符串末尾不能有空格，varchar不限于此。 
	2.char(n) 固定长度，char(4)不管是存入几个字符，都将占用4个字节，varchar是存入的实际字符数+1个字节（n<=255）或2个字节(n>255)，所以varchar(4),存入3个字符将占用4个字节。 
	3.char类型的字符串检索速度要比varchar类型的快。
		varchar和text： 
			1.varchar可指定n，text不能指定，内部存储varchar是存入的实际字符数+1个字节（n<=255）或2个字节(n>255)，text是实际字符数+2个字节。 
			2.text类型不能有默认值。 
			3.varchar可直接创建索引，text创建索引要指定前多少个字符。varchar查询速度快于text,在都创建索引的情况下，text的索引似乎不起作用。
```

```
4、定点数
浮点型在数据库中存放的是近似值，而定点类型在数据库中存放的是精确值。 
decimal(m,d) 参数m<65 是总个数，d<30且 d<m 是小数位。
```

```
5.二进制数据(_Blob)
1._BLOB和_text存储方式不同，_TEXT以文本方式存储，英文存储区分大小写，而_Blob是以二进制方式存储，不分大小写。
2._BLOB存储的数据只能整体读出。 
3._TEXT可以指定字符集，_BLO不用指定字符集。
```

```
6.日期时间类型
MySQL数据类型	含义
date	日期 '2008-12-2'
time	时间 '12:25:36'
datetime	日期时间 '2008-12-2 22:06:44'
timestamp	自动存储记录修改时间
若定义一个字段为timestamp，这个字段里的时间数据会随其他字段修改的时候自动刷新，所以这个数据类型的字段可以存放这条记录最后被修改的时间。
```

```
数据类型的属性:
MySQL关键字		含义
NULL			数据列可包含NULL值
NOT NULL		数据列不允许包含NULL值
DEFAULT			默认值
PRIMARY KEY		主键
AUTO_INCREMENT	自动递增，适用于整数类型
UNSIGNED		无符号
CHARACTER SET name	指定一个字符集
```

##### 选择数据类型的基本原则

前提：使用适合存储引擎。

选择原则：根据选定的存储引擎，确定如何选择合适的数据类型。

下面的选择方法按存储引擎分类：

- MyISAM 数据存储引擎和数据列：MyISAM数据表，最好使用固定长度(CHAR)的数据列代替可变长度(VARCHAR)的数据列。
- MEMORY存储引擎和数据列：MEMORY数据表目前都使用固定长度的数据行存储，因此无论使用CHAR或VARCHAR列都没有关系。两者都是作为CHAR类型处理的。
- InnoDB 存储引擎和数据列：建议使用 VARCHAR类型。

对于InnoDB数据表，内部的行存储格式没有区分固定长度和可变长度列（所有数据行都使用指向数据列值的头指针），因此在本质上，使用固定长度的CHAR列不一定比使用可变长度VARCHAR列简单。因而，主要的性能因素是数据行使用的存储总量。由于CHAR平均占用的空间多于VARCHAR，因 此使用VARCHAR来最小化需要处理的数据行的存储总量和磁盘I/O是比较好的。

下面说一下固定长度数据列与可变长度的数据列。

###### char与varchar

CHAR和VARCHAR类型类似，但它们保存和检索的方式不同。它们的最大长度和是否尾部空格被保留等方面也不同。在存储或检索过程中不进行大小写转换。

下面的表显示了将各种字符串值保存到CHAR(4)和VARCHAR(4)列后的结果，说明了CHAR和VARCHAR之间的差别：

| 值         | CHAR(4) | 存储需求 | VARCHAR(4) | 存储需求                                                     |
| ---------- | ------- | -------- | ---------- | ------------------------------------------------------------ |
| ''         | '  '    | 4个字节  | ''         | 1个字节                                                      |
| 'ab'       | 'ab '   | 4个字节  | 'ab '      | 3个字节                                                      |
| 'abcd'     | 'abcd'  | 4个字节  | 'abcd'     | 5个字节                                                      |
| 'abcdefgh' | 'abcd'  | 4个字节  | 'abcd'     | 5个字节请注意上表中最后一行的值只适用*不使用严格模式*时；如果MySQL运行在严格模式，超过列长度不的值*不\*保存**，并且会出现错误。 |

从CHAR(4)和VARCHAR(4)列检索的值并不总是相同，因为检索时从CHAR列删除了尾部的空格。通过下面的例子说明该差别：
mysql> CREATE TABLE vc (v VARCHAR(4), c CHAR(4));
Query OK, 0 rows affected (0.02 sec)

mysql> INSERT INTO vc VALUES ('ab ', 'ab ');
Query OK, 1 row affected (0.00 sec)

mysql> SELECT CONCAT(v, '+'), CONCAT(c, '+') FROM vc;
+----------------+----------------+
| CONCAT(v, '+') | CONCAT(c, '+') |
| -------------- | -------------- |
|                |                |
+----------------+----------------+
| ab + | ab+  |
| ---- | ---- |
|      |      |
+----------------+----------------+
1 row in set (0.00 sec)



###### text和blob

在使用text和blob字段类型时要注意以下几点，以便更好的发挥数据库的性能。

①BLOB和TEXT值也会引起自己的一些问题，特别是执行了大量的删除或更新操作的时候。删除这种值会在数据表中留下很大的"空洞"，以后填入这些"空洞"的记录可能长度不同,为了提高性能,建议定期使用 OPTIMIZE TABLE 功能对这类表进行碎片整理.

②使用合成的（synthetic）索引。合成的索引列在某些时候是有用的。一种办法是根据其它的列的内容建立一个散列值，并把这个值存储在单独的数据列中。接下来你就可以通过检索散列值找到数据行了。但是，我们要注意这种技术只能用于精确匹配的查询（散列值对于类似<或>=等范围搜索操作符 是没有用处的）。我们可以使用MD5()函数生成散列值，也可以使用SHA1()或CRC32()，或者使用自己的应用程序逻辑来计算散列值。请记住数值型散列值可以很高效率地存储。同样，如果散列算法生成的字符串带有尾部空格，就不要把它们存储在CHAR或VARCHAR列中，它们会受到尾部空格去除的影响。

合成的散列索引对于那些BLOB或TEXT数据列特别有用。用散列标识符值查找的速度比搜索BLOB列本身的速度快很多。

③在不必要的时候避免检索大型的BLOB或TEXT值。例如，SELECT *查询就不是很好的想法，除非你能够确定作为约束条件的WHERE子句只会找到所需要的数据行。否则，你可能毫无目的地在网络上传输大量的值。这也是 BLOB或TEXT标识符信息存储在合成的索引列中对我们有所帮助的例子。你可以搜索索引列，决定那些需要的数据行，然后从合格的数据行中检索BLOB或 TEXT值。

④把BLOB或TEXT列分离到单独的表中。在某些环境中，如果把这些数据列移动到第二张数据表中，可以让你把原数据表中 的数据列转换为固定长度的数据行格式，那么它就是有意义的。这会减少主表中的碎片，使你得到固定长度数据行的性能优势。它还使你在主数据表上运行 SELECT *查询的时候不会通过网络传输大量的BLOB或TEXT值。

###### 浮点数与定点数

为了能够引起大家的重视，在介绍浮点数与定点数以前先让大家看一个例子：
mysql> CREATE TABLE test (c1 float(10,2),c2 decimal(10,2));
Query OK, 0 rows affected (0.29 sec)

mysql> insert into test values(131072.32,131072.32);
Query OK, 1 row affected (0.07 sec)

mysql> select * from test;
+-----------+-----------+
| c1    | c2    |
+-----------+-----------+
| 131072.31 | 131072.32 |
+-----------+-----------+
1 row in set (0.00 sec)

从上面的例子中我们看到c1列的值由131072.32变成了131072.31，这就是浮点数的不精确性造成的。

在mysql中float、double（或real）是浮点数，decimal（或numberic）是定点数。

浮点数相对于定点数的优点是在长度一定的情况下，浮点数能够表示更大的数据范围；它的缺点是会引起精度问题。在今后关于浮点数和定点数的应用中，大家要记住以下几点：

1. 浮点数存在误差问题；
2. 对货币等对精度敏感的数据，应该用定点数表示或存储；
3. 编程中，如果用到浮点数，要特别注意误差问题，并尽量避免做浮点数比较；
4. 要注意浮点数中一些特殊值的处理。

## GIT

```
name: gaozhan
psd:gao9051074ZHAN
```



## screen

下载：

```
CentOS:
	yum -y install screen
Ubuntu:
	apt-get -y install screen
```

常用操作：

```
参数说明:
-A 　将所有的视窗都调整为目前终端机的大小。
-d <作业名称> 　将指定的screen作业离线。
-h <行数> 　指定视窗的缓冲区行数。
-m 　即使目前已在作业中的screen作业，仍强制建立新的screen作业。
-r <作业名称> 　恢复离线的screen作业。
-R 　先试图恢复离线的作业。若找不到离线的作业，即建立新的screen作业。
-s 　指定建立新视窗时，所要执行的shell。
-S <作业名称> 　指定screen作业的名称。
-v 　显示版本信息。
-x 　恢复之前离线的screen作业。
-ls或--list 　显示目前所有的screen作业。
-wipe 　检查目前所有的screen作业，并删除已经无法使用的screen作业。
删除：screen -S name -X quit

常用screen参数:
screen -S yourname -> 新建一个叫yourname的session
screen -ls -> 列出当前所有的session
screen -r yourname -> 回到yourname这个session
screen -d yourname -> 远程detach某个session
screen -d -r yourname -> 结束当前session并回到yourname这个session

在每个screen session 下，所有命令都以 ctrl+a(C-a) 开始。
C-a ? -> 显示所有键绑定信息
C-a c -> 创建一个新的运行shell的窗口并切换到该窗口
C-a n -> Next，切换到下一个 window
C-a p -> Previous，切换到前一个 window
C-a 0..9 -> 切换到第 0..9 个 window
Ctrl+a [Space] -> 由视窗0循序切换到视窗9
C-a C-a -> 在两个最近使用的 window 间切换
C-a x -> 锁住当前的 window，需用用户密码解锁
C-a d -> detach，暂时离开当前session，将目前的 screen session (可能含有多个 windows) 丢到后台执行，并会回到还没进 screen 时的状态，此时在 screen session 里，每个 window 内运行的 process (无论是前台/后台)都在继续执行，即使 logout 也不影响。
C-a z -> 把当前session放到后台执行，用 shell 的 fg 命令则可回去。
C-a w -> 显示所有窗口列表
C-a t -> Time，显示当前时间，和系统的 load
C-a k -> kill window，强行关闭当前的 window
C-a [ -> 进入 copy mode，在 copy mode 下可以回滚、搜索、复制就像用使用 vi 一样
C-b Backward，PageUp
C-f Forward，PageDown
H(大写) High，将光标移至左上角
L Low，将光标移至左下角
0 移到行首
$ 行末
w forward one word，以字为单位往前移
b backward one word，以字为单位往后移
Space 第一次按为标记区起点，第二次按为终点
Esc 结束 copy mode
C-a ] -> Paste，把刚刚在 copy mode 选定的内容贴上
```



