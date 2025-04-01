### 1. N/A

1. route definition

- Url: /form/example
- Method: POST
- Request: `FormExampleReq`
- Response: `-`

2. request definition



```golang
type FormExampleReq struct {
	Name string `form:"name"`
}
```


3. response definition


### 2. N/A

1. route definition

- Url: /list
- Method: GET
- Request: `-`
- Response: `[]ListItem`

2. request definition



3. response definition



```golang

```

### 3. N/A

1. route definition

- Url: /login
- Method: POST
- Request: `LoginReq`
- Response: `LoginResp`

2. request definition



```golang
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
```


3. response definition



```golang
type LoginResp struct {
	Name string `json:"name"`
}
```

### 4. N/A

1. route definition

- Url: /path/example/:id
- Method: GET
- Request: `PathExampleReq`
- Response: `PathExampleResp`

2. request definition



```golang
type PathExampleReq struct {
	ID string `path:"id"`
}
```


3. response definition



```golang
type PathExampleResp struct {
	Name string `json:"name"`
}
```

### 5. N/A

1. route definition

- Url: /ping
- Method: GET
- Request: `-`
- Response: `-`

2. request definition



3. response definition


### 6. N/A

1. route definition

- Url: /update
- Method: POST
- Request: `UpdateReq`
- Response: `-`

2. request definition



```golang
type UpdateReq struct {
	Arg1 string `json:"arg1"`
}
```


3. response definition


