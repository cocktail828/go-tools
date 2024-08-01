package kvstore

import "context"

type Option func(context.Context) context.Context
