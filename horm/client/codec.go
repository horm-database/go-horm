// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/horm-database/common/codec"
	cp "github.com/horm-database/common/proto"
	"github.com/horm-database/common/types"
)

var (
	defaultCodec = &clientCodec{}
)

// Used for rpc clientside codec.
type clientCodec struct{}

const (
	frameHeadLen    = uint16(10) // total length of frame head: 10 bytes
	protocolVersion = 1          // protocol version v1，影响 framer 帧解析
)

// Encode It encodes reqBody into reqBuf. New msg will be cloned by client stub.
func (c *clientCodec) Encode(msg *codec.Msg, reqBody []byte) ([]byte, error) {
	reqParam, ok := msg.FrameCodec().(*ReqParam)
	if !ok {
		reqParam = &ReqParam{}
	}

	frameHead := codec.NewFrameHead()

	reqHeader, err := getRequestHead(msg)
	if err != nil {
		return nil, err
	}

	reqHeaderBuf, err := proto.Marshal(reqHeader)
	if err != nil {
		return nil, err
	}

	reqBuf, err := frameHead.Construct(reqHeaderBuf, reqBody)
	if err != nil {
		return nil, err
	}

	if reqParam.Encryption == codec.FrameTypeSignature {
		// 签名包
		signFrameHead := codec.NewSignFrameHead()
		return signFrameHead.Construct(
			reqParam.WorkspaceID,
			types.StringToBytes(reqParam.Token), reqBuf)
	} else if reqParam.Encryption == codec.FrameTypeEncrypt {
		// 加密包
		encryptFrameHead := codec.NewEncryptFrameHead()
		return encryptFrameHead.Construct(
			reqParam.WorkspaceID,
			types.StringToBytes(reqParam.Token), reqBuf)
	}

	return reqBuf, nil
}

// Decode It decodes respBuf into respBody.
func (c *clientCodec) Decode(msg *codec.Msg, respBuf []byte) (*cp.ResponseHeader, []byte, error) {
	if len(respBuf) < int(frameHeadLen) {
		return nil, nil, errors.New("client decode rsp buf len invalid")
	}

	frameHead := &codec.FrameHead{
		FrameType: 0, // default
		Version:   protocolVersion,
	}

	msg.WithFrameCodec(frameHead)

	frameHead.Extract(respBuf)

	if frameHead.TotalLen != uint32(len(respBuf)) {
		return nil, nil, fmt.Errorf("total len %d is not actual buf len %d", frameHead.TotalLen, len(respBuf))
	}

	// get response head
	respHeader, err := getResponseHead(msg)
	if err != nil {
		return nil, nil, err
	}

	if frameHead.HeaderLen == 0 {
		return nil, nil, errors.New("client decode pb head len empty")
	}

	begin := int(frameHeadLen)
	end := int(frameHeadLen) + int(frameHead.HeaderLen)
	if end > len(respBuf) {
		return nil, nil, errors.New("client decode pb head len invalid")
	}

	err = proto.Unmarshal(respBuf[begin:end], respHeader)
	if err != nil {
		return nil, nil, err
	}

	return respHeader, respBuf[end:], nil
}

func getRequestHead(msg *codec.Msg) (*cp.RequestHeader, error) {
	if msg.ClientReqHead() != nil {
		reqHeader, ok := msg.ClientReqHead().(*cp.RequestHeader)
		if !ok {
			return nil, errors.New("client encode proto head type invalid, must be request protocol head")
		}
		return reqHeader, nil
	}

	reqHeader := &cp.RequestHeader{}
	msg.WithClientReqHead(reqHeader)
	return reqHeader, nil
}

func getResponseHead(msg *codec.Msg) (*cp.ResponseHeader, error) {
	if msg.ClientRespHead() != nil {
		respHeader, ok := msg.ClientRespHead().(*cp.ResponseHeader)
		if !ok {
			return nil, errors.New("client decode response head type invalid, must be response protocol head")
		}
		return respHeader, nil
	}

	respHeader := &cp.ResponseHeader{}
	msg.WithClientRespHead(respHeader)
	return respHeader, nil
}
