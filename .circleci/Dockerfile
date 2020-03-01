FROM circleci/golang:1.14.0


RUN curl -Lo protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip && \
      unzip protoc && \
      rm -rf protoc.zip
RUN sudo apt update -y && sudo apt install -y vim
