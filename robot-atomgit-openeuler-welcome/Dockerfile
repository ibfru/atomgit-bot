FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang && \
    go env -w GOPROXY=https://goproxy.cn,direct

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/robot-gitee-openeuler-welcome
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-gitee-openeuler-welcome -buildmode=pie --ldflags "-s -linkmode 'external' -extldflags '-Wl,-z,now'" .

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 openeuler-welcome && \
    useradd -u 1000 -g openeuler-welcome -s /sbin/nologin -m openeuler-welcome && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    mkdir /home/openeuler-welcome -p && \
    chmod 700 /home/openeuler-welcome && \
    chown openeuler-welcome:openeuler-welcome /home/openeuler-welcome && \
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && \
    rm -rf /tmp/*

USER openeuler-welcome

WORKDIR /opt/app

COPY  --chown=openeuler-welcome --from=BUILDER /go/src/github.com/opensourceways/robot-gitee-openeuler-welcome/robot-gitee-openeuler-welcome /opt/app/robot-gitee-openeuler-welcome

RUN chmod 550 /opt/app/robot-gitee-openeuler-welcome && \
    echo "umask 027" >> /home/openeuler-welcome/.bashrc && \
    echo 'set +o history' >> /home/openeuler-welcome/.bashrc

ENTRYPOINT ["/opt/app/robot-gitee-openeuler-welcome"]
