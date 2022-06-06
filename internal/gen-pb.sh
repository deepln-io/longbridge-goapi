#!/bin/sh
# This script is used to update the longbridge protobuf files from longbridge repository and re-generate the go packages for longbridge client communication.
# When longbridge has published new updates, run this script and check in the newly generated code.

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
export PATH=$PATH:$HOME/go/bin

cd ../../..
mkdir -p longbridgeapp
cd longbridgeapp
if [ ! -d openapi-protobufs ]
then
  git clone https://github.com/longbridgeapp/openapi-protobufs
else
  echo "openapi-protobufs exists, skipping clone"
fi
cd openapi-protobufs
echo pull latest longbridge protobuf version
git pull
cd ../../..  # to home dir containing github.com dir
rm -fr github.com/deepln-io/longbridge-goapi/internal/pb/*
mkdir github.com/deepln-io/longbridge-goapi/internal/pb/src

for dir in trade control quote
do
  for file in github.com/longbridgeapp/openapi-protobufs/$dir/*.proto
  do
    name=`basename $file`
    out=github.com/deepln-io/longbridge-goapi/internal/pb/src/$name
    cat $file | sed '/^[ \t]*package/a \
option go_package = "github.com/deepln-io/longbridge/internal/pb/'$dir'";
' >$out
    clang-format -i $out
  done
done

for file in github.com/deepln-io/longbridge-goapi/internal/pb/src/*.proto
do
  protoc -I=github.com/deepln-io/longbridge-goapi/internal/pb/src --go_out=. $file
done
