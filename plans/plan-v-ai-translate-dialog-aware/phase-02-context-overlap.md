# Phase 02: Context Overlap for Forced Cuts
Status: ⬜ Pending
Dependencies: Phase 01 (gap-based chunker phải merge trước)

## Objective
Với các batch vẫn phải cắt ở giữa hội thoại dày (sitcom, rap monolog, news), gửi thêm N cue trước và N cue sau batch vào prompt dưới dạng "context only". LLM đọc context để hiểu mạch hội thoại (speaker, pronoun, continuation) nhưng **chỉ dịch phần giữa**. Chỉ áp dụng cho 3 AI provider (OpenAI/Gemini/Anthropic), không đụng DeepL/Google.

## Rationale
- Phase 1 loại được ~95% cắt xấu nhờ gap-aware, nhưng với Malcolm sitcom vẫn còn 2-3 forced cuts/episode do dialog quá dày (89.6% gap <0.3s).
- Tại các forced cuts đó, LLM mất mạch → dịch sai pronoun Việt (anh/em, mày/tao), sai tone.
- Overlap **chỉ** áp dụng tại boundary `forced=true` (từ metadata Phase 1 chunker). Natural boundary không cần context — tiết kiệm token. Với Avatar movie (0 forced cuts) → overhead = 0. Với Malcolm (3 forced / 17 boundaries) → overhead chỉ trên 3 batches.

