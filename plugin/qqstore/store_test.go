package qqstore

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func TestGetGoodsList(t *testing.T) {
	var ctx *zero.Ctx
	ctxMock := gomonkey.ApplyMethod(reflect.TypeOf(ctx), "SendChain", func(*zero.Ctx, ...message.MessageSegment) message.MessageID {
		return message.MessageID{}
	})
	defer ctxMock.Reset()
	qqstoredata.getGoodsList(ctx)
}

func TestGetUserInfo(t *testing.T) {
	ctx := &zero.Ctx{
		Event: &zero.Event{
			UserID: 408029164,
		},
	}
	ctxMock := gomonkey.ApplyMethod(reflect.TypeOf(ctx), "SendChain", func(*zero.Ctx, ...message.MessageSegment) message.MessageID {
		return message.MessageID{}
	})
	defer ctxMock.Reset()
	qqstoredata.getUserInfo(ctx)
}

func TestGetGoodDetail(t *testing.T) {
	ctx := &zero.Ctx{
		Event: &zero.Event{
			UserID: 408029164,
		},
		State: zero.State{
			"regex_matched": []string{"1", "1"},
		},
	}
	ctxMock := gomonkey.ApplyMethod(reflect.TypeOf(ctx), "SendChain", func(*zero.Ctx, ...message.MessageSegment) message.MessageID {
		return message.MessageID{}
	})
	defer ctxMock.Reset()
	qqstoredata.getGoodDetail(ctx)
}
