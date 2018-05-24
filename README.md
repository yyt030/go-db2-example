# go-db2-example
ibm db2是款商用的db，默认支持的编程语言中没有Go，也没有介绍过，今天就演示下Go通过odbc方式连接ibm db2的例子。

## DB2 ODBC driver
### 安装DB2 ODBC driver
DB2 ODBC driver的来源有一下几种：
- db2安装包自带odbc驱动，和常用的jdbc驱动一样，odbc驱动一般都是在安装包中自带。
- 此外，一些单独的db2套件也含有odbc驱动，如：DB2 Application Development Client, Db2 Run-time Client, Db2 Data Server Runtime client等等
- 单独的免安装odbc驱动包：db2 driver for ODBC Cli

> 其中，db2 cli（db2 Call Level Interface）是一个和db2交互的SQL接口，基于ODBC规范和SQL / CLI国际标准。

下载macos64_odbc_cli.tar.gz，解压到任意目录下；

### 创建Data Source配置
db2 有两种data source配置文件，具体路径在<install-odbc-dir>/cfg下面
- db2cli.ini       # text格式，该文件可用于所有的odbc驱动
- db2dsdriver.cfg  # xml格式， 9.5 版本后引入的，可以公DB2 Data Server Driver for ODBC and CLI使用

db2cli.ini范例
```
; db2cli.ini data source
[DB2_SAMPLE]
Database=SAMPLE
Protocol=TCPIP
Port=50000
Hostname=my-db2-machine
UID=my-OS-user
PWD=my-OS-password
```

db2dsdriver.cfg范例
```
<parameter name="name" value="value"/>
<!--  db2dsdriver.cfg data source -->
<configuration>
   <dsncollection>
      <dsn alias="DB2_SAMPLE" name="SAMPLE" host="my-db2-machine" port="50000">
         <parameter name="UserID" value="my-db2-user"/>
         <parameter name="Password" value="my-db2-password"/>
      </dsn>
   </dsncollection>
</configuration>
```
> 比较简单，不做过多介绍，cfg下也有example。注意，配置文件中存在相同的dsn时，优先加载db2cli.ini文件；

### 配置完成后，验证下文件是否正确
```bash
$ cd /home/myuser/db2/odbc_cli/clidriver/bin/
$ ./db2cli validate -dsn DB2_SAMPLE
 db2cli validate -dsn sample

===============================================================================
Client information for the current copy:
===============================================================================

Client Package Type       : IBM Data Server Driver For ODBC and CLI
Client Version (level/bit): DB2 v10.5.0.5 (special_35187/64-bit)
Client Platform           : Darwin
Install/Instance Path     : .../clidriver
DB2DSDRIVER_CFG_PATH value: <not-set>
db2dsdriver.cfg Path      : .../clidriver/cfg/db2dsdriver.cfg
DB2CLIINIPATH value       : <not-set>
db2cli.ini Path           : .../clidriver/cfg/db2cli.ini
db2diag.log Path          : .../clidriver/db2dump/db2diag.log

===============================================================================
db2dsdriver.cfg schema validation for the entire file:
===============================================================================

Note: The validation utility could not find the configuration file
db2dsdriver.cfg. The file is searched at
".../clidriver/cfg/db2dsdriver.cfg".

===============================================================================
db2cli.ini validation for data source name "sample":
===============================================================================

[ Keywords used for the connection ]

Keyword                   Value
---------------------------------------------------------------------------
DATABASE                  sample
PROTOCOL                  TCPIP
HOSTNAME                  127.0.0.1
SERVICENAME               50000
UID                       db2inst1
PWD                       ********
```

### Test下连接
```bash
$ echo "select count(1) from syscat.tables" |db2cli execsql -dsn sample [ -user *** -passwd *** ]
IBM DATABASE 2 Interactive CLI Sample Program
(C) COPYRIGHT International Business Machines Corp. 1993,1996
All Rights Reserved
Licensed Materials - Property of IBM
US Government Users Restricted Rights - Use, duplication or
disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
> select count(1) from syscat.tables
FetchAll:  Columns: 1
  1
  643
FetchAll: 1 rows fetched.
```
> 其中，用户名和密码可省略，配置ini文件中；

