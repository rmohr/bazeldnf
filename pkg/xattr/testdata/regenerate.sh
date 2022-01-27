#!/usr/bin/env bash

TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

touch ${TMPDIR}/selinux
touch ${TMPDIR}/cap_net_bind_service
touch ${TMPDIR}/cap_chown
touch ${TMPDIR}/cap_sys_ptrace
touch ${TMPDIR}/cap_all
sudo chcon -t user_home_t ${TMPDIR}/selinux

sudo setcap 'cap_net_bind_service=+ep' ${TMPDIR}/cap_net_bind_service
sudo setcap 'cap_chown=+ep' ${TMPDIR}/cap_chown
sudo setcap 'cap_sys_ptrace=+ep' ${TMPDIR}/cap_sys_ptrace
sudo setcap 'cap_net_bind_service,cap_chown,cap_sys_ptrace=+ep' ${TMPDIR}/cap_all
tar -C ${TMPDIR} --xattrs -cvf xattr.tar .
