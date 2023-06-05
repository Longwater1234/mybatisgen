# mybatisgen

Custom MyBatisGenerator written in Golang! 
Generates _mapper_, _model_ (with Lombok annotations) and _controller_ files for a SpringBoot MyBatis project. Uses the
original MyBatis Generator (MBG v1.4) under the hood. Please see
the  [official MBG docs](https://mybatis.org/generator/) .By default, this works for MySQL only (v5.6 or higher).
However, you can add dependencies for other Db vendors, then update `main.go`  / `initDbConn` function. You may also be
required to adjust some SQL queries inside `dbutil.go` to match your Database system.

## Requirements

- Go 1.18 or higher. (for building this project)
- Java JRE 8 or higher. (for simply executing the JAR files).
- Both above need to be in your SYSTEM PATH. Verify in your terminal, run:

```bash
  go version
  java --version
```

## How to use

- Rename file [env.sample.json](env.sample.json) --> `env.json`. Please ensure `env.json` is in .gitignore!
- Edit file [env.json](env.json), replace with your Db credentials (**DON'T** include password!) and your Java/Kotlin
  Springboot full package Name (e.g. `com.mycompany.projectName)`.
- Inside your Terminal/CMD at this project's root directory, run the command below:

```bash
  go build
```

- Then execute the built program `./mybatisgen`
- Type your DB Password (Hidden, will not be stored!). Hit Enter.
- After a short while, the finished result will be saved in `output/` folder.
- Done. Now you can inspect & copy content from `output/` into your SpringBoot Project. 

## How it works (Overview)

1. It reads Db credentials and Java project packageName from [env.json](env.json)
2. Then connects to Database to get list of all tables, keys and relations.
3. From the result from (1) and (2) above, it uses Go Templates to create the *generatorConfig.xml* file. (can be
   modified, see the [reference](https://mybatis.org/generator/configreference/xmlconfig.html))
4. It then starts Java process to run`mybatis-generator-core-1.4.1.jar` with XML file from (3) above passed as args.  (
   IMPORTANT: requires Java 8 or higher.).
5. *Model* and *Mapper* classes are created for each table, with Lombok annotations.
6. Please note, it will generate 100% Java-based "Dynamic Mappers" by default. No classic XML mappers are created.
   Therefore, your SpringBoot project will require an extra dependency : `mybatis-dynamic-sql`.
7. As a reminder, you can still generate classic XML mappers or Kotlin src files if you wish, follow link in Step (3) above.
8. Finally, it generates Spring _Rest Controllers_ for common actions for each table.
9. On Success, results are exported into `output/` folder in the current directory, and *generatorConfig.xml* file is
   deleted (for security purposes).

### Important FOLDERS

- [assets](assets)/ (JAR libs, DB drivers, shell scripts)
- [mytemplates](mytemplates)/ (go templates)

### License

[MIT License](LICENSE). (c) Davis 大卫, 2023