## unixODBC
首先要安装[unixODBC](http://www.unixodbc.org/)驱动管理器，该odbc管理器是开源项目，能够帮助非win平台下的用户轻松使用odbc访问目标数据库.

与其它连接方式不同的是，ODBC应用程序app通常加载链接到ODBC驱动程序管理器而不是特定的ODBC驱动程序。
ODBC驱动程序管理器是ODBC应用程序和ODBC驱动程序之间的接口和桥梁。

在运行时，应用程序提供了一个连接字符串，即DSN，该连接字符串定义了要连接的ODBC数据源，并依次定义将处理连接的ODBC驱动程序。
unixODBC加载所请求的ODBC驱动程序并将所有ODBC API调用传递给目标驱动程序，也就是db2 odbc driver;
流程如下：
app ---> unixODBC ---> db2 ODBC driver

对于DB2 ODBC驱动程序来说，ODBC应用程序需要提供一个与DB2 ODBC驱动程序数据源同名的ODBC数据源。
unixODBC加载相应的数据源驱动程序(DB2 ODBC driver)，并将数据源配置信息传递给加载的DB2 ODBC driver,DB2 ODBC驱动程序检查它的数据源的配置文件，判断它的名称与它传递的名称相同;

### 安装unixODBC
```bash
brew install unixODBC
```

> 其它平台常见unixODBC官方网站即可；

### 注册DB2 ODBC driver和数据源
查看odbc配置文件
```
$ odbcinst -j
unixODBC 2.3.6
DRIVERS............: /usr/local/etc/odbcinst.ini
SYSTEM DATA SOURCES: /usr/local/etc/odbc.ini
FILE DATA SOURCES..: /usr/local/etc/ODBCDataSources
USER DATA SOURCES..: .../.odbc.ini
SQLULEN Size.......: 8
SQLLEN Size........: 8
SQLSETPOSIROW Size.: 8
```

其中，odbcinst.ini配置范例如下：
```
# Example odbcinst.ini driver definition for DB2 Data Server Driver for ODBC and
# CLI
[DB2]
Description=DB2 ODBC Driver
Driver=/usr/lib/libdb2.so # Replace /usr/lib with the directory where your
                          # driver shared object is located.
Fileusage=1               #
Dontdlclose=1             # IBM recommend setting Dontdlclose to 1, which stops
                          # unixODBC unloading the ODBC Driver on disconnect.
                          # Note that in unixODBC 2.2.11 and later, the Driver
                          # Manager default for Dontdlclose is 1.
```
> 上面已经提到过，为了让应用程序通过odbc方式连接db，unixODBC管理器需要知道要加载的驱动在哪里，同时需要指定odbc驱动的连接参数，如：ip,user等；
要应用这些参数，就要配置相关的DSN 参数连接；

~/.odbc.ini配置如下：
```
[DB2_SAMPLE]
Driver=DB2
```
> 故此，当用户程序需要连接DB2_SAMPLE时，
    - 首先，unixODBC会加载DB2 ODBC driver驱动，
    - 之后，DB2 ODBC driver会在db2cli.ini/db2dsdriver.cfg找相同名字的dns的相关配置；
    - 若用户程序提供的用户名和密码，配置文件中的用户名和密码会忽略。

如下isql是unixODBC自带的ODBC程序，通过它可以验证dsn的配置是否正确,连接是否ok；
```bash
$ isql -v DB2_SAMPLE username password
错误1：说明~/.odbc.ini配置文件中的Data Source Name（DSN）没有找到；
[IM002][unixODBC][Driver Manager]Data source name not found, and no default driver specified
[ISQL]ERROR: Could not SQLConnect

$ isql -v DB2_SAMPLE username password
错误2：说明DB2 ODBC driver配置文件没有找到匹配的DSN名字的配置信息
[     ][unixODBC][IBM][CLI Driver] SQL1531N  The connection failed because the name specified
with the DSN connection string keyword could not be found in either the db2dsdriver.cfg configuration file or the db2cli.ini configuration file.
Data source name specified in the connection string: "DB2_SAMPLE".

$ isql -v DB2_SAMPLE
+---------------------------------------+
| Connected!                            |
|                                       |
| sql-statement                         |
| help [tablename]                      |
| quit                                  |
|                                       |
+---------------------------------------+
SQL>
```

## 使用go连接db2
见源码


PS：对于win下的go，抽空再尝试。