package entity

type User struct {
	Account string
	Password string
	IsWhite string // 是否启用白名单
}

// TODO UserWhite 登录白名单表
type UserWhite struct {

}

// TODO JWT 下发的 jwt   ip对应一个jwt, 登录会验证地址

