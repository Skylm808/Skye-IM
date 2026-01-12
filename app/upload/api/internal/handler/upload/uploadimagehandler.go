package upload

import (
	"net/http"

	"SkyeIM/app/upload/api/internal/logic/upload"
	"SkyeIM/app/upload/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadImageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析multipart form
		err := r.ParseMultipartForm(10 << 20) // 10MB
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取文件
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		// 调用logic
		l := upload.NewUploadImageLogic(r.Context(), svcCtx)
		resp, err := l.UploadImage(fileHeader)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
