syntax v1
project github.com/xxx/demo

struct Login {
    // binding:"required" 表示该字段必填；json:"username" 对应请求体的 "username" 键
    Username string `json:"username" binding:"required,min=3,max=20"` // 用户名：必填，长度 3-20
    Password string `json:"password" binding:"required,min=6"`       // 密码：必填，最小长度 6
    Email    string `json:"email" binding:"omitempty,email"`        // 邮箱：可选，格式需合法
    Age      int    `json:"age" binding:"omitempty,min=18,max=120"` // 年龄：可选，范围 18-120
}

struct LoginX {
    A int
    B string
}

struct Resp {
    Code int
    Msg string
    Data any
}

service a b {
    group /api  x e{
        @handler handlelogin4
        post /usr (Login) return (Resp)
    }
}
