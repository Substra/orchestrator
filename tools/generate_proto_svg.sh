#!/bin/bash

export PROTO_SRC_PATH=$(pwd)/lib/asset
export PROTO_DOC_PATH=$(pwd)/docs/assets/protos

cd $(pwd)/tools/

if [ ! -d protodot ]; then
    # install protodot
    git clone git@github.com:seamia/protodot.git
    cd protodot

    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # linux
        apt-install graphviz
        sed -i 's@"${HOME}/protodot/generated"@"${PROTO_DOC_PATH}"@g' config.json
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac OSX
        brew install graphviz
        gsed -i 's@"${HOME}/protodot/generated"@"${PROTO_DOC_PATH}"@g' config.json
    fi
    go install
fi

#find modified protofiles
find $PROTO_SRC_PATH -name "*.proto" -mtime -1 | while read fname; do
    protodot -src $fname
    
    # remove intermediary dot file
    find $PROTO_DOC_PATH -name "*.dot" -print0 \
        | xargs -0 ls -1 -t \
        | head -1 \
        | xargs rm

    # set friendly filname name for svg output
    find $PROTO_DOC_PATH -name "*.dot.svg" -print0 \
        | xargs -0 ls -1 -t \
        | head -1 \
        | xargs -I '{}' mv {} $PROTO_DOC_PATH/$(basename $fname .proto).svg 
done