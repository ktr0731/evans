FROM circleci/golang:1.11.2

RUN curl -Lo protoc.zip https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-linux-x86_64.zip && \
      unzip protoc && \
      rm -rf protoc.zip
RUN go get github.com/golang/protobuf/protoc-gen-go && \
      go get github.com/mitchellh/gox && \
      go get github.com/tcnksm/ghr

RUN go get -u gopkg.in/alecthomas/gometalinter.v2 && \
      ln -sf $GOPATH/bin/gometalinter.v2 $GOPATH/bin/gometalinter && \
      gometalinter --install
