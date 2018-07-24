#!/bin/bash

set -e
set -x

scriptdir=$(cd $(dirname $0); pwd)

out=${out:-bench}

[ -e ${out}/config ] && . ${out}/config ||:

if [ -z "${ECHOST}" ]; then
    echo "need to set host variables"
    exit 1
fi

# FOR EC ODATALITE TESTING ONLY: set TOKEN to 'oauthtest token' output and this script will set and unset as appropriate to benchark each stack

export CURL_OPTS=-k
export prot=https
export goapacheuri=${goapacheuri:-/redfish/v1/Managers/CMC.Integrated.1/Attributes}

# run TOP and save results during each run
export profile=1

# run basic auth tests for each that supports
export runbasic=1

# run token auth tests for each that supports
export runtoken=1

# either run 'oauthtest token' and set TOKEN= to that value
#  -or-
# get root ssh access to the system:
#   ssh-copy-id root@IP.ADD.RESS
#
# since you need ssh access to get the 'top' results anyways, this is the best option
export BACKUP_TOKEN=$(ssh root@$ECHOST oauthtest token | grep ^Local | cut -d: -f2)
export TOKEN=${TOKEN:-${BACKUP_TOKEN}}

mkdir bench ||:

####################
# go-redfish tests
####################
export user=Administrator
export pass=password
host=$ECHOST TOKEN= port=8443 ${scriptdir}/walk.sh        ${out}/go-https-walk
host=$ECHOST TOKEN= port=8443 ${scriptdir}/vegeta-test.sh ${out}/go-https-vegeta
host=$ECHOST TOKEN= port=8443 ${scriptdir}/runhey.sh      ${out}/go-https-hey
host=$ECHOST TOKEN= port=8443 ${scriptdir}/runab.sh       ${out}/go-https-ab

# test go spacemonkey openssl integration
# apparently incopatible with the new certs
#host=$ECHOST TOKEN= port=9443 ${scriptdir}/walk.sh        ${out}/go-openssl-walk
#host=$ECHOST TOKEN= port=9443 ${scriptdir}/runab.sh       ${out}/go-openssl-ab
#host=$ECHOST TOKEN= port=9443 ${scriptdir}/vegeta-test.sh ${out}/go-openssl-vegeta
#host=$ECHOST TOKEN= port=9443 ${scriptdir}/runhey.sh      ${out}/go-openssl-hey

# test go-redfish through apache
mkdir -p ${out}/go-apache-vegeta
grep /Attributes ${out}/go-https-vegeta/to-visit.txt > ${out}/go-apache-vegeta/to-visit.txt
host=$ECHOST port=443 ${scriptdir}/vegeta-test.sh     ${out}/go-apache-vegeta

uri=${goapacheuri} \
host=$ECHOST port=443 ${scriptdir}/runhey.sh          ${out}/go-apache-hey

uri=${goapacheuri} \
host=$ECHOST port=443 ${scriptdir}/runab.sh           ${out}/go-apache-ab

export user=root
export pass=calvin
host=$ECHOST port=443 ${scriptdir}/walk.sh        ${out}/odatalite-walk
host=$ECHOST port=443 ${scriptdir}/runab.sh       ${out}/odatalite-ab
host=$ECHOST port=443 ${scriptdir}/vegeta-test.sh ${out}/odatalite-vegeta