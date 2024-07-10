#!/bin/bash

set -e

rm -rf dist

git_commit="21232f297a57a5a743894a0e4a801fc3"
ldflags="-s -w -extldflags -static"
if [ -d ".git" ]; then
  # 检查是否有变更
  if git diff-index --quiet HEAD --; then
    echo "No changes to commit"
  else
    # 使用默认提交消息，如果没有提供
    commit_message=${1:-"update"}

    # 提交变更
    git commit -am "$commit_message"
  fi

  # 获取当前Git提交哈希
  git_commit=$(git rev-parse HEAD)
  ldflags="$ldflags -X main.commit=$git_commit"
  echo "commit: $git_commit"
fi

# 编译目标平台列表
targets=("windows/amd64" "linux/amd64" "darwin/arm64")

# 遍历所有目标平台并进行编译
for target in "${targets[@]}"
do
  IFS='/' read -ra parts <<< "$target"
  GOOS=${parts[0]}
  GOARCH=${parts[1]}

  export GOOS=$GOOS
  export GOARCH=$GOARCH

  for APP in `ls cmd`; do
    bin_name="$APP-$GOOS-$GOARCH"
    if [ "$GOOS" == "windows" ]; then
      bin_name+=".exe"
    fi
    output_name="dist/$bin_name"
    echo "building $bin_name"
    if [ "$GOOS" == "linux" ]; then
      CGO_ENABLED=0 go build -o "$output_name" -trimpath -a -ldflags "$ldflags" cmd/$APP/*
    else
      CGO_ENABLED=0 go build -o "$output_name" -trimpath -a -ldflags "$ldflags" cmd/$APP/*;
    fi
  done
done