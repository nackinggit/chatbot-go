package queue

import (
	"context"

	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/xredis"
)

func jsonSerialize[T any](v *T) (string, error) {
	bs, err := util.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bs), err
}

func jsonDeserialize[T any](b string) (T, error) {
	var v T
	err := util.Unmarshal([]byte(b), &v)
	if err != nil {
		return v, err
	}
	return v, err
}

type Queue[T any] struct {
	key         string
	size        int32
	serialize   func(*T) (string, error)
	deserialize func(string) (T, error)
}

func NewQueue[T any](key string) *Queue[T] {
	return &Queue[T]{key: key, size: 0, serialize: jsonSerialize[T], deserialize: jsonDeserialize[T]}
}

func (q *Queue[T]) Size() int32 {
	return q.size
}

func (q *Queue[T]) Enqueue(ctx context.Context, values ...T) error {
	ss := []any{}
	for _, v := range values {
		bs, err := q.serialize(&v)
		if err != nil {
			return err
		}
		ss = append(ss, bs)
	}
	_, err := xredis.ExecRedisCmd(func(mr *xredis.XRedisClient) (any, error) {
		cmd := mr.LPush(ctx, q.key, ss...)
		return cmd.Result()
	})
	return err
}

func (q *Queue[T]) Dequeue(ctx context.Context, count int) ([]T, error) {
	return xredis.ExecRedisCmd(func(mr *xredis.XRedisClient) ([]T, error) {
		cmd := mr.RPopCount(ctx, q.key, count)
		vals, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		ret := []T{}
		for _, v := range vals {
			val, err := q.deserialize(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, val)
		}
		return ret, nil
	})
}
