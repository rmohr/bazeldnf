#!/usr/bin/env bash

TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

touch ${TMPDIR}/selinux
touch ${TMPDIR}/cap_net_bind_service
touch ${TMPDIR}/cap_net_admin
touch ${TMPDIR}/cap_net_raw
touch ${TMPDIR}/cap_chown
touch ${TMPDIR}/cap_sys_ptrace
touch ${TMPDIR}/cap_all
sudo chcon unconfined_u:object_r:user_home_t:s0 ${TMPDIR}/selinux

sudo setcap 'cap_net_bind_service=+ep' ${TMPDIR}/cap_net_bind_service
sudo setcap 'cap_net_admin=+ep' ${TMPDIR}/cap_net_admin
sudo setcap 'cap_net_raw=+ep' ${TMPDIR}/cap_net_raw
sudo setcap 'cap_chown=+ep' ${TMPDIR}/cap_chown
sudo setcap 'cap_sys_ptrace=+ep' ${TMPDIR}/cap_sys_ptrace
sudo setcap 'cap_net_bind_service,cap_net_admin,cap_net_raw,cap_chown,cap_sys_ptrace=+ep' ${TMPDIR}/cap_all
tar -C ${TMPDIR} --xattrs -cvf xattr.tar .
