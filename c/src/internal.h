#ifndef NEXTIMAGE_INTERNAL_H
#define NEXTIMAGE_INTERNAL_H

#include "nextimage.h"
#include <stdarg.h>

// 内部用メモリ管理関数（カウンター付き）
void* nextimage_malloc(size_t size);
void* nextimage_calloc(size_t nmemb, size_t size);
void* nextimage_realloc(void* ptr, size_t size);
void nextimage_free(void* ptr);

// 内部用エラーメッセージ設定
void nextimage_set_error(const char* format, ...);

// デバッグビルド専用
#ifdef NEXTIMAGE_DEBUG
void nextimage_increment_alloc_counter(void);
void nextimage_decrement_alloc_counter(void);
#endif

#endif // NEXTIMAGE_INTERNAL_H
