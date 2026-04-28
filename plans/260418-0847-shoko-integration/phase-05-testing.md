# Phase 05: Testing
Status: ⬜ Pending
Dependencies: phase-04-shoko.md

## Objective
Xác nhận rằng kiến trúc "Fallback tự động" hoàn toàn hoạt động và không gặp vỡ luồng cho file non-anime thông thường, bảo vệ trải nghiệm 100% của user cơ bản.

## Implementation Steps
1. [ ] Fake API server hoặc Moq Test để test Unit phần "Matcher ưu tiên".
2. [ ] Test truyền vào một Title Mỹ (Không phải anime) -> Shoko ko có -> Fallback AniList ko có -> TMDb Có -> Nhận đúng.
3. [ ] Test truyền vào Anime có "Shoko ID" -> Match chuẩn.

---
End of Plan.
