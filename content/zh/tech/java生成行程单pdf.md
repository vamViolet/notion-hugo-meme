---
title: "JAVA生成行程单PDF"
description: ""
date: 2023-04-11
image: ""
math: false
license:
comments: false
draft: false
build:
    list: always
tags : [PDF, java]
categories: 技术
lastmod: 2023-04-11
---
# JAVA generated itinerary PDF
## 一、pom依赖
首先引入PDF需要的pom依赖

```java
         <!-- pdf-->
        <dependency>
            <groupId>com.itextpdf</groupId>
            <artifactId>itextpdf</artifactId>
            <version>5.5.13</version>
            <optional>true</optional>
        </dependency>
        <dependency>
            <groupId>com.itextpdf</groupId>
            <artifactId>itext-asian</artifactId>
            <version>5.2.0</version>
            <optional>true</optional>
        </dependency>
```
## 二、PDF操作工具类及行程单实体类
创建PDF操作工具类及行程单实体类，调用 PDFUtil.createPDFFile()方法创建行程单。

```java
 /**
 * PDF操作工具类
 */
public class PDFUtil {
    private static Logger logger = LoggerFactory.getLogger(PDFUtil.class);
    public static final String PDF_FILE_SUFFIX = ".pdf";
    public static Font headFont;
    public static Font keyFont;
    // 标题
    public static Font titleFont;
    // 设置字体大小
    public static Font textFont;
    public static final int MAX_WIDTH = 520;
    static {
        //中文格式
        BaseFont bfChinese;
        try {
            // 设置中文显示
            bfChinese = BaseFont.createFont("STSong-Light", "UniGB-UCS2-H", BaseFont.NOT_EMBEDDED);
            headFont = new Font(bfChinese, 10, java.awt.Font.BOLD);
            keyFont = new Font(bfChinese, 8, java.awt.Font.BOLD);
            textFont = new Font(bfChinese, 8, Font.NORMAL);
            titleFont = new Font(bfChinese, 16, Font.NORMAL);
        } catch (Exception e) {
            logger.error("创建pdf格式失败，异常={}", e);
            throw new RuntimeException(e);
        }
    }
    /**
     * 为表格添加一个内容
     *
     * @param value 值
     * @param font  字体
     * @param align 对齐方式
     * @return 添加的文本框
     */
    public static PdfPCell createCell(String value, Font font, int align) {
        PdfPCell cell = new PdfPCell();
        cell.setVerticalAlignment(Element.ALIGN_MIDDLE);
        cell.setHorizontalAlignment(align);
        font.setColor(BaseColor.WHITE);
        cell.setPhrase(new Phrase(value, font));
        cell.setBackgroundColor(BaseColor.GRAY);
        return cell;
    }
    /**
     * 为表格添加一个内容
     *
     * @param value 值
     * @param font  字体
     * @return 添加的文本框
     */
    public static PdfPCell createCell(String value, Font font) {
        PdfPCell cell = new PdfPCell();
        cell.setVerticalAlignment(Element.ALIGN_MIDDLE);
        cell.setHorizontalAlignment(Element.ALIGN_CENTER);
        cell.setPhrase(new Phrase(value, font));
        return cell;
    }
    /**
     * 为表格添加一个内容
     *
     * @param value   值
     * @param font    字体
     * @param align   对齐方式
     * @param colspan 占多少列
     * @return 添加的文本框
     */
    public static PdfPCell createCell(String value, Font font, int align, int colspan) {
        PdfPCell cell = new PdfPCell();
        cell.setVerticalAlignment(Element.ALIGN_MIDDLE);
        cell.setHorizontalAlignment(align);
        cell.setColspan(colspan);
        cell.setPhrase(new Phrase(value, font));
        return cell;
    }
    /**
     * 为表格添加一个内容
     *
     * @param value     值
     * @param font      字体
     * @param align     对齐方式
     * @param colspan   占多少列
     * @param boderFlag 是否有有边框
     * @return 添加的文本框
     */
    public static PdfPCell createCell(String value, Font font, int align, int colspan,
                                      boolean boderFlag) {
        PdfPCell cell = new PdfPCell();
        cell.setVerticalAlignment(Element.ALIGN_MIDDLE);
        cell.setHorizontalAlignment(align);
        cell.setColspan(colspan);
        cell.setPhrase(new Phrase(value, font));
        cell.setPadding(3.0f);
        if (!boderFlag) {
            cell.setBorder(0);
            cell.setPaddingTop(4.0f);
            cell.setPaddingBottom(8.0f);
        }
        return cell;
    }
    /**
     * 创建一个表格对象
     *
     * @param colNumber 表格的列数
     * @return 生成的表格对象
     */
    public static PdfPTable createTable(int colNumber) {
        PdfPTable table = new PdfPTable(colNumber);
        table.setTotalWidth(MAX_WIDTH);
        table.setLockedWidth(true);
        table.setHorizontalAlignment(Element.ALIGN_CENTER);
        table.getDefaultCell().setBorder(1);
        return table;
    }
    public PdfPTable createTable(float[] widths) {
        PdfPTable table = new PdfPTable(widths);
        try {
            table.setTotalWidth(MAX_WIDTH);
            table.setLockedWidth(true);
            table.setHorizontalAlignment(Element.ALIGN_CENTER);
            table.getDefaultCell().setBorder(1);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return table;
    }
    public static PdfPTable createBlankTable() {
        PdfPTable table = new PdfPTable(1);
        table.getDefaultCell().setBorder(0);
        table.addCell(createCell("", keyFont));
        table.setSpacingAfter(20.0f);
        table.setSpacingBefore(20.0f);
        return table;
    }
    /**
     * 提供外界调用的接口，生成以head为表头，list为数据的pdf
     *
     * @param fileName  文件名
     * @param titleName 表单名
     * @param headName  数据表头
     * @param headField 对应类的属性名
     * @param list      数据
     */
    public static <T> File generatePDFs(String fileName, String titleName, String[] headName, String[] headField,
                                        List<T> list) {
        // 建立一个Document对象
        Document document = new Document();
        // 设置页面大小
        document.setPageSize(PageSize.A4);
        File tempFile = null;
        try {
            tempFile = File.createTempFile(fileName, PDF_FILE_SUFFIX);
            PdfWriter.getInstance(document, new FileOutputStream(tempFile));
            document.open();
        } catch (Exception e) {
            e.printStackTrace();
        }
        Class classType = list.get(0).getClass();
        int colNum = headName.length;
        PdfPTable table = createTable(colNum);
        // 添加标题
        table.addCell(createCell(titleName, titleFont, Element.ALIGN_CENTER, colNum, false));
        //设置表头
        for (int i = 0; i < colNum; i++) {
            table.addCell(createCell(headName[i], keyFont, Element.ALIGN_CENTER));
        }
        for (int i = 0; i < list.size(); i++) {
            T t = list.get(i);
            for (int j = 0; j < colNum; j++) {
                //获得get方法,getName,getAge等
                String getMethodName = "get" + headField[j].substring(0, 1).toUpperCase() + headField[j].substring(1);
                try {
                    //通过反射获得相应的get方法，用于获得相应的属性值
                    Method method = classType.getMethod(getMethodName, new Class[]{});
                    logger.debug(getMethodName + ":" + method.invoke(t, new Class[]{}) + ",");
                    //添加数据
                    table.addCell(createCell(method.invoke(t, new Class[]{}).toString(), textFont));
                } catch (Exception e) {
                    logger.error("属性设值错误：入参={},异常={}", JSON.toJSONString(t), e);
                    throw new RuntimeException(e);
                }
            }
        }
        try {
            //将表格添加到文档中
            document.add(table);
        } catch (Exception e) {
            logger.error("");
            throw new RuntimeException(e);
        }
        //关闭流
        document.close();
        return tempFile;
    }
    /**
     * 构建行程单PDF文件
     * @param date 构建日期
     * @param phone 手机号
     * @param baseTripInfos 行程信息
     * @param outputStream 输出流
     */
    public static void createPDFFile(Date date, String phone, List<BaseTripInfo> baseTripInfos, OutputStream outputStream) {
        String[] header = new String[]{"序号", "出站时间", "入站站点", "出站站点", "金额（元）"};
        int[] width = new int[]{20, 45, 35, 35, 35};
        Document document = null;
        try {
            // 建立一个Document对象
            document = new Document();
            // 设置页面大小
            document.setPageSize(PageSize.A4);
            PdfWriter.getInstance(document, outputStream);
            document.open();
            // 设置列以及列宽度
            PdfPTable table = PDFUtil.createTable(header.length);
            table.setWidths(width);
            // 添加标题以及空行
            table.addCell(PDFUtil.createCell("广州地铁—行程单", PDFUtil.titleFont, Element.ALIGN_CENTER, header.length, false));
            table.addCell(PDFUtil.createCell(" ", PDFUtil.titleFont, Element.ALIGN_CENTER, header.length, false));
            // 设置申请日期以及行程起止日期
            table.addCell(PDFUtil.createCell("申请时间：" + cn.hutool.core.date.DateUtil.format(date, "yyyy-MM-dd"), PDFUtil.textFont, Element.ALIGN_LEFT, 2, false));
            table.addCell(PDFUtil.createCell("行程人手机号：" + phone, PDFUtil.textFont, Element.ALIGN_LEFT, 2, false));
            // 计算该行程单总金额
            double totalAmount = 0;
            for (BaseTripInfo baseTripInfo : baseTripInfos) {
                totalAmount += Double.parseDouble(baseTripInfo.getPrice());
            }
            table.addCell(PDFUtil.createCell("共" + baseTripInfos.size() + "笔行程, 合计 " + NumberUtil.round(totalAmount, 2).toString() + "元",
                    PDFUtil.textFont, Element.ALIGN_LEFT, header.length, false));
            // 设置表头
            for (String headName : header) {
                table.addCell(PDFUtil.createCell(headName, PDFUtil.keyFont, Element.ALIGN_CENTER));
            }
            for (int i = 0; i < baseTripInfos.size(); i++) {
                // 序号
                table.addCell(createTextCell(String.valueOf(i + 1)));
                // 出站时间
                table.addCell(createTextCell(baseTripInfos.get(i).getExitDate() == null ? "" :
                        baseTripInfos.get(i).getExitDate()));
                // 进站站点
                table.addCell(createTextCell(baseTripInfos.get(i).getEnterStation() == null ? "" :
                        baseTripInfos.get(i).getEnterStation()));
                // 出站站点
                table.addCell(createTextCell(baseTripInfos.get(i).getExitStation() == null ? "" :
                        baseTripInfos.get(i).getExitStation()));
                // 扣费金额
                double price = Double.parseDouble(baseTripInfos.get(i).getPrice());
                table.addCell(createTextCell(NumberUtil.round(price, 2).toString()));
            }
            //将表格添加到文档中
            document.add(table);
        } catch (DocumentException e) {
            logger.error("Document添加PdfWriter实例失败", e);
        } finally {
            //关闭流
            if (document != null) {
                document.close();
            }
        }
    }
    private static PdfPCell createTextCell(String value) {
        return PDFUtil.createCell(value, PDFUtil.textFont);
    }
}

```
```java
/**
 * 行程单信息
 */
public class BaseTripInfo {
    /**
     * 行程ID
     */
    private String tripId;
    /**
     * 入站站点
     */
    private String enterStation;
    /**
     * 入站时间
     */
    private String enterDate;
    /**
     * 出站站点
     */
    private String exitStation;
    /**
     * 出站时间
     */
    private String exitDate;
    /**
     * 金额
     */
    private String price;
    /**
     * 设备ID
     */
    private String deviceId;
    public String getTripId() {
        return tripId;
    }
    public void setTripId(String tripId) {
        this.tripId = tripId;
    }
    public String getEnterStation() {
        return enterStation;
    }
    public void setEnterStation(String enterStation) {
        this.enterStation = enterStation;
    }
    public String getEnterDate() {
        return enterDate;
    }
    public void setEnterDate(String enterDate) {
        this.enterDate = enterDate;
    }
    public String getExitStation() {
        return exitStation;
    }
    public void setExitStation(String exitStation) {
        this.exitStation = exitStation;
    }
    public String getExitDate() {
        return exitDate;
    }
    public void setExitDate(String exitDate) {
        this.exitDate = exitDate;
    }
    public String getPrice() {
        return price;
    }
    public void setPrice(String price) {
        this.price = price;
    }
    public String getDeviceId() {
        return deviceId;
    }
    public void setDeviceId(String deviceId) {
        this.deviceId = deviceId;
    }
    @Override
    public String toString() {
        return "BaseTripInfo{" +
                "tripId='" + tripId + '\'' +
                ", enterStation='" + enterStation + '\'' +
                ", enterDate='" + enterDate + '\'' +
                ", exitStation='" + exitStation + '\'' +
                ", exitDate='" + exitDate + '\'' +
                ", price='" + price + '\'' +
                ", deviceId='" + deviceId + '\'' +
                '}';
    }
}
```

