syntax v1
project demo

service g0 g1 {
    group api grp0 grp1 {
        post /usr (Login) return (resp)
        post /usr (LoginX) return (resp)
    }
}