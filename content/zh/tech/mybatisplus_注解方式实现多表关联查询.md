---
title: "MyBatisPlus 注解方式实现多表关联查询"
description: ""
date: 2021-06-16
image: ""
math: false
license:
comments: false
draft: false
build:
    list: always
tags : [mybatis]
categories: 技术
lastmod: 2021-06-16
---
[toc]

## 官方API
[官方文档](https://mp.baomidou.com/guide/)：https://mp.baomidou.com/guide/
 [具体位置](https://mp.baomidou.com/guide/wrapper.html#%E7%94%A8%E6%B3%A8%E8%A7%A3)：https://mp.baomidou.com/guide/wrapper.html#%E7%94%A8%E6%B3%A8%E8%A7%A3

## 使用概述
> Mapper层方法上添加 ${ew.customSqlSegment}和@Param(Constants.WRAPPER)；
 查询vo添加对应的查询条件字段，结果vo添加所想要展示的字段；
 service层方法中构造相应的QueryWrapper。
 即可实现多表联查、动态条件查询。


## 具体写法
### Mapper写法：
```java
@Select("SELECT * FROM tableA a LEFT JOIN tableB b on a.key = b.key ${ew.customSqlSegment}")
List method1(@Param(Constants.WRAPPER) QueryWrapper wrapper);
IPage method2(Page<> page, @Param(Constants.WRAPPER) QueryWrapper wrapper);
```
**需要注意：wrapper不能为null，可以用new QueryWrapper<>();**

### entity写法：
> 查询model中，如果既有A表参数，又有B表参数，需要在entity中添加字段
 返回结果vo中，和A、B表对应上的字段都会自动赋值


### service写法：
> 封装wrapper时，column字段最好写明表名。例：wrapper.eq(StringUtils.isNotBlank("xxx"), "A.column","value");


```java
@Override
    public void getRecord() {

        //返回值为list
        QueryWrapper<PassRecord> wrapper = new QueryWrapper<>();
        wrapper.eq(StringUtils.isNotBlank(""), "user_name","aaa");
        wrapper.eq("p.card_id","44520xxx");
        DateTimeFormatter dateTimeFormatter = DateTimeFormatter.ofPattern("yyyyMMdd HH:mm:ss");
        LocalDateTime dateTime = LocalDateTime.parse("20210611 18:04:00", dateTimeFormatter);
        Date startDate = Date.from(dateTime.atZone(ZoneId.systemDefault()).toInstant());
        LocalDateTime dateTime2 = LocalDateTime.parse("20210615 18:04:00", dateTimeFormatter);
        Date endDate = Date.from(dateTime2.atZone(ZoneId.systemDefault()).toInstant());
        wrapper.between("p.create_time",startDate, endDate);
        List<PassRecord> list = passRecordMapper.getRecordParam(wrapper);
        list.forEach(passRecord -> {
            System.out.println(passRecord.getUserName() +"=="+ passRecord.getWorkPlace());
        });

        //返回page对象
        Page<PassInfoDto> page = new Page<>();
        page.setCurrent(1L);
        page.setSize(2L);
        QueryWrapper<PassRecord> queryWrapper = new QueryWrapper<>();
        List<PassRecord> recordList = passRecordMapper.getRecordPageParam(page, queryWrapper);
        recordList.forEach(record -> {
            System.out.println(record.getUserName() + "" + record.getRecordId());
        });

        IPage<PassRecord> iPage = passRecordMapper.getRecordParamToPage(page, queryWrapper);
        System.out.println(JSON.toJSONString(iPage));
    }
```
### 日志报错
**wrapper不能为null**

```java
Creating a new SqlSession
SqlSession [org.apache.ibatis.session.defaults.DefaultSqlSession@5e207103] was not registered for synchronization because synchronization is not active
Closing non transactional SqlSession [org.apache.ibatis.session.defaults.DefaultSqlSession@5e207103]
2021-06-16 08:57:19.297 | ERROR |  | http-nio-8080-exec-1 | c.g.v.c.e.ExceptionCatch:160:exception - catch Exception:org.mybatis.spring.MyBatisSystemException: nested exception is org.apache.ibatis.builder.BuilderException: Error evaluating expression 'ew.customSqlSegment'. Cause: org.apache.ibatis.ognl.OgnlException: source is null for getProperty(null, "customSqlSegment")
	at org.mybatis.spring.MyBatisExceptionTranslator.translateExceptionIfPossible(MyBatisExceptionTranslator.java:92)
	at org.mybatis.spring.SqlSessionTemplate$SqlSessionInterceptor.invoke(SqlSessionTemplate.java:440)
	at com.sun.proxy.$Proxy110.selectList(Unknown Source)
	at org.mybatis.spring.SqlSessionTemplate.selectList(SqlSessionTemplate.java:223)
	at com.baomidou.mybatisplus.core.override.MybatisMapperMethod.executeForMany(MybatisMapperMethod.java:173)
	at com.baomidou.mybatisplus.core.override.MybatisMapperMethod.execute(MybatisMapperMethod.java:78)
	at com.baomidou.mybatisplus.core.override.MybatisMapperProxy$PlainMethodInvoker.invoke(MybatisMapperProxy.java:148)
	at com.baomidou.mybatisplus.core.override.MybatisMapperProxy.invoke(MybatisMapperProxy.java:89)
	at com.sun.proxy.$Proxy121.getRecordPageParam(Unknown Source)
	at com.grg.virusControl.mgr.service.impl.PassMgrServiceImpl.getRecord(PassMgrServiceImpl.java:316)
	at com.grg.virusControl.mgr.service.impl.PassMgrServiceImpl$$FastClassBySpringCGLIB$$a74fa41e.invoke(<generated>)
	at org.springframework.cglib.proxy.MethodProxy.invoke(MethodProxy.java:218)
	at org.springframework.aop.framework.CglibAopProxy$DynamicAdvisedInterceptor.intercept(CglibAopProxy.java:685)
	at com.grg.virusControl.mgr.service.impl.PassMgrServiceImpl$$EnhancerBySpringCGLIB$$63af0f69.getRecord(<generated>)
	at com.grg.virusControl.mgr.controller.PassMgrController.getRecord(PassMgrController.java:59)
	at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
	at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
	at java.lang.reflect.Method.invoke(Method.java:498)
	at org.springframework.web.method.support.InvocableHandlerMethod.doInvoke(InvocableHandlerMethod.java:190)
	at org.springframework.web.method.support.InvocableHandlerMethod.invokeForRequest(InvocableHandlerMethod.java:138)
	at org.springframework.web.servlet.mvc.method.annotation.ServletInvocableHandlerMethod.invokeAndHandle(ServletInvocableHandlerMethod.java:105)
	at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.invokeHandlerMethod(RequestMappingHandlerAdapter.java:893)
	at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.handleInternal(RequestMappingHandlerAdapter.java:798)
	at org.springframework.web.servlet.mvc.method.AbstractHandlerMethodAdapter.handle(AbstractHandlerMethodAdapter.java:87)
	at org.springframework.web.servlet.DispatcherServlet.doDispatch(DispatcherServlet.java:1040)
	at org.springframework.web.servlet.DispatcherServlet.doService(DispatcherServlet.java:943)
	at org.springframework.web.servlet.FrameworkServlet.processRequest(FrameworkServlet.java:1006)
	at org.springframework.web.servlet.FrameworkServlet.doGet(FrameworkServlet.java:898)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:634)
	at org.springframework.web.servlet.FrameworkServlet.service(FrameworkServlet.java:883)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:741)
	at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:231)
	at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:166)
	at org.apache.tomcat.websocket.server.WsFilter.doFilter(WsFilter.java:53)
```
### 结果截图
![](/post-images/1623814772810.png)

![](/post-images/1623814827312.png)


