package vectrace

import (
	"unsafe"
)

const sizeofWord = unsafe.Sizeof(Word(0))

// Compression relies on C zlib, so we disable it.

// C flips bitmaps by using negative bitmap strides, which we cannot represent in Go with slices.

func bm_flip(bm *Bitmap) {
	dy := bm.Dy
	if dy < 0 {
		dy = -dy
	}
	h := bm.H
	tmp := make([]Word, dy)
	for y := 0; y < h/2; y++ {
		y1 := int64(y) * int64(dy)
		y2 := int64(h-1-y) * int64(dy)
		// 行をまるごと入れ替える
		copy(tmp, bm.Map[y1:y1+int64(dy)])
		copy(bm.Map[y1:y1+int64(dy)], bm.Map[y2:y2+int64(dy)])
		copy(bm.Map[y2:y2+int64(dy)], tmp)
	}
}

// unused in the code

type potrace_privstate_s struct{}
