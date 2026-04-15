# Plan V: AI Translate — Dialog-Aware Batching
Created: 2026-04-15
Status: ⬜ Pending

## Overview
Hiện tại `TranslateSRT` chia batch LLM cứng theo chỉ số cue (mỗi 50 cue). Cách này cắt ngang hội thoại, phá mạch context mà LLM cần để suy luận speaker/pronoun/tone. Plan này thay logic batching bằng thuật toán gap-aware (tiered thresholds theo khoảng lặng giữa cue) và bổ sung context overlap cho các ranh giới buộc phải cắt giữa hội thoại dày.

Đã verify vấn đề bằng 2 mẫu thực tế lấy từ NAS:
- **Malcolm in the Middle S01E01** (sitcom, 22min, 551 cues, 24.6 cues/phút) — rất nhiều cắt xấu ở gap 1.5–2s (Q&A pairs, mid-thought).
- **Avatar: Fire and Ash** (movie, 187min, 2442 cues, 13 cues/phút) — gap tự nhiên dồi dào, thuật toán đơn giản là đủ.

## Goal
- Giữ nguyên `Translator` interface cho Phase 1 (không đổi public API) để risk thấp.
- Ưu tiên cắt batch tại scene break / dialog pause tự nhiên.
- Cắt batch ở giữa hội thoại dày vẫn giữ được mạch (Phase 2 context overlap).
- Giảm số lần LLM trả translation sai/lệch count do mất context.

## Non-Goals
- Không thay đổi provider (OpenAI/Gemini/Anthropic), không đổi model.
- Không sửa DeepL/Google translator (các dịch vụ non-LLM không được hưởng lợi từ context).
- Không đổi định dạng SRT lưu trữ, không thêm migration.
- Không đụng tới auto-translate background worker logic (chỉ đụng layer translate pkg).

## Evidence (từ phân tích 2 mẫu)

| Threshold | Malcolm batches | Malcolm forced | Avatar batches | Avatar forced |
|---|---|---|---|---|
| 1.5s | 17 | 2 | 108 | 0 |
| 2.0s | 16 | 2 | 106 | 0 |
| 3.0s | 16 | 3 | 98 | 1 |

→ Threshold đơn không vừa cả 2 loại. Tiered thresholds (3s ưu tiên, 2s khi batch ≥30 cue, 1.5s khi ≥40 cue) cho Malcolm 14/16 natural cuts, Avatar 106/106 natural cuts.

## Phases

| # | Tên | Mô tả ngắn | Risk |
|---|---|---|---|
| 01 | [Tiered Gap-Based Splitting](phase-01-tiered-gap-splitting.md) | Thay vòng lặp `i += batchSize` bằng greedy chunker gap-aware. Chunker emit `Batch{LeftForced, RightForced}` metadata để Phase 2 dùng. Không đổi `Translator` interface. | Thấp |
| 02 | [Context Overlap for Forced Cuts](phase-02-context-overlap.md) | **Pre-req: fix parser count bug** (pre-existing, silent corruption). Sau đó gửi overlap cue **chỉ** tại forced boundaries. Natural boundaries không có overhead. Update 3 AI provider. | Trung |

### Phát hiện trong review

**Round 1:** `decodeLLMTranslations` có bug — nhánh indexed parser không so `len(translations)` với `expected`. LLM trả extra items (VD 6 thay vì 3) → `TranslateSRT` silent ghi đè cue của batch kế tiếp qua `cues[b.start+j].Text = t`. Pre-existing bug. Fix là Step 1 của Phase 2, ship độc lập được.

**Round 2:** Fix count chưa đủ — parser còn 2 lỗ hổng nữa:
- `byIndex[t.Index] = t.Text` silent-overwrite duplicates → `[0,1,1,2]` co thành 3 item giả, bypass count check.
- `maxIndex:=-1` initial + `index:-1` trong payload → loop không chạy → return 0 items "thành công", cũng bypass.
- Validation đúng: every index unique, `0 <= index < expected`.

**Round 2 (coupling):** Phase 1 fallback fixed-size khi timing hỏng, nếu set `RightForced=true` cho mọi interior → Phase 2 overlap khắp nơi → ngược với "graceful degrade". Giải pháp: chunker trả `ChunkResult{Batches, TimingValid}`, Phase 2 early-out không overlap khi `TimingValid==false`.

## Tech Scope
- Backend Go only: `backend/pkg/translate/`
- Files chính: `translate.go` (chunker), `ai.go` (prompt + response parsing), tests đi kèm.
- Không đụng DB, không đụng handler, không đụng service/subtitle_auto_translate.go.

## Success Criteria
- Phase 1 ship: trên 1 movie điển hình (Avatar), ≥95% batch boundaries rơi vào gap ≥2s.
- Phase 2 ship: trên 1 sitcom điển hình (Malcolm), không còn forced cut nào thiếu context (pairs Q&A không rơi vào 2 batch khác nhau mà không có overlap).
- Unit tests cho chunker cover: dense dialog, sparse movie, no-gap (forced cut fallback), single cue, empty.
- Không regression: existing subtitle auto-translate flow chạy bình thường, output SRT giữ nguyên số cue và thứ tự.

## Quick Commands
- Start Phase 1: `/code plans/plan-v-ai-translate-dialog-aware/phase-01-tiered-gap-splitting.md`
- Start Phase 2: `/code plans/plan-v-ai-translate-dialog-aware/phase-02-context-overlap.md`
- Run tests: `cd backend && go test ./pkg/translate/...` (go.mod ở `backend/`, không chạy từ repo root được)
