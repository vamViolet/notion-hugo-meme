---
title: "关于ShardingSphere-JDBC的简单使用"
description: ""
date: 2021-08-31
image: ""
math: false
license:
comments: true
draft: false
build:
    list: always
tags : [java]
categories: 技术
lastmod: 2021-08-31
---
-   
    - 现状及目标
    - ShardingSphere-JDBC简介  
        - ShardingSphere-JDBC是什么
        - ShardingSphere - 简介
        - ShardingSphere-JDBC - 优势
        - Sharding-JDBC 分片算法
        - Sharding-JDBC 分片策略
    - ShardingSphere-JDBC - 使用案例

(目录)

## 现状及目标
> 由于旧的业务表中数据量会持续增长，且没有对数据做分表、索引，最终导致查询上亿数据时报错。因此决定使用ShardingSphere-JDBC对表数据做数据分片处理。本文主要介绍ShardingSphere-JDBC的主要功能、优势及用法。


## ShardingSphere-JDBC简介
### ShardingSphere-JDBC是什么
> ShardingSphere-JDBC作为Apache ShardingSphere的一个独立的产品。定位为轻量级 Java 框架，在 Java 的 JDBC 层提供的额外服务。 它使用客户端直连数据库，以 jar 包形式提供服务，无需额外部署和依赖，可理解为增强版的 JDBC 驱动，完全兼容 JDBC 和各种 ORM 框架。


