// Package json provides a JSON formatter implementation.
package json

import (
	"bytes"
	gojson "encoding/json"
	"io"

	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/golang/protobuf/proto"  //nolint:staticcheck
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/format"
	"github.com/ktr0731/evans/present"
	"github.com/ktr0731/evans/present/json"
	pb "github.com/ktr0731/evans/proto"
	_ "google.golang.org/genproto/googleapis/rpc/errdetails" // For calling RegisterType.
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
)

// responseFormatter is a formatter that formats *usecase.GRPCResponse into a JSON object.
type responseFormatter struct {
	w io.Writer
	s struct {
		Status struct {
			Code    string        `json:"code"`
			Number  uint32        `json:"number"`
			Message string        `json:"message"`
			Details []interface{} `json:"details,omitempty"`
		} `json:"status,omitempty"`
		Header   *metadata.MD             `json:"header,omitempty"`
		Messages []map[string]interface{} `json:"messages,omitempty"`
		Trailer  *metadata.MD             `json:"trailer,omitempty"`
	}
	p             present.Presenter
	pbMarshaler   *jsonpb.Marshaler
	pbMarshalerV2 *protojson.MarshalOptions
}

func NewResponseFormatter(w io.Writer, emitDefaults bool) format.ResponseFormatterInterface {
	return &responseFormatter{
		w: w,
		p: json.NewPresenter("  "),
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
	p.s.Header = &header
}

func (p *responseFormatter) FormatMessage(v interface{}) error {
	m, err := p.convertProtoMessageToMap(v.(proto.Message))
	if err != nil {
		return err
	}
	p.s.Messages = append(p.s.Messages, m)
	return nil
}

func (p *responseFormatter) FormatTrailer(trailer metadata.MD) {
	p.s.Trailer = &trailer
}

func (p *responseFormatter) FormatStatus(s *status.Status) error {
	var details []interface{}
	if len(s.Details()) != 0 {
		details = make([]interface{}, 0, len(s.Details()))
		for _, d := range s.Proto().Details {
			m, err := p.convertProtoMessageToMap(d)
			if err != nil {
				return err
			}
			details = append(details, m)
		}
	}

	p.s.Status = struct {
		Code    string        `json:"code"`
		Number  uint32        `json:"number"`
		Message string        `json:"message"`
		Details []interface{} `json:"details,omitempty"`
	}{
		Code:    s.Code().String(),
		Number:  uint32(s.Code()),
		Message: s.Message(),
		Details: details,
	}
	return nil
}

func (p *responseFormatter) Done() error {
	s, err := p.p.Format(p.s)
	if err != nil {
		return err
	}
	_, err = io.WriteString(p.w, s+"\n")
	return err
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
