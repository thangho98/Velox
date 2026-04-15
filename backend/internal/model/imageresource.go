package model

import "fmt"

type ImageResource struct {
	URL      string            `json:"url"`
	Srcset   map[string]string `json:"srcset"`
	Type     string            `json:"type"`   // poster|backdrop|still|logo
	Aspect   string            `json:"aspect"` // "2:3" | "16:9" | "variable"
	Width    int               `json:"width,omitempty"`
	Height   int               `json:"height,omitempty"`
	Blurhash *string           `json:"blurhash,omitempty"`
}

type ImageMetadata struct {
	Path       string
	Blurhash   string
	Width      int
	Height     int
	ComputedAt string
}

func BuildImageResource(rawPath string, kind string, meta *ImageMetadata) *ImageResource {
	if rawPath == "" {
		return nil
	}

	var aspect string
	switch kind {
	case "poster":
		aspect = "2:3"
	case "backdrop", "still":
		aspect = "16:9"
	default:
		aspect = "variable"
	}

	buckets := getTMDbBuckets(kind)

	base := ImagePath(rawPath).NormalizeAs(kind)

	srcset := make(map[string]string)
	for _, b := range buckets {
		wStr := fmt.Sprintf("%d", b)
		srcset[wStr] = fmt.Sprintf("%s?width=%d", base, b)
	}
	srcset["original"] = fmt.Sprintf("%s?width=original", base)

	url := fmt.Sprintf("%s?width=%d", base, getDefaultWidth(kind))

	res := &ImageResource{
		URL:    url,
		Srcset: srcset,
		Type:   kind,
		Aspect: aspect,
	}

	if meta != nil {
		res.Width = meta.Width
		res.Height = meta.Height
		if meta.Blurhash != "" {
			v := meta.Blurhash
			res.Blurhash = &v
		}
	}

	return res
}

var TMDbSizes = map[string][]int{
	"poster":   {92, 154, 185, 342, 500, 780},
	"backdrop": {300, 780, 1280},
	"still":    {92, 185, 300},
	"logo":     {45, 92, 154, 185, 300, 500},
}

var TMDbDefaultWidth = map[string]int{
	"poster":   500,
	"backdrop": 780,
	"still":    300,
	"logo":     500,
}

func getTMDbBuckets(kind string) []int {
	if b, ok := TMDbSizes[kind]; ok {
		return b
	}
	return []int{}
}

func getDefaultWidth(kind string) int {
	if w, ok := TMDbDefaultWidth[kind]; ok {
		return w
	}
	return 500
}
