---
title: "skyWalking自动建表-逻辑梳理(SkyWalking automatic table creation - logical sorting)"
description: ""
date: 2023-04-14
image: ""
math: false
license:
comments: false
draft: false
build:
    list: always
tags : [skywalking, java]
categories: 技术
lastmod: 2023-04-14
---
# skyWalking自动建表-逻辑梳理
> 使用skyWalking后，发现我们不需要创建表，启动skywalking会自动创建表，遂研究官方源码，感觉oap-server设计的自动建表功能很强大，并进行逻辑梳理，仅供参考。


> 源码地址：https://github.com/apache/skywalking.git


## 架构图
![架构图](/images/21004_BB80B109DFA242108A20A5B0D3C8CF91)

- Agent（代理/探针） ：负责从应用中，无侵入式的收集，并通过HTTP或者gRPC方式发送数据到SkyWalking OAP 服务器；
- SkyWalking OAP ：负责接收Agent发送的Tracing数据信息，然后进行分析(Analysis Core)，存储到外部存储器(Storage)，最终提供查询(Query)功能；
- Storage：Tracing数据存储，目前支持ES、MySQL、Sharding Sphere、TiDB、H2等多种存储器，SkyWalking开发团队自己的生产环境采用ES为主；
- SkyWalking UI：负责提供控制台，查看链路等等；

## 建表逻辑
1、启动时执行脚本
 

![1.启动脚本](/images/21004_0441844F58B8448CBD663ACA3EDEB866)

2、脚本内容指向：org.apache.skywalking.oap.server.starter.OAPServerStartUp
 

![2.脚本内容](/images/21004_823CA8E3885247ACB1CE72F99EE0250B)

3、进行初始化操作
 

![3.初始化操作](/images/21004_0907217894CC46C1B6C3C1FE999034C5)

4、moduleManager#init方法调用BootstrapFlow#start方法，BootstrapFlow#start调用ModuleProvider#start方法
 

![4.bootstrapflow-start](/images/21004_A0E2F70F3AFE41BC885F3B51EABE94FE)

5、抽象类ModuleProvider#start方法实现类JDBCStorageProvider#start方法，此方法开始调用 modelInstaller#start方法用于重写列，使语法兼容MySQL，然后调用StorageModels#addModelListener方法用于把创建或修改表的操作通知到ModelInstaller#whenCreating
 

![5.listener-creating](/images/21004_5751FE29D8EE4D108666372353E5824C)

6、收到通知后，ModelInstaller#whenCreating开始建表操作
 

![6.when-creating](/images/21004_6C8DC72785E3442AB54F0B44A8BD1C21)

7、方法内部依次进行 1.创建/更新 表字段；2.添加/更新 字段索引；3.创建/更新 关联表的字段及索引
 

![7.create-table](/images/21004_E5214527BC9947FB9C349B439542124F)

8、为表添加ID字段、普通字段；添加字段类型及长度；执行建表SQL
 

![8.execute-sql](/images/21004_67844556EC704862B8B3DE0A30AD8B1A)

9、添加字段类型，进行字段长度转换
 

![9.column-definition](/images/21004_4B0BAC51DFE6456E94C9F80C79E751C8)

p{text-indent:2em}


