FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang && \
    go env -w GOPROXY=https://goproxy.cn,direct

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/robot-gitee-label
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-gitee-label -buildmode=pie --ldflags "-s -linkmode 'external' -extldflags '-Wl,-z,now'" .

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 label && \
    useradd -u 1000 -g label -s /sbin/nologin -m label && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    mkdir /home/label -p && \
    chmod 700 /home/label && \
    chown label:label /home/label && \
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && \
    rm -rf /tmp/*

USER label

WORKDIR /opt/app

COPY  --chown=label --from=BUILDER /go/src/github.com/opensourceways/robot-gitee-label/robot-gitee-label /opt/app/robot-gitee-label

RUN chmod 550 /opt/app/robot-gitee-label && \
    echo "umask 027" >> /home/label/.bashrc && \
    echo 'set +o history' >> /home/label/.bashrc

ENTRYPOINT ["/opt/app/robot-gitee-label"]
