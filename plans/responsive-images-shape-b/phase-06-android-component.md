# Phase 06: Android Coil + Blurhash Placeholder

Status: ⬜ Pending
Dependencies: Phase 02 (new fields in responses)

## Objective

Create a Compose `ResponsiveImage` composable that consumes `ImageResource`, picks the appropriate srcset variant based on composable size, and shows a blurhash placeholder while loading.

## Changes

### 1. Dependencies

`android/app/build.gradle.kts`:
```kotlin
dependencies {
    implementation("io.coil-kt:coil-compose:2.7.0")  // verify version already present
    implementation("com.vanniktech:blurhash:0.2.0")  // lightweight pure-Kotlin blurhash decoder
}
```

(vanniktech/blurhash is maintained + no Coil coupling. Alternative: `io.github.Snowdh:coil-blurhash`, but less maintained.)

### 2. Domain model

[domain/model/Image.kt](../../android/app/src/main/java/com/velox/app/domain/model/Image.kt) — new:

```kotlin
data class ImageResource(
    val url: String,
    val srcset: Map<String, String>,
    val type: String,        // poster | backdrop | still | logo
    val aspect: String,      // "2:3" | "16:9" | "variable"
    val width: Int?,
    val height: Int?,
    val blurhash: String?,
)
```

### 3. DTO + mapper

[data/model/ImageDto.kt](../../android/app/src/main/java/com/velox/app/data/model/ImageDto.kt):
```kotlin
@Serializable
data class ImageResourceDto(
    val url: String,
    val srcset: Map<String, String>,
    val type: String,
    val aspect: String,
    val width: Int? = null,
    val height: Int? = null,
    val blurhash: String? = null,
)

fun ImageResourceDto.toDomain() = ImageResource(...)
```

Repository wires this through `MediaRepositoryImpl` alongside existing `posterPath` string (both live during transition).

### 4. Composable

[presentation/ui/components/ResponsiveImage.kt](../../android/app/src/main/java/com/velox/app/presentation/ui/components/ResponsiveImage.kt):

```kotlin
@Composable
fun ResponsiveImage(
    data: ImageResource?,
    contentDescription: String?,
    modifier: Modifier = Modifier,
    contentScale: ContentScale = ContentScale.Crop,
) {
    if (data == null) {
        Box(modifier.background(Color.Black.copy(alpha = 0.4f)))
        return
    }

    val density = LocalDensity.current
    var containerWidthPx by remember { mutableStateOf(0) }

    Box(
        modifier = modifier
            .aspectRatio(aspectToRatio(data.aspect))
            .onSizeChanged { containerWidthPx = it.width }
            .clip(RectangleShape),
    ) {
        // Blurhash placeholder (always drawn underneath; image fades in on top)
        if (data.blurhash != null) {
            val bitmap = remember(data.blurhash) {
                BlurHashDecoder.decode(data.blurhash, 32, 32)
            }
            if (bitmap != null) {
                Image(
                    bitmap = bitmap.asImageBitmap(),
                    contentDescription = null,
                    modifier = Modifier.matchParentSize(),
                    contentScale = ContentScale.Crop,
                )
            }
        }

        val url = remember(containerWidthPx, data.srcset) {
            pickSrcsetVariant(data, containerWidthPx, density)
        }
        AsyncImage(
            model = url,
            contentDescription = contentDescription,
            modifier = Modifier.matchParentSize(),
            contentScale = contentScale,
        )
    }
}

private fun aspectToRatio(aspect: String): Float = when (aspect) {
    "2:3" -> 2f / 3f
    "16:9" -> 16f / 9f
    else -> 1f
}

// Pick the smallest srcset variant >= container width (in px).
private fun pickSrcsetVariant(data: ImageResource, containerPx: Int, density: Density): String {
    if (containerPx == 0) return data.url
    val candidates = data.srcset.entries
        .mapNotNull { (k, url) -> k.toIntOrNull()?.let { it to url } }
        .sortedBy { it.first }
    return candidates.firstOrNull { it.first >= containerPx }?.second
        ?: candidates.lastOrNull()?.second
        ?: data.url
}
```

### 5. Migrate call sites

Replace existing `AsyncImage` + `ImageUrlResolver.resolve(media.posterPath)` patterns in:
- [ui/screens/home/](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/home/)
- [ui/screens/detail/MediaDetailScreen.kt](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/detail/MediaDetailScreen.kt), [SeriesDetailScreen.kt](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/detail/SeriesDetailScreen.kt)
- [ui/screens/browse/](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/browse/), [movies/](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/movies/), [series/](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/series/), [search/](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/search/), [favorites/](../../android/app/src/main/java/com/velox/app/presentation/ui/screens/favorites/)
- [ui/components/](../../android/app/src/main/java/com/velox/app/presentation/ui/components/) — cards

Usage:
```kotlin
ResponsiveImage(
    data = media.poster,
    contentDescription = media.title,
    modifier = Modifier.fillMaxWidth(),
)
```

### 6. PlaybackManager metadata

[PlaybackManager buildPlaybackMetadata](../../android/app/src/main/java/com/velox/app/presentation/viewmodel/PlayerViewModel.kt) currently uses `mediaDetail.backdropPath ?: posterPath` as URL. Switch to `mediaDetail.backdrop?.url ?: mediaDetail.poster?.url` when Phase 02 response wires up.

## Tests

- Unit test `pickSrcsetVariant` with various container widths.
- Preview: `@Preview` for `ResponsiveImage` with fake blurhash + fake URL.

## Acceptance

- `./gradlew :app:compileDebugKotlin` green.
- `./gradlew :app:testDebugUnitTest` green.
- Manual: launch app → home screen → cards show blurhash flash → fade to poster. No CLS / reflow.
- Notification artwork still works (picks best variant for small icon size).

## Notes

- Coil caches images by URL — different srcset variants are separate cache entries. Disk cache grows proportionally. Acceptable.
- `onSizeChanged` fires after first layout → first frame uses `data.url` (default size) as fallback. Good-enough; subsequent recomposes use correct variant.
- Blurhash decode is ~5ms for 32x32 bitmap. Fast enough for every card.

---
Next: [Phase 07 — Drop legacy string fields + verify](phase-07-drop-legacy-verify.md)
