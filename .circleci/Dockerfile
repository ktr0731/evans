FROM circleci/golang:1.11.5

RUN curl -Lo protoc.zip https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-linux-x86_64.zip && \
      unzip protoc && \
      rm -rf protoc.zip
RUN sudo apt update -y && sudo apt install -y vim
