#!/bin/bash
#获取环境变量
target=$APP_TARGET
if [ ${brand_code} != "" ]
    then ./product-api export -t ${target}
else 
    ./product-api api-server
fi