### ShardingSphere - 简介
> ShardingSphere是一套开源的分布式数据库中间件解决方案组成的生态圈，它由Sharding-JDBC、Sharding-Proxy和Sharding-Sidecar（计划中）这3款相互独立的产品组成。 他们均提供标准化的数据分片、分布式事务和数据库治理功能，可适用于如Java同构、异构语言、云原生等各种多样化的应用场景。
 官网：[官网](https://shardingsphere.apache.org/index_zh.html)
 官方API：[sharding-jdbc手册](https://shardingsphere.apache.org/document/legacy/4.x/document/cn/manual/sharding-jdbc/)

### ShardingSphere-JDBC - 优势
> Sharding-JDBC的优势在于对Java应用的友好度。
 主要功能


> ![主要功能](/images/1771c908a63d46d99dd693237b222859.png)


区别
 

![对比](/images/9ba9f9a310174c1180f44488f0c55f3a.png)

### Sharding-JDBC 分片算法
通过分片算法将数据分片，支持通过=、>=、<=、>、<、BETWEEN和IN分片。分片算法需要应用方开发者自行实现，可实现的灵活度非常高。
 目前提供4种分片算法。

1.  **精确分片算法**
 对应PreciseShardingAlgorithm，用于处理使用单一键作为分片键的=与IN进行分片的场景。需要配合StandardShardingStrategy使用。 
1.  **范围分片算法**
 对应RangeShardingAlgorithm，用于处理使用单一键作为分片键的BETWEEN AND、>、<、>=、<=进行分片的场景。需要配合StandardShardingStrategy使用。 
1.  **复合分片算法**
 对应ComplexKeysShardingAlgorithm，用于处理使用多键作为分片键进行分片的场景，包含多个分片键的逻辑较复杂，需要应用开发者自行处理其中的复杂度。需要配合ComplexShardingStrategy使用。 
1.  **Hint分片算法**
 对应HintShardingAlgorithm，用于处理使用Hint行分片的场景。需要配合HintShardingStrategy使用。 

### Sharding-JDBC 分片策略
包含分片键和分片算法，由于分片算法的独立性，将其独立抽离。真正可用于分片操作的是分片键 + 分片算法，也就是分片策略。目前提供5种分片策略。

1.  **标准分片策略**
 对应StandardShardingStrategy。提供对SQL语句中的=, >, <, >=, <=, IN和BETWEEN AND的分片操作支持。StandardShardingStrategy只支持单分片键，提供PreciseShardingAlgorithm和RangeShardingAlgorithm两个分片算法。PreciseShardingAlgorithm是必选的，用于处理=和IN的分片。RangeShardingAlgorithm是可选的，用于处理BETWEEN AND, >, <, >=, <=分片，如果不配置RangeShardingAlgorithm，SQL中的BETWEEN AND将按照全库路由处理。 
1.  **复合分片策略**
 对应ComplexShardingStrategy。复合分片策略。提供对SQL语句中的=, >, <, >=, <=, IN和BETWEEN AND的分片操作支持。ComplexShardingStrategy支持多分片键，由于多分片键之间的关系复杂，因此并未进行过多的封装，而是直接将分片键值组合以及分片操作符透传至分片算法，完全由应用开发者实现，提供最大的灵活度。 
1.  **行表达式分片策略**
 对应InlineShardingStrategy。使用Groovy的表达式，提供对SQL语句中的=和IN的分片操作支持，只支持单分片键。对于简单的分片算法，可以通过简单的配置使用，从而避免繁琐的Java代码开发，如: t_user_$->{u_id % 8} 表示t_user表根据u_id模8，而分成8张表，表名称为t_user_0到t_user_7。 
1.  **Hint分片策略**
 对应HintShardingStrategy。通过Hint指定分片值而非从SQL中提取分片值的方式进行分片的策略。 
1.  **不分片策略**
 对应NoneShardingStrategy。不分片的策略。 

## ShardingSphere-JDBC - 使用案例
1. 引入Maven依赖

```xml
        <dependency>
            <groupId>org.apache.shardingsphere</groupId>
            <artifactId>sharding-jdbc-spring-boot-starter</artifactId>
            <version>${version.sharding-jdbc}</version>
        </dependency>
        <dependency>
            <groupId>org.apache.shardingsphere</groupId>
            <artifactId>sharding-jdbc-spring-namespace</artifactId>
            <version>${version.sharding-jdbc}</version>
        </dependency>
```
1. 基于springboot properties的配置文件

```xml
###############################    sharding-jdbc相关配置   #############################
#配置数据源
spring.shardingsphere.datasource.names =ds0
spring.shardingsphere.datasource.ds0.type =com.alibaba.druid.pool.DruidDataSource
spring.shardingsphere.datasource.ds0.url =jdbc:mysql://127.0.0.1:3311/sharding-test?serverTimezone=Asia/Shanghai&useSSL=false&useUnicode=true&characterEncoding=UTF-8&allowMultiQueries=true&autoReconnect=true&failOverReadOnly=false&useAffectedRows=true
spring.shardingsphere.datasource.ds0.username =ENC(s2SThtwAB3Cxlwc+2awSm6qe5CEpa9Fa)
spring.shardingsphere.datasource.ds0.password =ENC(Fu/RreOI95TySDuYErAG9ucleo9jY/Wv)
spring.shardingsphere.datasource.ds0.driver-class-name =com.mysql.jdbc.Driver
spring.shardingsphere.props.sql.show =true
#连接池初始化连接数
spring.shardingsphere.datasource.ds0.initial-size=5
#连接池最大连接数
spring.shardingsphere.datasource.ds0.max-active=30
#连接池最小连接数
spring.shardingsphere.datasource.ds0.min-idle=5
#获取连接时最大等待时间，单位毫秒
spring.shardingsphere.datasource.ds0.max-wait=60000
#配置间隔多久才进行一次检测，检测需要关闭的空闲连接，单位是毫秒
spring.shardingsphere.datasource.ds0.time-between-eviction-runs-millis=60000
#连接保持空闲而不被驱逐的最小时间
spring.shardingsphere.datasource.ds0.min-evictable-idle-time-millis=300000
#用来检测连接是否有效的sql，要求是一个查询语句
spring.shardingsphere.datasource.ds0.validation-query=SELECT 1 FROM DUAL
#建议配置为true，不影响性能，并且保证安全性。申请连接的时候检测，如果空闲时间大于timeBetweenEvictionRunsMillis，执行validationQuery检测连接是否有效。
spring.shardingsphere.datasource.ds0.test-while-idle=true
#申请连接时执行validationQuery检测连接是否有效，做了这个配置会降低性能。
spring.shardingsphere.datasource.ds0.test-on-borrow=false
#归还连接时执行validationQuery检测连接是否有效，做了这个配置会降低性能。
spring.shardingsphere.datasource.ds0.test-on-return=false
#是否缓存preparedStatement，也就是PSCache。PSCache对支持游标的数据库性能提升巨大，比如说oracle。在mysql下建议关闭。
spring.shardingsphere.datasource.ds0.pool-prepared-statements=true
#要启用PSCache，必须配置大于0，当大于0时，poolPreparedStatements自动触发修改为true。
spring.shardingsphere.datasource.ds0.max-pool-prepared-statement-per-connection-size=50
#配置监控统计拦截的filters，去掉后监控界面sql无法统计
spring.shardingsphere.datasource.ds0.filters=stat,wall
#通过connectProperties属性来打开mergeSql功能；慢SQL记录
spring.shardingsphere.datasource.ds0.connection-properties=druid.stat.mergeSql=true;druid.stat.slowSqlMillis=500
#合并多个DruidDataSource的监控数据
spring.shardingsphere.datasource.ds0.use-global-data-source-stat=true

spring.shardingsphere.sharding.tables.trip_ycym_record.actual-data-nodes=ds0.trip_ycym_record_$->{2020..2023}${(1..12).collect{t ->t.toString().padLeft(2,'0')}}
spring.shardingsphere.sharding.tables.trip_ycym_record.table-strategy.standard.sharding-column=paytime
#精确分片算法类名称，用于=和IN。该类需实现PreciseShardingAlgorithm接口并提供无参数的构造器
spring.shardingsphere.sharding.tables.trip_ycym_record.table-strategy.standard.precise-algorithm-class-name=com.vamdawn.config.sharding.TripPreciseShardingAlgorithm
#范围分片算法类名称，用于BETWEEN，可选。该类需实现RangeShardingAlgorithm接口并提供无参数的构造器
spring.shardingsphere.sharding.tables.trip_ycym_record.table-strategy.standard.range-algorithm-class-name=com.vamdawm.config.sharding.TripRangeShardingAlgorithm
```
1. PreciseShardingAlgorithm：用于保存数据时

```java
@Slf4j
public class TripPreciseShardingAlgorithm implements PreciseShardingAlgorithm<String> {

    @Override
    public String doSharding(Collection<String> availableTargetNames, PreciseShardingValue<String> shardingValue) {
        log.info("availableTargetNames : {}", availableTargetNames);
        log.info("shardingValue:{}", JSON.toJSONString(shardingValue));

        //获取paytime格式化后的年份月份(2108)
        Date paytime = DateUtil.strBeauty2Date(shardingValue.getValue());
        String payTimeFormat = DateUtil.date2LocalDateTime(paytime).format(DateTimeFormatter.ofPattern("yyyyMM"));
        for (String tableName : availableTargetNames) {
            String tableNameFormat = tableName.substring(tableName.lastIndexOf("_") + 1);
            if (ObjectUtils.nullSafeEquals(payTimeFormat, tableNameFormat)) {
                log.info("返回最终配置的表名====={}", tableName);
                return tableName;
            }
        }
        throw new IllegalArgumentException();
    }
}
```
1. RangeShardingAlgorithm：用于根据时间段查询数据时

```java
@Slf4j
public class TripRangeShardingAlgorithm implements RangeShardingAlgorithm<String> {

    private static final DateTimeFormatter FORMATTER = DateTimeFormatter.ofPattern("yyyyMM");
    private static final String TABLENAMEPREFIX = "trip_sharding_record_";
    private static final String TIMEFORMATTER = "yyyy-MM-dd HH:mm:ss";

    @Override
    public Collection<String> doSharding(Collection<String> availableTargetNames, RangeShardingValue<String> shardingValue) {
        log.info("目标表名集合availableTargetNames : {}", availableTargetNames);
        log.info("shardingValue:{}", JSON.toJSONString(shardingValue));

        //最终要查询的实际表名集合
        Collection<String> actualTableNameList = new LinkedHashSet(availableTargetNames.size());
        //初始化上线日期
        String initDateStr = "2021-05-01 00:00:00";
        Date initDate = DateUtil.strBeauty2Date(initDateStr);
        Range<String> valueRange = shardingValue.getValueRange();
        log.info("日期范围:{}", valueRange);
        LocalDateTime initDateTime = DateUtil.date2LocalDateTime(initDate);
        String lowerDateStr = valueRange.lowerEndpoint();
        LocalDateTime lowerDateTime = LocalDateTimeUtil.parse(lowerDateStr, TIMEFORMATTER);
        //设置最早开始查询时间为上线时间
        if (lowerDateTime.isBefore(initDateTime)) {
            lowerDateTime = initDateTime;
        }
        String upperDateStr = valueRange.upperEndpoint();
        LocalDateTime upperDateTime = LocalDateTimeUtil.parse(upperDateStr, TIMEFORMATTER);
        //获取到相差的月份，计算出之间每个月份表
        long intervalMonth = lowerDateTime.until(upperDateTime, ChronoUnit.MONTHS);
        log.info("获取到相差的月份:{}", intervalMonth);
        String tableNameMonth = TABLENAMEPREFIX + lowerDateTime.format(FORMATTER);
        for (long i = 0; i <= intervalMonth; i++) {
            String monthFormat = lowerDateTime.plusMonths(i).format(FORMATTER);
            tableNameMonth = TABLENAMEPREFIX + monthFormat;
            actualTableNameList.add(tableNameMonth);
        }
        log.info("最终要查询的实际表名集合为:{}", actualTableNameList);
        return actualTableNameList;
    }
}
```

