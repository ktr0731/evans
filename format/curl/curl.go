// Package curl provides a curl-like formatter implementation.
package curl

import (
	"bytes"
	gojson "encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/golang/protobuf/proto"  //nolint:staticcheck
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/format"
	"github.com/ktr0731/evans/present"
	"github.com/ktr0731/evans/present/json"
	pb "github.com/ktr0731/evans/proto"
	_ "google.golang.org/genproto/googleapis/rpc/errdetails" // For calling RegisterType.
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
)

type responseFormatter struct {
	w io.Writer

	json          present.Presenter
	pbMarshaler   *jsonpb.Marshaler
	pbMarshalerV2 *protojson.MarshalOptions

	wroteHeader, wroteMessage, wroteTrailer bool
}

func NewResponseFormatter(w io.Writer, emitDefaults bool) format.ResponseFormatterInterface {
	return &responseFormatter{
		w:    w,
		json: json.NewPresenter("  "),
		pbMarshaler: &jsonpb.Marshaler{
			EmitDefaults: emitDefaults,
		},
		pbMarshalerV2: &protojson.MarshalOptions{
			EmitUnpopulated: emitDefaults,
			Resolver:        pb.NewAnyResolver(format.DescriptorSource),
		},
	}
}

func (p *responseFormatter) FormatHeader(header metadata.MD) {
	var s []string
	for k, v := range header {
		for _, vv := range v {
			s = append(s, fmt.Sprintf("%s: %s", k, vv))
		}
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	fmt.Fprintf(p.w, "%s\n", strings.Join(s, "\n"))

	p.wroteHeader = true
}

func (p *responseFormatter) FormatMessage(v interface{}) error {
	if p.wroteHeader {
		fmt.Fprintf(p.w, "\n")
	}

	m, err := p.convertProtoMessageToMap(v.(proto.Message))
	if err != nil {
		return err
	}

	msg, err := p.json.Format(m)
	if err != nil {
		return err
	}
	fmt.Fprintf(p.w, "%s\n", msg)

	p.wroteMessage = true

	return nil
}

func (p *responseFormatter) FormatTrailer(trailer metadata.MD) {
	if len(trailer) == 0 {
		return
	}
	if p.wroteHeader || p.wroteMessage {
		fmt.Fprintf(p.w, "\n")
	}

	var s []string
	for k, v := range trailer {
		for _, vv := range v {
			s = append(s, fmt.Sprintf("%s: %s", k, vv))
		}
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	fmt.Fprintf(p.w, "%s\n", strings.Join(s, "\n"))

	p.wroteTrailer = true
}

var replacer = strings.NewReplacer("\n", "", ",", ", ")

func (p *responseFormatter) FormatStatus(status *status.Status) error {
	if p.wroteHeader || p.wroteMessage || p.wroteTrailer {
		fmt.Fprintf(p.w, "\n")
	}
	fmt.Fprintf(p.w, "code: %s\nnumber: %d\nmessage: %q\n", status.Code().String(), status.Code(), status.Message())
	if len(status.Details()) > 0 {
		details := make([]string, 0, len(status.Details()))
		for _, d := range status.Proto().Details {
			m, err := p.convertProtoMessageToMap(d)
			if err != nil {
				return err
			}

			b, err := gojson.MarshalIndent(m, "", "")
			if err != nil {
				return err
			}
			details = append(details, "  "+replacer.Replace(string(b)))
		}
		fmt.Fprintf(p.w, "details: \n%s\n", strings.Join(details, "\n"))
	}
	if status.Code() != codes.OK {
		fmt.Fprintf(p.w, "\n")
	}
	return nil
}

func (p *responseFormatter) Done() error {
	return nil
}

func (p *responseFormatter) convertProtoMessageToMap(m proto.Message) (map[string]interface{}, error) {
	var b []byte
	switch m.(type) {
	case *dynamic.Message:
		// Use jsonpb because protojson can't marshal *dynamic.Message correctly.
		var buf bytes.Buffer
		if err := p.pbMarshaler.Marshal(&buf, m); err != nil {
			return nil, err
		}

		b = buf.Bytes()
	case protov2.Message:
		mb, err := p.pbMarshalerV2.Marshal(proto.MessageV2(m))
		if err != nil {
			return nil, err
		}

		b = mb
	}
	var res map[string]interface{}
	if err := gojson.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	return res, nil
}
