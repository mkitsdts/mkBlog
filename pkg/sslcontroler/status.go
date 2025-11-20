package sslcontroler

const (
	FAILED_TO_GET_STATUS         int = -1
	STATUS_VALID                 int = 0  // 有效且过期时间还有两个星期以上
	NEED_RENEWAL                 int = 1  // 需要续期
	APPLYING_CERTIFICATE         int = 3  // 申请中
	APPLYING_CERTIFICATE_FAILED  int = 4  // 申请失败
	APPLYING_CERTIFICATE_SUCCESS int = 5  // 申请成功，等待验证
	ADDING_DOMAIN_RECORD         int = 6  // 添加域名解析中
	ADD_DOMAIN_RECORD_FAILED     int = 7  // 添加域名解析失败
	ADD_DOMAIN_RECORD_SUCCESS    int = 8  // 添加域名解析成功
	VALIDATING_DOMAIN            int = 9  // 域名验证中
	VALID_DOMAIN_FAILED          int = 10 // 域名验证失败
	VALID_DOMAIN_SUCCESS         int = 11 // 域名验证成功

	STATUS_UNKNOWN int = 99 // 未知状态
)
