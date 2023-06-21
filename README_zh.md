# mybatisgen

[README_en](README.md)

用 Golang 编写的自定义 MyBatisGenerator！
为 SpringBoot MyBatis 项目生成 _mapper_, _model_（带有 Lombok 注释）和 CRUD _controllers_ 。在引擎盖下使用原始的 MyBatis Generator (MBG v1.4)。请参阅[官方 MBG 文档](https://mybatis.org/generator/)。默认情况下，这仅适用于 MySQL（v5.6 或更高版本）。但是，您可以为其他 Db 供应商添加依赖项，然后更新 main.go / `initDbConn`函数。您可能还需要调整 `dbutil.go`中的一些 SQL 查询以适合您的 DBMS。

## 要求

- Go 1.18+ 或更高。 （用于构建此项目）
- Java 8+ 或更高版本。 （用于简单地执行 JAR 文件）。
- 以上两者都需要在您的系统路径中。在您的 terminal 中验证，运行：

```bash
  go version
  java --version

```

## 如何使用

- 重命名文件[env.sample.json ](env.sample.json) --> `env.json` 。请确保`env.json`在 .gitignore 中！
- 编辑文件 env.json ，替换为您的 Db 凭据（**不要包括密码！** ）和您的 Java/Kotlin Springboot
  完整包名称（例如`com.mycompany.projectName`） 。
- 在此项目根目录的 terminal/CMD 中，运行以下命令：

```bash
  go build
```

- 然后执行构建的程序`./mybatisgen`
- 输入您的数据库密码（隐藏，不会被存储！）。按回车键。
- 片刻之后，完成的结果将保存在 output/文件夹中。
- 完毕。现在您可以检查输`output/`中的内容并将其复制到您的 SpringBoot 项目中。

## 它是如何工作的 (概述)

1. `env.json`读取 Db 连接参数和 Java packageName 。从命令行接受密码。
2. 然后连接到数据库以获取所有表、键和关系的列表。
3. 根据上面（1）和（2）的结果，它使用 Go Templates 创建了*generatorConfig.xml*文件。
   （可以修改，见[参考资料](https://mybatis.org/generator/configreference/xmlconfig.html)）
4. 然后它启动 Java 进程以运行 mybatis-generator-core-1.4.1.jar ，并将上面 (3) 中的 XML 文件作为参数传递。 （重要提示：需要
   Java 8 或更高版本。）。
5. 为每个表创建*Model*和*Mapper 类，并带有 Lombok 注释。*
6. 请注意，默认情况下它将生成 100% 基于 Java 的“动态映射器”。没有创建经典的 XML 映射器。因此，您的 SpringBoot 项目将需要一个额外的依赖项：
   mybatis-dynamic-sql 。
7. 提醒一下，如果您愿意，您仍然可以生成经典的 XML 映射器或 Kotlin src 文件，请点击上面步骤 (3) 中的链接。
8. 为每个表的常见操作生成 Spring _Rest 控制器。_
9. 成功时，结果将导出到当前目录中的`output`/文件夹，并删除*generatorConfig.xml 文件（出于安全目的）。*

### 重要文件夹

- [assets](assets)/（JAR 库、数据库驱动程序、shell 脚本）
- [mytemplates](mytemplates)/ (去模板)

## 执照

[MIT LICENSE](LICENSE)。 (c) Davis 大卫，2023 年
