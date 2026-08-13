---
title: "关于Jayspt加密明文敏感信息的使用"
description: ""
date: 2021-09-06
image: ""
math: false
license:
comments: true
draft: false
build:
    list: always
tags : [java]
categories: 技术
lastmod: 2021-09-06
---
TOC

## 现状及目标
> 项目中会有配置文件来存放敏感信息（比如数据库密码、中间件密码等），这些明文存储的敏感信息一旦泄露，会引发严重的安全事故。为了消除安全隐患，最直接的方式就是把明文敏感信息加密，在程序需要用到的时候进行解密。针对加密、解密，Jasypt 框架提供了很好的解决方案。


## Jasypt - 简介
### Jasypt - 是什么
> JASYPT: Java Simplified Encryption 是一个Java类库，它允许开发人员以很简单的方式添加基本加密功能，且无需深入研究加密原理。


### Jasypt - 简介
> Jasypt是基于标准的加密技术所做的加强，不仅简化了加密和检查密码的操作，还具有高度的可配置性。开发时可以自主选择不同的加密算法、salt生成等高级功能。
 GitHub地址：[GitHub项目地址](https://github.com/ulisesbocchio/jasypt-spring-boot)
 官网：[jasypt官网](http://www.jasypt.org/index.html)

### Jasypt - 特点
> Jasypt 具有安全性高、线程安全、配置性高、跨语言平台等优点。具体内容见官网。
 主要特点：[官网主要特点](http://www.jasypt.org/features.html)

### Jasypt - 原理
> 后续补充。


### Jasypt - 集成方式
Jasypt Spring Boot 为 Spring Boot 应用程序中的属性源提供加密支持。 有 3 种方法可以将 jasypt-spring-boot 集成到您的项目中：

1. 使用 @SpringBootApplication 或 @EnableAutoConfiguration 将在整个 Spring 环境中启用可加密属性，只需将 starter jar jasypt-spring-boot-starter 添加到您的类路径中
1. 将 jasypt-spring-boot 添加到您的类路径并将 @EnableEncryptableProperties 添加到您的主配置类以在整个 Spring 环境中启用可加密的属性
1. 将 jasypt-spring-boot 添加到您的类路径并使用 @EncrytablePropertySource 声明各个可加密的属性源

## Jasypt - starter方式使用案例
1. 引入Maven依赖

```xml
            <dependency>
                <groupId>com.github.ulisesbocchio</groupId>
                <artifactId>jasypt-spring-boot-starter</artifactId>
                <version>${jasypt-spring-boot}</version>
            </dependency>
```
1. properties 配置文件修改

```xml
spring.shardingsphere.datasource.ds0.username =ENC(Pvpo02TGnqgo0KCdy2KhYssSZ9ZAvzzZ)
spring.shardingsphere.datasource.ds0.password =ENC(fxXySsO/WECGD6AjNs1wDI7lWZDLdgjw)
```
1. salt设置

```xml
jasypt.encryptor.password= xxx
```
1. 启动类启用配置

```java
@Configuration
@EnableEncryptableProperties
public class MyApplication {
    ...
}
```
1. 测试类测试

```java
    @Test
    public void jasyptTest() {
        BasicTextEncryptor encryptor = new BasicTextEncryptor();
        encryptor.setPassword("G3CvKz6pLd9");
        String encrypted = encryptor.encrypt("testuser");
        //加密后的数据：Pvpo02TGnqgo0KCdy2KhYssSZ9ZAvzzZ
        System.out.println(encrypted);
        String decrypt = encryptor.decrypt("s2SThtwAB3Cxlwc+2awSm6qe5CEpa9Fa");
        //解密后的数据：testuser
        System.out.println(decrypt);
    }
```

