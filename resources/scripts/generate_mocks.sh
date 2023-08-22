#!/bin/bash
set -e
set -u

BASE_DIR="${1%/}"
LOOKUP_DIR="${2#/}"
PACKAGE="${3:-}"

for file in $(grep -lre '^type [A-Z].* interface ' ${BASE_DIR}/${LOOKUP_DIR}); do
    mock_path="$(dirname ${file})/mock_$(basename ${file})"
    package="${PACKAGE:-"$(basename $(dirname "${file}"))"}"

    echo -e "// +build mock\n" > $mock_path
    mockgen -source="${file}" -package="${package}" >> $mock_path
done

for dir in $(grep -lr --include=\*_test.go 'gomock' ${BASE_DIR}/${LOOKUP_DIR} | xargs -I{} dirname {} | uniq); do
    echo "-tags=mock" > ${dir}/mock.goconvey
done
echo

exit 0
