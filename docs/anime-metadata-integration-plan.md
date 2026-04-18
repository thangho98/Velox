# Anime Metadata Integration Plan

## Bối Cảnh (Context)
Velox hiện tại hỗ trợ TheTVDB, TMDb, OMDb, TVmaze, và Fanart để lấy metadata. Tuy nhiên, với thư viện chuyên Anime, dữ liệu từ các nguồn cơ bản đôi khi nghèo nàn và thiếu các yếu tố đặc thù (Tags wibu, Romaji titles, Seiyuu casts, Absolute Order). Cần có kế hoạch tích hợp Database từ bên thứ ba.

## Phân Tích Các Nguồn Dữ Liệu Anime
### 1. AniList (Khuyên Dùng Nhất)
* **API Structure**: GraphQL
* **Điểm mạnh**: API siêu hiện đại, tối ưu số lượng request. Artwork sắc nét cực đẹp.
* **Độ Phù hợp**: Rất cao. Có thể thiết kế Velox query metadata bằng Text Search, nhẹ nhàng thân thiện với việc index file từ Cloud Storage (Fshare).

### 2. Jikan (MyAnimeList Unofficial API)
* **API Structure**: RESTful API trả về JSON.
* **Điểm mạnh**: Kho dữ liệu cộng đồng mạnh mẽ nhất thế giới.
* **Độ Phù hợp**: Cao. Có thể dùng làm backend điểm đánh giá hoặc recommendation dự phòng.

### 3. AniDB / Shoko Server
* **API Structure**: UDP / XML API.
* **Chướng ngại vật với Velox**: AniDB dựa trên nguyên lý bắt buộc **Hash toàn bộ Video File (ED2K / CRC32)** để match 100% chính xác. Velox tích hợp **Fshare Cloud Storage**, việc tải nguyên 1 tập phim 1GB từ Fshare về server chỉ để băm Hash tính Checksum sẽ gây thảm họa băng thông và block hệ thống quét (Scan Pipeline). Cố tình xài giả lập Text Search thì API trả về XML cũ, data không ngon bằng AniList.
* **Hướng đi đề xuất**: Thay vì bắt Velox tự xử lý Hash. Hãy làm một plugin cầu nối để pull data qua mạng nội bộ từ một **Shoko Server** bên ngoài (áp dụng cho các hardcore user thích set up riêng).

## Hướng Triển Khai Thực Tế (Roadmap)
* **Giai đoạn 1 (Tùy chỉnh TVDB Core hiện tại)**
  * Cập nhật `nameparser` của Velox để nhận dạng syntax Anime (ví dụ: `[Tên Nhóm] Tên Phim - Tập 12 (1080p).mkv`).
  * Force chuyển mode truyền qua TheTVDB bằng `Absolute Order` thay vì tách season, giúp không lỗi tập đối với kho phim đang có.

* **Giai đoạn 2 (Tích hợp API Độc Tôn)**
  * Thêm thuộc tính Library (hoặc Media Type): `Anime`.
  * Xây dựng `pkg/anilist` hoặc `pkg/jikan`. Khi Pipeline quét trúng Library Type là Anime sẽ tự động bypass TMDb và gọi sang 1 trong 2 nền tảng này lấy ảnh dọc poster / background ngang. Lợi thế là search được bằng chữ (chuỗi gốc), không cần bắt tải file Fshare về tính Hash.

* **Giai đoạn 3 (Support Hardcore Local)**
  * Dành cho các thư viện `Local Folder`. Lúc này có thể dùng FFprobe kết hợp một Hash module để lấy ED2K đập vào AniDB.
