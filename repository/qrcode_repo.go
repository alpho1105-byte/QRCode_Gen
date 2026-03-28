package repository

// 跟資料存取交互
import (
	"qrcode-gen/model"
)

/*
方法			對應的API					 做什麼

CreatePOST 		/v1/qr_code					新增一筆紀錄
GetByTokenGET 	/v1/qr_code/{token}			用 token 查紀錄
GetByUserID		（管理頁面）			 	 查某使用者的所有 QR code
Update    		PUT /v1/qr_code/{token}		改 URL
Delete     		DELETE /v1/qr_code/{token}	刪除
TokenExists		（建立時檢查碰撞用）		  確認 token 是否已存在
*/
type Repository interface {
	Create(qr *model.QRCode) error
	GetByToken(qrToken string) (*model.QRCode, error)
	GetByUserID(userID string) ([]*model.QRCode, error)
	Update(qrToken string, url string) error
	Delete(qrToken string) error
	TokenExists(qrToken string) (bool, error)
}

// --- In-Memory 實作 ---
