package grpc

import (
	"strconv"

	"golang.org/x/net/context"
	grpc1 "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	grpcHeaderCode          = "resp_code"
	grpcHeaderMessage       = "resp_msg"
	grpcHeaderRequestUserID = "req_uid"
	grpcHeaderCode2         = "x-server-error-code"
	grpcHeaderMessage2      = "x-server-error-message"
)



func SetStatus(ctx context.Context, code int, message ...string) {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	} else {
		msg = MessageFromCode(code)
	}

	grpc1.SetHeader(ctx, metadata.Pairs(
		grpcHeaderCode, strconv.Itoa(code),
		grpcHeaderMessage, msg,
		grpcHeaderCode2, strconv.Itoa(code),
		grpcHeaderMessage2, msg,
	))
}

func GetStatus(md metadata.MD) (code int, message string) {
	if v, ok := md[grpcHeaderCode2]; ok && len(v) > 0 {
		code, _ = strconv.Atoi(v[0])
	} else if v, ok := md[grpcHeaderCode]; ok && len(v) > 0 {
		code, _ = strconv.Atoi(v[0])
	}

	if v, ok := md[grpcHeaderMessage2]; ok && len(v) > 0 {
		message = v[0]
	} else if v, ok := md[grpcHeaderMessage]; ok && len(v) > 0 {
		message = v[0]
	}
	return
}
