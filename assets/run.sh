#!/bin/bash

# This depends on if JAVA_HOME is already in SYSTEM PATH. If not, this script will fail.
# Alternatively, you may replace this with absolute JAVA_HOME path
set -e;
java -cp "mybatis-generator-core-1.4.1.jar:mybatis-lombok-plugin-1.0.0.jar" \
org.mybatis.generator.api.ShellRunner -configfile "generatorConfig.xml" -overwrite
