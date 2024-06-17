FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang && \
    go env -w GOPROXY=https://goproxy.cn,direct

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/robot-gitee-cla
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-gitee-cla -buildmode=pie --ldflags "-s -linkmode 'external' -extldflags '-Wl,-z,now'" .

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 cla && \
    useradd -u 1000 -g cla -s /sbin/nologin -m cla && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    mkdir /home/cla -p && \
    chmod 700 /home/cla && \
    chown cla:cla /home/cla && \
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && \
    rm -rf /tmp/*

USER cla

WORKDIR /opt/app

COPY  --chown=cla --from=BUILDER /go/src/github.com/opensourceways/robot-gitee-cla/robot-gitee-cla /opt/app/robot-gitee-cla

RUN chmod 550 /opt/app/robot-gitee-cla && \
    echo "umask 027" >> /home/cla/.bashrc && \
    echo 'set +o history' >> /home/cla/.bashrc

ENTRYPOINT ["/opt/app/robot-gitee-cla"]
