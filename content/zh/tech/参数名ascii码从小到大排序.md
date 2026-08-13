---
title: "参数名ASCII码从小到大排序"
description: ""
date: 2023-04-11
image: ""
math: false
license:
comments: false
draft: false
build:
    list: always
tags : [ASCII码, 签名/验签, java]
categories: 技术
lastmod: 2023-04-11
---
TOC

### 背景介绍
签名/验签时，经常需要将参数名按照ASCII码从小到大排序后，然后进行加密/解密操作。这里介绍三种ASCII码从小到大排序方法。

1. Guava [1]
1. StringBuffer
1. hutool工具类

### 测试方法
```
下面的测试方法都是以TreeMap存储参数后，做的排序。因为**TreeMap可以自动排序**。需要注意，Guava的字符串拼接类Joiner并没有做排序，如果用HashMap，需要先使用Arrays.sort排序。
```
```java
    public static void main(String[] args) {
        //签名/验签时，将参数名ASCII码从小到大排序
        Map<String, Object> map = new TreeMap<String,Object>();
        map.put("logicId", "logicId()");
        map.put("cardType", "tripRecentForm.getCardType()");
        map.put("traTime", "tripRecentForm.getTraTime()");
        map.put("deviceId", "tripRecentForm.getDeviceId()");
        map.put("traStation", "tripRecentForm.getTraStation()");
        map.put("version", "tripRecentForm.getVersion()");

        //1.Guava
        String content1 = Joiner.on("&").withKeyValueSeparator("=").join(map);
        System.out.println("content1:"+content1);

        //2.StringBuffer
        StringBuffer stringBuffer = new StringBuffer();
        map.forEach((key, value) -> stringBuffer.append(key).append("=").append(value).append("&"));
        String content2 = stringBuffer.deleteCharAt(stringBuffer.length()-1).toString();
        System.out.println("content2:"+content2);

        //3.hutool工具类
        String content3 = MapUtil.sortJoin(map, "&", "=", true, null);
        System.out.println("content3:"+content3);
    }
```
### 对比结果
```java
d:\zen_v1.0\jdk-8u201\bin\java.exe 
content1:cardType=tripRecentForm.getCardType()&deviceId=tripRecentForm.getDeviceId()&logicId=logicId()&traStation=tripRecentForm.getTraStation()&traTime=tripRecentForm.getTraTime()&version=tripRecentForm.getVersion()
content2:cardType=tripRecentForm.getCardType()&deviceId=tripRecentForm.getDeviceId()&logicId=logicId()&traStation=tripRecentForm.getTraStation()&traTime=tripRecentForm.getTraTime()&version=tripRecentForm.getVersion()
content3:cardType=tripRecentForm.getCardType()&deviceId=tripRecentForm.getDeviceId()&logicId=logicId()&traStation=tripRecentForm.getTraStation()&traTime=tripRecentForm.getTraTime()&version=tripRecentForm.getVersion()

Process finished with exit code 0
```
[1] Guava工程包含了若干被Google的 Java项目广泛依赖 的核心库，提供了一些常用的便利的操作工具类，减少因为 空指针、异步操作等引起的问题BUG，提高开发效率。例如：集合 、缓存、原生类型支持 、并发库  、通用注解 、字符串处理 、I/O 等等。


