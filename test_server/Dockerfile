FROM alpine:latest

# app name
ENV AppName=test-httpserver

# mkdir directory
RUN mkdir -p /app/logs

# copy file to image
COPY ${AppName} /app/

# copy configs file to image
#COPY configs /app/configs

# add timezone
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
        && apk --no-cache add ca-certificates tzdata\
        && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
        && echo "Asia/Shanghai" > /etc/timezone \
        && apk del tzdata

WORKDIR /app

# 日志挂载
#VOLUME /app/logs

# port
EXPOSE 12300

LABEL maintainer = "2912882908@qq.com"

CMD ["sh","-c","./$AppName"]