## Pre-requisite: Fix parser count bug (pre-existing)
Trong quá trình review plan phát hiện `decodeLLMTranslations` ([ai.go:663-694](../../backend/pkg/translate/ai.go#L663-L694)) có bug: nhánh indexed (`tryDecodeIndexedTranslations`) chỉ check `missing == 0`, KHÔNG so `len(translations)` với `expected`. Nếu LLM trả 6 items index 0-5 cho `expected=3`, parser return thành công với 6 items → caller `TranslateSRT` làm `cues[b.start+j].Text = t` với j=0..5, **ghi đè silent 3 cue của batch kế tiếp**.

Bug này tồn tại trước Phase 2 nhưng là foundation cho anti-leak của Phase 2 (chặn LLM dịch luôn context section). Phải fix trước mới overlap work được.

## Requirements

### Functional
- [ ] Thêm khả năng truyền context cues (prior + following) vào translator khi gọi `Translate`.
- [ ] Prompt 3 provider chia 3 section rõ ràng: PRIOR CONTEXT (không dịch), TRANSLATE THESE (dịch), FOLLOWING CONTEXT (không dịch).
- [ ] Response expected count = số cue cần dịch (không tính context).
- [ ] Parser reject response trả sai count hoặc index nằm ngoài range dịch.
- [ ] Overlap size configurable: default 3 prior + 3 following, cap ở `len(prevBatch)` và `len(nextBatch)`.
- [ ] Nếu batch đầu tiên → không có prior context. Nếu batch cuối → không có following context. Render prompt có điều kiện (bỏ section rỗng).

### Non-Functional
- [ ] Backward-compatible: nếu không có context (overlap=0 hoặc non-AI translator), prompt render như cũ, parser hoạt động như cũ.
- [ ] DeepL/Google không bị đụng (không implement context interface).
- [ ] Không tăng số API call (context chỉ đi kèm trong cùng request).

## Design Decision: Interface Extension

Hai option, chọn Option B vì clean và backward-compatible:

**Option A — đổi signature `Translator.Translate`** (reject):
```go
Translate(ctx, texts, priorContext, followingContext, targetLang) ([]string, error)
```
→ Break DeepL/Google/test helpers.

**Option B — optional interface (adopt)**:
```go
type ContextualTranslator interface {
    Translator
    TranslateWithContext(ctx context.Context, texts []string, prior, following []string, targetLang string) ([]string, error)
}
```
→ `TranslateSRT` type-assert; nếu translator implement `ContextualTranslator` thì gọi method kia, else fall back về `Translate`. 3 AI provider implement, DeepL/Google không.

## Implementation Steps

1. [x] **Fix parser count bug (PRE-REQUISITE — làm đầu tiên, có thể ship độc lập)** ✅ shipped
    - File: `backend/pkg/translate/ai.go` → `decodeLLMTranslations` và `tryDecodeIndexedTranslations`.
    - Đổi signature `tryDecodeIndexedTranslations(jsonBlock string, expected int)` để có `expected` ở validation layer.
    - **Validation rules (TẤT CẢ phải pass):**
        1. Mỗi `t.Index` phải thỏa `0 <= t.Index < expected` → reject nếu negative hoặc ≥ expected. Bug hiện tại: `maxIndex := -1` initial + item có `index:-1` → loop không chạy → return empty "thành công".
        2. Tất cả `t.Index` phải **unique** trong input array. Bug hiện tại: `byIndex[t.Index] = t.Text` silent-overwrite duplicates, payload `[{idx:0},{idx:1},{idx:1},{idx:2}]` có `len(byIndex)=3==expected` → PASS với dữ liệu bị co lại.
        3. `len(translations) == expected` (sau khi rebuild) — đã bàn ở review trước.
    - Implementation guide: trước khi build `byIndex`, validate toàn bộ array trong 1 pass:
      ```go
      seen := make(map[int]bool, expected)
      for _, t := range indexed {
          if t.Index < 0 || t.Index >= expected {
              return nil, nil, fmt.Errorf("index %d out of range [0, %d)", t.Index, expected)
          }
          if seen[t.Index] {
              return nil, nil, fmt.Errorf("duplicate index %d", t.Index)
          }
          seen[t.Index] = true
      }
      ```
      Sau đó mới reconstruct.
    - Trong `decodeLLMTranslations`, nhánh indexed success: thêm check `if len(translations) != expected: return nil, err`. Tuyệt đối **KHÔNG fallthrough** sang partial-missing flow khi validation fail (tránh context-leak giả dạng missing).
    - Fallback non-indexed path đã có `finalizeTranslations` check count — giữ nguyên.
    - Test cases:
        - `TestParser_ValidCount`: Indexed 3 items (0,1,2), expected=3 → pass.
        - `TestParser_RejectOverflow`: 6 items (0-5), expected=3 → reject, message "got 6, expected 3" hoặc "index 3 out of range".
        - `TestParser_RejectIndexOutOfRange`: indexes (0,1,5), expected=3 → reject (5 ≥ 3).
        - `TestParser_RejectNegativeIndex`: indexes (-1,0,1), expected=3 → reject.
        - `TestParser_RejectDuplicateIndex`: indexes (0,1,1,2), expected=3 → reject ("duplicate index 1"). KHÔNG được co lại thành 3 item giả.
        - `TestParser_PartialMissing_StillWorks`: indexes (0,2), expected=3 → partial-missing flow cũ (missing=[1]) — giữ behavior legitimate retry.
    - **Acceptance:** 6 test trên pass; existing tests trong `ai_test.go` + `translate_test.go` không regression.

2. [ ] **Định nghĩa `ContextualTranslator` interface**
    - File: `backend/pkg/translate/translate.go`.
    - Add method `TranslateWithContext(ctx, texts, prior, following []string, targetLang) ([]string, error)`.

3. [ ] **Update `TranslateSRT` — overlap CHỈ ở forced boundaries VÀ chỉ khi timing valid**
    - Consume `ChunkResult{Batches, TimingValid}` từ chunker Phase 1.
    - **Early-out:** nếu `result.TimingValid == false` → KHÔNG apply overlap cho bất kỳ batch nào (prior=following=nil ở tất cả). Lý do: chunker đã fallback fixed-size vì timing hỏng, các flag `LeftForced/RightForced` không còn đáng tin → giữ behavior gần với legacy (no overlap) thay vì áp overlap khắp nơi.
    - Khi `TimingValid == true`, cho mỗi batch `b` tại index `k`:
        - `priorTexts = nil` nếu `b.LeftForced == false` (natural cut, LLM không cần context — tiết kiệm token).
        - `priorTexts = textsOf(cues[max(0, b.Start-overlap) : b.Start])` nếu `b.LeftForced == true`.
        - Tương tự `followingTexts` dựa trên `b.RightForced`.
    - `overlap` constant = 3 (sau data analysis; có thể tune sau).
    - Type-assert translator:
        - `ContextualTranslator` + có context (prior hoặc following non-nil) → gọi `TranslateWithContext`.
        - `ContextualTranslator` + cả hai nil → gọi `Translate` (path cũ, không overhead).
        - Không phải `ContextualTranslator` (DeepL/Google) → gọi `Translate` luôn.
    - **Kết quả 3 case:**
      - `TimingValid=true`, natural boundary → zero overhead.
      - `TimingValid=true`, forced boundary → full overlap (3+3).
      - `TimingValid=false` (malformed SRT) → zero overhead toàn bộ, tương đương legacy behavior.

4. [ ] **Update prompt builder**
    - File: `ai.go`. Đổi `buildLLMUserPrompt` để nhận thêm `prior, following []string`.
    - Section layout:
      ```
      [media context block nếu có]

      --- PRIOR CONTEXT (do NOT translate, for flow only) ---
      [prior-3] ...
      [prior-2] ...
      [prior-1] ...

      --- TRANSLATE THESE (return JSON for items [0..N-1]) ---
      Translate these subtitle cues to target language: {lang}.
      IMPORTANT: Return EXACTLY {N} items in same order. Each item indexed 0..{N-1}.
      Return shape: {"translations":[{"index":0,"text":"..."},...]}
      [0] ...
      [1] ...
      ...

      --- FOLLOWING CONTEXT (do NOT translate, for flow only) ---
      [next-1] ...
      [next-2] ...
      [next-3] ...
      ```
    - Nếu prior rỗng, bỏ section. Tương tự following.
    - Cập nhật `defaultAIModelPrompt` (system prompt): thêm rule "Do NOT translate items in PRIOR/FOLLOWING CONTEXT sections. Only translate items under 'TRANSLATE THESE'. Return exactly the requested number of items, indexed 0..N-1."

5. [ ] **Implement `TranslateWithContext` cho 3 provider**
    - `openAICompatibleTranslator.TranslateWithContext` — call `buildLLMUserPrompt(texts, targetLang, ctx, prior, following)`, rest giống `Translate`.
    - Tương tự cho `geminiTranslator` và `anthropicTranslator`.
    - Refactor: extract common path vào helper để tránh duplicate (ví dụ `t.translateCore(ctx, texts, prior, following, targetLang)` return `([]string, error)`).

6. [ ] **Giữ `Translate` như wrapper**
    - `func (t *openAICompatibleTranslator) Translate(...) (...) { return t.TranslateWithContext(ctx, texts, nil, nil, targetLang) }`. Tránh duplicate code.
    - Tương tự 2 provider còn lại.

7. [ ] **Update retry path — preserve forced-boundary info VÀ timingValid**
    - `translateBatchWithRetry` signature (sau Phase 1 đã có `cues`): cần thêm 3 param `leftForced, rightForced bool, timingValid bool` — boundary metadata của batch gốc + flag toàn cục.
    - **Global kill-switch:** nếu `timingValid == false` → retry tuyệt đối không apply overlap (prior=following=nil ở mọi sub-call), bất kể `leftForced/rightForced`. Đây là mirror của early-out ở `TranslateSRT` top-level (Step 3) — giữ invariant "timing không hợp lệ → không overlap" ở mọi layer có thể gọi `TranslateWithContext`.
    - **Downsize retry** (batch contiguous, `timingValid==true`): khi chia batch gốc thành n sub-batches fixed-size, sub-batches INTERIOR (không phải đầu/cuối của batch gốc) KHÔNG có prior/following (vì chúng được quy vào "sub-batch forced at max" — nhưng đây là trong retry, không phải boundary gốc). Chỉ sub-batch đầu tiên kế thừa `leftForced` của batch gốc; sub-batch cuối cùng kế thừa `rightForced`. Interior boundaries của retry = forced trong scope retry nhưng KHÔNG có cue overlap (vì cue trước/sau là cue cùng batch gốc — redundant).
    - **Partial retry** (non-contiguous missing): Phase 1 đã quyết định disable gap-aware → cũng **disable overlap** cho path này. Lý do: cue thiếu rời rạc, "cue kế trước/sau" trong missing slice không có ý nghĩa context. Pass `prior=nil, following=nil, timingValid=false` cho retry của `missingTexts` (dùng `timingValid=false` làm tín hiệu "đừng overlap" thay vì phải nhớ thêm flag riêng).
    - Caller `TranslateSRT` truyền `result.TimingValid` từ chunker Phase 1 xuống mỗi call `translateBatchWithRetry`.

8. [ ] **Unit tests — anti-leak assertable**
    - Test prompt rendering (`ai_test.go`):
        - `TestPrompt_FirstBatch_NoPriorSection` — batch đầu (prior=nil) → output KHÔNG chứa `PRIOR CONTEXT`.
        - `TestPrompt_LastBatch_NoFollowingSection` — tương tự.
        - `TestPrompt_MiddleBatch_BothSections` — prior=3, following=3 → có cả 2 section, đếm số dòng `[prior-N]` và `[next-N]` đúng.
    - Test parser (mở rộng từ Step 1):
        - `TestParser_AntiLeak_ExtraItems` — LLM trả 8 items cho expected=5 → reject, error message chứa "got 8, expected 5".
        - `TestParser_AntiLeak_IndexOutOfRange` — indexes [0,1,2,3,4,7] expected=5 → reject (maxIndex=7 ≥ expected).
        - `TestParser_ValidCount` — LLM trả đúng 5/5 → pass.
    - Mock HTTP test (`openai_test.go` nếu chưa có):
        - `TestOpenAI_ContextualRequest_SectionsPresent` — feed prior+following → intercept HTTP body → assert có `"PRIOR CONTEXT"` + `"FOLLOWING CONTEXT"`.
        - `TestOpenAI_NonContextualRequest_SectionsAbsent` — call `Translate` (không context) → assert body KHÔNG chứa `"PRIOR CONTEXT"`.

9. [ ] **Integration smoke — assertable metric**
    - `TestTranslateSRT_OverlapOnlyForcedBoundaries` với mock `ContextualTranslator` trả text echo + log lại mỗi call `(prior, following)`:
        - Feed `testdata/movie_sparse.srt` (0 forced cuts expected từ Phase 1) → assert mọi call có `prior==nil && following==nil`.
        - Feed `testdata/sitcom_dense.srt` (≥1 forced cut expected) → assert ít nhất 1 call có `prior!=nil || following!=nil`, và số call có overlap == số forced boundaries từ Phase 1 chunker.
    - `TestTranslateSRT_MalformedTiming_NoOverlap`: tạo SRT synthetic có 1 cue timing hỏng → assert `result.TimingValid==false` và mọi call translator đều có `prior==nil && following==nil` (kể cả khi fixed-size có nhiều boundary interior).
    - KHÔNG manual inspect — assert bằng counter.

## Files to Create/Modify

- `backend/pkg/translate/translate.go` — thêm `ContextualTranslator` interface, update `TranslateSRT` + retry path.
- `backend/pkg/translate/ai.go` — update `buildLLMUserPrompt`, `defaultAIModelPrompt`, thêm `TranslateWithContext` cho 3 provider, refactor chung vào helper.
- `backend/pkg/translate/ai_test.go` — prompt rendering tests, parser rejection tests.
- `backend/pkg/translate/translate_test.go` — integration test với `ContextualTranslator` mock.

## Test Criteria
- [ ] `cd backend && go test ./pkg/translate/...` pass (go.mod ở `backend/`).
- [ ] **Parser fix (Step 1) ship được độc lập** — `TestParser_RejectOverflow` + existing tests pass ngay cả khi Step 2-9 chưa merge.
- [ ] Prompt rendered correctly cho cả 3 case (first batch / middle / last batch) — assert bằng substring match trên HTTP body.
- [ ] Parser anti-leak: 2 test `TestParser_AntiLeak_ExtraItems` và `TestParser_AntiLeak_IndexOutOfRange` pass.
- [ ] **Overlap chỉ ở forced boundaries** — `TestTranslateSRT_OverlapOnlyForcedBoundaries` verify bằng counter từ mock translator, không manual inspect.
- [ ] DeepL/Google path không regression (chạy `Translate` như cũ, không nhận context).
- [ ] Tokens count (manual verify 1 lần): batch 50 cue + 3 overlap forced → overhead trong khoảng 8-15% input tokens.

## Notes
- Các AI model có thể "ngoan cố" dịch luôn context section dù prompt cấm. Anti-leak validation ở parser là bắt buộc, không optional.
- Nếu Gemini với `responseMimeType: "application/json"` trả invalid JSON vì confused với 3 section → fallback: wrap context trong code fence ` ```context ... ``` ` để làm tín hiệu "ignore this block".
- Overlap=3 là default từ data. Có thể cân nhắc hạ xuống 2 nếu token cost thành vấn đề. Không nên tăng >5 vì marginal gain thấp.
- Sau Phase 2, dòng prompt "Count the input items first, then count your output items" vẫn giữ — nó redundant với indexed format nhưng là safety net.

---
Previous Phase: [phase-01-tiered-gap-splitting.md](phase-01-tiered-gap-splitting.md)
