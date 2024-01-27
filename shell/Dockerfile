FROM alpine:3
RUN apk add openssh
RUN adduser -D sif
USER sif
RUN ssh-keygen -q -N "" -t ed25519 -f /home/sif/ssh_host_ed25519_key
CMD ["/bin/sh", "-c", "/usr/sbin/sshd -D -e -p 2022 -h /home/sif/ssh_host_ed25519_key -o AllowTcpForwarding=yes"]
